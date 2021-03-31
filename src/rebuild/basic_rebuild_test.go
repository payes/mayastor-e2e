package basic_rebuild_test

import (
	"mayastor-e2e/common/k8stest"
	"testing"

	"mayastor-e2e/common"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	podName      = "rebuild-test-fio"
	pvcName      = "rebuild-test-pvc"
	storageClass = "rebuild-test-nvmf"
)

func basicRebuildTest() {
	err := k8stest.MkStorageClass(storageClass, 1, common.ShareProtoNvmf, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", storageClass)

	// Create a PVC
	k8stest.MkPVC(common.DefaultVolumeSizeMb, pvcName, storageClass, common.VolFileSystem, common.NSDefault)
	pvc, err := k8stest.GetPVC(pvcName, common.NSDefault)
	Expect(err).To(BeNil())
	Expect(pvc).ToNot(BeNil())

	timeout := "90s"
	pollPeriod := "1s"

	// Create an application pod and wait for the PVC to be bound to it.
	_, err = k8stest.CreateFioPod(podName, pvcName, common.VolFileSystem, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Failed to create rebuild test fio pod")
	Eventually(func() bool { return k8stest.IsPvcBound(pvcName, common.NSDefault) }, timeout, pollPeriod).Should(Equal(true))

	uuid := string(pvc.ObjectMeta.UID)
	repl, err := k8stest.GetNumReplicas(uuid)
	Expect(err).To(BeNil())
	Expect(repl).Should(Equal(int64(1)))

	// Wait for volume to be published before adding a child.
	// This ensures that a nexus exists when the child is added.
	Eventually(func() bool { return k8stest.IsVolumePublished(uuid) }, timeout, pollPeriod).Should(Equal(true))

	// Add another child which should kick off a rebuild.
	err = k8stest.UpdateNumReplicas(uuid, 2)
	Expect(err).ToNot(HaveOccurred(), "Update the number of replicas")
	repl, err = k8stest.GetNumReplicas(uuid)
	Expect(err).To(BeNil())
	Expect(repl).Should(Equal(int64(2)))

	// Wait for the added child to show up.
	Eventually(func() int { return k8stest.GetNumChildren(uuid) }, timeout, pollPeriod).Should(BeEquivalentTo(2))

	getChildrenFunc := func(uuid string) []k8stest.NexusChild {
		children, err := k8stest.GetChildren(uuid)
		if err != nil {
			panic("Failed to get children")
		}
		Expect(len(children)).Should(Equal(2))
		return children
	}

	// Check the added child and nexus are both degraded.
	Eventually(func() string { return getChildrenFunc(uuid)[1].State }, timeout, pollPeriod).Should(BeEquivalentTo("CHILD_DEGRADED"))
	Eventually(func() (string, error) { return k8stest.GetNexusState(uuid) }, timeout, pollPeriod).Should(BeEquivalentTo("NEXUS_DEGRADED"))

	// Check everything eventually goes healthy following a rebuild.
	Eventually(func() string { return getChildrenFunc(uuid)[0].State }, timeout, pollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
	Eventually(func() string { return getChildrenFunc(uuid)[1].State }, timeout, pollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
	Eventually(func() (string, error) { return k8stest.GetNexusState(uuid) }, timeout, pollPeriod).Should(BeEquivalentTo("NEXUS_ONLINE"))
	err = k8stest.DeletePod(podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Deleting rebuild test fio pod")
	k8stest.RmPVC(pvcName, storageClass, common.NSDefault)
	err = k8stest.RmStorageClass(storageClass)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", storageClass)
}

func TestRebuild(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Rebuild Test Suite", "rebuild")
}

var _ = Describe("Mayastor rebuild test", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should run a rebuild job to completion", func() {
		basicRebuildTest()
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
