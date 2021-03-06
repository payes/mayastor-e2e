package maximum_vols_io

import (
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMaximumVolsIO(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, " Test maximum number of  volumes", "maximum_vols_io")
}

var _ = BeforeSuite(func(done Done) {
	err := k8stest.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnv %v", err)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	err := k8stest.TeardownTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)
})

var _ = Describe("Maximum number of  volumes tests", func() {

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
	It("should verify maximum number of  volumes test with 1 replica", func() {
		c := generateMaxVolConfig("max-volume", 1)
		c.maxVolumeTest()
	})
	It("should verify maximum number of  volumes test with 2 replica", func() {
		c := generateMaxVolConfig("max-volume", 2)
		c.maxVolumeTest()
	})
	It("should verify maximum number of  volumes test with 3 replica", func() {
		c := generateMaxVolConfig("max-volume", 3)
		c.maxVolumeTest()
	})
})

func (c *maxVolConfig) maxVolumeTest() {
	c.createSC()
	c = c.createPVC()
	c.createFioPods()
	c.checkFioPodsComplete()
	c.deleteFioPods()
	c.deletePVC()
	c.deleteSC()
}
