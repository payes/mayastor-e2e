package primitive_max_volumes_in_pool

import (
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPrimitiveMaximumVolInPool(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, " Test large number of volumes in pool", "primitive_max_volumes_in_pool")
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

var _ = Describe("Large number of volumes in pool tests", func() {

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
	It("should verify maximum number of  volumes in pool test", func() {
		c := generatePrimitiveMaxVolConfig("primitive-max-volume-pool", 3)
		c.primitiveMaxVolumeInPoolTest()
	})

})

func (c *primitiveMaxVolConfig) primitiveMaxVolumeInPoolTest() {
	c.createSC()
	c.createPVC()
	c.deletePVC()
	c.deleteSC()
}
