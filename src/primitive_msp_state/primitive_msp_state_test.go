package primitive_msp_state

import (
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMspState(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, " Test mayastor pool state", "primitive_msp_state")
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

var _ = Describe("Mayastor pool state tests", func() {

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
	It("should verify mayastor pool state", func() {
		verifyMspCrdAndGrpcState()
	})
	It("should verify correctness of all msp CRD fields for all operations", func() {
		c := generateMspStateConfig("primitive-msp-state", 1)
		c.mspCrdPresenceTest()
	})
})

func (c *mspStateConfig) mspCrdPresenceTest() {
	verifyMspCrdAndGrpcState()
	c.createSC()
	c.createPVC()
	c.createFioPods()
	c.getMsvDetails()
	c.verifyMspUsedSize()
	k8stest.WaitPodComplete(c.fioPodName, 5, int(c.timeout))
	c.deleteFioPods()
	c.deletePVC()
	c.deleteSC()
}
