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
	It("should verify serial creation of maximum number of  volumes in pool test", func() {
		c := generatePrimitiveMaxVolConfig("primitive-max-volume-pool")
		c.serialMaxVolumeInPoolTest()
	})

	It("should verify concurrent creation of maximum number of  volumes in pool test", func() {
		c := generatePrimitiveMaxVolConfig("primitive-max-volume-pool")
		c.concurrentMaxVolumeInPoolTest()
	})

})

func (c *primitiveMaxVolConfig) serialMaxVolumeInPoolTest() {
	c.createSC()
	c.createPVCs()
	c.verifyVolumesCreation()
	c.verifyMspUsedSize(uint64(c.pvcSize))
	c.deletePVC()
	c.deleteSC()
	c.verifyMspUsedSize(0)
}

func (c *primitiveMaxVolConfig) concurrentMaxVolumeInPoolTest() {
	c.createSC()
	c.createVolumes()
	c.verifyVolumesCreation()
	c.verifyMspUsedSize(uint64(c.pvcSize))
	c.removeVolumes()
	c.verifyVolumesDeletion()
	c.deleteSC()
	c.verifyMspUsedSize(0)
}
