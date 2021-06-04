package replica

import (
	"mayastor-e2e/common/custom_resources"
	v1alpha1Api "mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	"testing"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	pvcName      = "replica-test-pvc"
	storageClass = "replica-test-nvmf"
)

const fioPodName = "fio"

func addUnpublishedReplicaTest() {
	err := k8stest.MkStorageClass(storageClass, 1, common.ShareProtoNvmf, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", storageClass)

	// Create a PVC
	k8stest.MkPVC(common.DefaultVolumeSizeMb, pvcName, storageClass, common.VolFileSystem, common.NSDefault)
	pvc, err := k8stest.GetPVC(pvcName, common.NSDefault)
	Expect(err).To(BeNil())
	Expect(pvc).ToNot(BeNil())

	timeout := "90s"
	pollPeriod := "1s"

	// Add another child before publishing the volume.
	uuid := string(pvc.ObjectMeta.UID)
	err = custom_resources.UpdateVolumeReplicaCount(uuid, 2)
	Expect(err).ToNot(HaveOccurred(), "Update number of replicas")
	replicas, err := custom_resources.GetVolumeReplicas(uuid)
	Expect(err).To(BeNil())
	Expect(len(replicas)).Should(Equal(int64(2)))

	// Use the PVC and wait for the volume to be published
	_, err = k8stest.CreateFioPod(fioPodName, pvcName, common.VolFileSystem, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Create fio pod")
	Eventually(func() bool { return custom_resources.IsVolumePublished(uuid) }, timeout, pollPeriod).Should(Equal(true))

	getChildrenFunc := func(uuid string) []v1alpha1Api.NexusChild {
		children, err := custom_resources.GetVolumeNexusChildren(uuid)
		if err != nil {
			panic("Failed to get children")
		}
		Expect(len(children)).Should(Equal(2))
		return children
	}

	// Check the added child and nexus are both degraded.
	Eventually(func() string { return getChildrenFunc(uuid)[1].State }, timeout, pollPeriod).Should(BeEquivalentTo("CHILD_DEGRADED"))
	Eventually(func() (string, error) { return custom_resources.GetVolumeNexusState(uuid) }, timeout, pollPeriod).Should(BeEquivalentTo("NEXUS_DEGRADED"))

	// Check everything eventually goes healthy following a rebuild.
	Eventually(func() string { return getChildrenFunc(uuid)[0].State }, timeout, pollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
	Eventually(func() string { return getChildrenFunc(uuid)[1].State }, timeout, pollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
	Eventually(func() (string, error) { return custom_resources.GetVolumeNexusState(uuid) }, timeout, pollPeriod).Should(BeEquivalentTo("NEXUS_ONLINE"))

	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Delete fio test pod")
	k8stest.RmPVC(pvcName, storageClass, common.NSDefault)

	err = k8stest.RmStorageClass(storageClass)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", storageClass)
}

func TestReplica(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Replica Test Suite", "replica")
}

var _ = Describe("Mayastor replica tests", func() {

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

	It("should test the addition of a replica to an unpublished volume", func() {
		addUnpublishedReplicaTest()
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
