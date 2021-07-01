package primitive_volumes

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"sync"

	. "github.com/onsi/gomega"
)

var defTimeoutSecs = "90s"

// createStorageClass will create storageclass
func (c *pvcConcurrentConfig) createStorageClass() {
	err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithFileSystemType(c.fsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.scName)
}

// deleteSC will delete storageclass
func (c *pvcConcurrentConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createSerialPVC will create pvc in serial
func (c *pvcConcurrentConfig) createSerialPVC(pvcName string) {
	// Create the volumes
	k8stest.MkPVC(c.pvcSize, pvcName, c.scName, common.VolFileSystem, common.NSDefault)
}

// deleteSerialPVC will delete pvc in serial
func (c *pvcConcurrentConfig) deleteSerialPVC(pvcName string) {
	// Create the volumes
	k8stest.RmPVC(pvcName, c.scName, common.NSDefault)
}

func (c *pvcConcurrentConfig) verifyVolumesCreation() {
	for ix := 0; ix < len(c.pvcNames); ix++ {
		// Confirm that the PVC has been created
		Expect(c.createErrs[ix]).To(BeNil(), "failed to create PVC %s", c.pvcNames[ix])
	}
}

// verify deletion of pvc and corresponding msv
func (c *pvcConcurrentConfig) verifyVolumesDeletion() {
	for ix := 0; ix < len(c.pvcNames); ix++ {
		// Confirm that the PVC has been created
		Expect(c.deleteErrs[ix]).To(BeNil(), "failed to delete PVC %s", c.pvcNames[ix])
		// Wait for the PVC to be deleted.
		Eventually(func() bool {
			return k8stest.IsPVCDeleted(c.pvcNames[ix], common.NSDefault)
		},
			defTimeoutSecs, // timeout
			"1s",           // polling interval
		).Should(Equal(true))
	}
}

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
