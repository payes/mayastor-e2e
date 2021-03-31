package ms_pod_disruption

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var env DisruptionEnv

const gStorageClass = "mayastor-nvmf-pod-remove-test-sc"

func TestMayastorPodLoss(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Replica pod removal tests", "replica-pod-remove")
}

var _ = Describe("Mayastor replica pod removal test", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := k8stest.RmStorageClass(gStorageClass)
		Expect(err).ToNot(HaveOccurred())

		// Check resource leakage.
		err = k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should verify nvmf nexus behaviour when a mayastor pod is removed", func() {
		err := k8stest.MkStorageClass(gStorageClass, 2, common.ShareProtoNvmf, common.NSDefault)
		Expect(err).ToNot(HaveOccurred())
		env = Setup("loss-test-pvc-nvmf", gStorageClass, "fio-pod-remove-test")
		env.PodLossTest()
		env.Teardown() // removes fio pod and volume
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
