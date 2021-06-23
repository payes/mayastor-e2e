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
	// It("should verify maximum number of  volumes test with 2 replica", func() {
	// 	c := generateMaxVolConfig("max-volume", 2)
	// 	c.maxVolumeTest()
	// })
	// It("should verify maximum number of  volumes test with 3 replica", func() {
	// 	c := generateMaxVolConfig("max-volume", 3)
	// 	c.maxVolumeTest()
	// })
})

// func (c *maxVolConfig) maxVolumeTest() {
// 	c.createSC()
// 	c = c.createPVC()
// 	c.createFioPods()
// 	c.checkFioPodsComplete()
// 	c.deleteFioPods()
// 	c.deletePVC()
// 	c.deleteSC()
// }

func mspStateTest() {
}
