package msv_rebuild

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestMayastorNode(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Mayastor rebuild test", "msv_rebuild")
}

func mayastorRebuildTest() {
	params := e2e_config.GetConfig().MsvRebuild
	scName := strings.ToLower(fmt.Sprintf("msv-rebuild-%d", params.Replicas))

	// Create storage class
	createStorageClass(scName, params.Replicas)
	// Create a PVC
	pvcName := strings.ToLower(fmt.Sprintf("msv-rebuild-volume-%d", params.Replicas))
	uuid := k8stest.MkPVC(common.DefaultVolumeSizeMb, pvcName, scName, common.VolFileSystem, common.NSDefault)
	log.Log.Info("Volume", "uid", uuid)

	fioPodName := "fio-" + pvcName

	// Create fio pod
	err := createFioPod(fioPodName, pvcName, params.DurationSecs, params.VolSize)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Creating fio pod %s %v", fioPodName, err))

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		params.Timeout,
		params.PollPeriod,
	).Should(Equal(true))

	// Check replicas
	replicas, err := custom_resources.GetMsVolReplicas(uuid)
	Expect(err).To(BeNil())
	Expect(len(replicas)).Should(Equal(params.Replicas))

	// Wait for volume to be published before adding a child.
	// This ensures that a nexus exists when the child is added.
	Eventually(func() bool { return custom_resources.IsMsVolPublished(uuid) }, params.Timeout, params.PollPeriod).Should(Equal(true))

	for i := 0; i < 2; i++ {
		// Add another child which should kick off a rebuild.
		err = custom_resources.UpdateMsVolReplicaCount(uuid, params.UpdatedReplica)
		Expect(err).ToNot(HaveOccurred(), "Update the number of replicas")

		// Check replica after changing replica count
		Eventually(func() bool {
			replicas, err = custom_resources.GetMsVolReplicas(uuid)
			if err != nil {
				panic("Failed to get children")
			}
			return len(replicas) == params.UpdatedReplica
		},
			params.Timeout,
			params.PollPeriod,
		).Should(Equal(true))

		// Wait for the added child to show up.
		Eventually(func() int {
			msv, err := custom_resources.GetMsVol(uuid)
			if err == nil {
				return len(msv.Status.Nexus.Children)
			}
			return 0
		}, params.Timeout, params.PollPeriod).Should(BeEquivalentTo(params.UpdatedReplica))

		// Verify children count should equal to replicas
		Eventually(func() bool {
			return getChildrenCount(uuid, params.UpdatedReplica)
		},
			params.Timeout,
			params.PollPeriod,
		).Should(Equal(true))

		// Check everything eventually goes healthy following a rebuild.
		Eventually(func() string { return getChildren(uuid)[0].State }, params.Timeout, params.PollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
		Eventually(func() string { return getChildren(uuid)[1].State }, params.Timeout, params.PollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
		Eventually(func() (string, error) { return custom_resources.GetMsVolNexusState(uuid) }, params.Timeout, params.PollPeriod).Should(BeEquivalentTo("NEXUS_ONLINE"))

		// remove one child of nexus
		err = custom_resources.UpdateMsVolReplicaCount(uuid, params.Replicas)
		Expect(err).ToNot(HaveOccurred(), "Update the number of replicas")

		// Check replicas after changing replica count
		Eventually(func() bool {
			replicas, err = custom_resources.GetMsVolReplicas(uuid)
			if err != nil {
				panic("Failed to get replicas")
			}
			return len(replicas) == params.Replicas
		},
			params.Timeout,
			params.PollPeriod,
		).Should(Equal(true))

		// Check everything remains in healthy state.
		Eventually(func() string { return getChildren(uuid)[0].State }, params.Timeout, params.PollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
		Eventually(func() (string, error) { return custom_resources.GetMsVolNexusState(uuid) }, params.Timeout, params.PollPeriod).Should(BeEquivalentTo("NEXUS_ONLINE"))
	}
	// Wait untill fio pod is in completed state
	err = k8stest.WaitPodComplete(fioPodName, params.SleepSecs, params.DurationSecs)
	Expect(err).ToNot(HaveOccurred())

	// Delete fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Deleting fio pod")

	// Delete pvc
	k8stest.RmPVC(pvcName, scName, common.NSDefault)
	// Delete storage class
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

var _ = Describe("Rebuild mayastor check", func() {
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
	It("Verify mayastor rebuild check", func() {
		mayastorRebuildTest()
	})
})
var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()
	close(done)
}, 60)
var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
