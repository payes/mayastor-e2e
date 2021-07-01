package primitive_volumes

import (
	"sync"
	"testing"

	"mayastor-e2e/common"
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
		c := generatePvcConcurrentConfig("serial-pvc-create")
		c.pvcSerialTest()
	})
	It("should verify concurrent volume creation", func() {
		c := generatePvcConcurrentConfig("concurrent-pvc-create")
		c.pvcConcurrentTest()
	})
})

func (c *pvcConcurrentConfig) pvcConcurrentTest() {
	c.createStorageClass()
	var wg sync.WaitGroup
	wg.Add(len(c.pvcNames))
	for i := 0; i < len(c.pvcNames); i++ {
		go k8stest.CreatePvc(&c.optsList[i], &c.createErrs[i], &c.uuid[i], &wg)
	}
	wg.Wait()
	c.verifyVolumesCreation()
	wg.Add(len(c.pvcNames))
	for i := 0; i < len(c.pvcNames); i++ {
		go k8stest.DeletePvc(c.pvcNames[i], common.NSDefault, &c.createErrs[i], &wg)
	}
	wg.Wait()
	c.verifyVolumesDeletion()
	c.deleteSC()
}

func (c *pvcConcurrentConfig) pvcSerialTest() {
	c.createStorageClass()

	for _, pvcName := range c.pvcNames {
		c.createSerialPVC(pvcName)
	}
	for _, pvcName := range c.pvcNames {
		c.deleteSerialPVC(pvcName)
	}
	c.deleteSC()
}
