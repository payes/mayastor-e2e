package fsx_ext4_stress

import (
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDataPlaneCorrectness(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-1783", "MQ-1783")
}

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})

var _ = Describe("Data plane correctness, ext4 stress test:", func() {

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
	It("should verify data plane correctness, ext4 stress using fsx", func() {
		c := generateFsxExt4StressConfig("fsx-ext4-stress")
		c.fsxExt4StressTest()
	})

})

func (c *fsxExt4StressConfig) fsxExt4StressTest() {
	c.createSC()
	c.createPVC()
	c.createFsx()
	c.verifyVolumeStateOverGrpcAndCrd()
	c.verifyUninterruptedIO()
	c.getNexusDetail()
	c.faultNexusChild()
	c.verifyFaultedReplica()
	c.patchMsvReplica()
	c.verifyUpdatedReplica()
	c.verifyMsvStatus()
	c.verifyUninterruptedIO()
	c.verifyVolumeStateOverGrpcAndCrd()
	c.waitForFsxPodCompletion()
	c.deleteFsx()
	c.deletePVC()
	c.deleteSC()
}
