package primitive_volumes

import (
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestConcurrentPvc(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, " Primitive large number of volume operations", "primitive_volumes")
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

var _ = Describe("Primitive large number of volume operations", func() {

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
	It("should verify serial volume creation", func() {
		volCount := msnList()
		c := generatePvcConcurrentConfig("serial-pvc-create", volCount)
		c.pvcSerialTest()
	})
	It("should verify concurrent volume creation", func() {
		volCount := msnList()
		c := generatePvcConcurrentConfig("concurrent-pvc-create", volCount)
		c.pvcConcurrentTest()
	})
})
