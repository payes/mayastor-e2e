package xfstests

import (
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDataPlaneCorrectness(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-1775", "MQ-1775")
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
	It("should verify data plane correctness, xfs tests using xfstests", func() {
		c := generateXFSTestsConfig("xfstest")
		c.xfstest()
	})

})

func (c *xfsTestConfig) xfstest() {
	c.createSC()
	c.createPVCs()
	c.createXFSTestPod()
	c.verifyUninterruptedIO()
	c.waitForXFSTestPodCompletion()
	c.deleteXFSTest()
	c.deletePVC()
	c.deleteSC()
}
