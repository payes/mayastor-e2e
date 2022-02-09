package MQ_2644_invalid_volume_sizes

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"mayastor-e2e/common/k8stest"
	"testing"
)

func TestInvalidVolumeSizes(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-2644", "MQ-2644")
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

var _ = Describe("Test invalid volume sizes", func() {

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

	It("should verify to not create pvc with negative size", func() {
		testNames = nil
		c := generatePvc("negative-size", 3, -1000)
		c.pvcZeroOrNegativeSizeTest()
		cleanUp()
	})

	It("should verify to not create pvc with zero size", func() {
		testNames = nil
		c := generatePvc("zero-size", 3, 0)
		c.pvcZeroOrNegativeSizeTest()
		cleanUp()
	})

	It("should verify to not create pvc bigger than pool", func() {
		testNames = nil
		c := generatePvc("bigger-than-pool", 3, 11000)
		c.pvcInvalidSizeTest()
		cleanUp()
	})

	It("should verify to not create pvc without enough space left all pools", func() {
		testNames = nil
		c := generatePvc("normal-size", 3, 8000)
		c.pvcNormalFioTest()
		c2 := generatePvc("bigger-than-remaining", 3, 8000)
		c2.pvcInvalidSizeTest()
		c.runAndDeleteFio()
		cleanUp()
	})

	It("should verify to not create pvc without enough space left in one pool", func() {
		testNames = nil
		c := generatePvc("normal-size-3-replicas", 3, 1000)
		c.pvcNormalFioTest()
		c2 := generatePvc("normal-size-1-replica", 1, 5000)
		c2.pvcNormalFioTest()
		c3 := generatePvc("not-enough-space-one-pool", 3, 5000)
		c3.pvcInvalidSizeTest()
		c.runAndDeleteFio()
		c2.runAndDeleteFio()
		cleanUp()
	})
})
