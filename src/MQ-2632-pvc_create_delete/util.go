package pvc_create_delete

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// createStorageClass will create storageclass
func (c *pvcCreateDeleteConfig) createStorageClass() {
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
func (c *pvcCreateDeleteConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createPVC will create pvc in serial
func (c *pvcCreateDeleteConfig) createPVC(pvcName string) {
	logf.Log.Info("Creating", "volume", pvcName, "storageClass", c.scName, "volume type", common.VolFileSystem)
	volSizeMbStr := fmt.Sprintf("%dMi", c.pvcSizeMB)
	var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem
	// PVC create options
	createOpts := &coreV1.PersistentVolumeClaim{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      pvcName,
			Namespace: common.NSDefault,
		},
		Spec: coreV1.PersistentVolumeClaimSpec{
			StorageClassName: &c.scName,
			AccessModes:      []coreV1.PersistentVolumeAccessMode{coreV1.ReadWriteOnce},
			Resources: coreV1.ResourceRequirements{
				Requests: coreV1.ResourceList{
					coreV1.ResourceStorage: resource.MustParse(volSizeMbStr),
				},
			},
			VolumeMode: &fileSystemVolumeMode,
		},
	}
	// Create the volumes
	pvc, err := k8stest.CreatePVC(createOpts, common.NSDefault)
	Expect(err).To(BeNil(), "Failed to create pvc, error %v", err)
	Expect(pvc).ToNot(BeNil())
}

// deleteSerialPVC will delete pvc in serial
func (c *pvcCreateDeleteConfig) deletePVC(pvcName string) {
	err := k8stest.DeletePVC(pvcName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "PVC deletion failed, pvc name: %s , Error: %v", pvcName, err)

}

func (c *pvcCreateDeleteConfig) pvcCreateDeleteTest() {
	c.createStorageClass()
	for ix := 0; ix < c.iterations; ix++ {
		for _, pvcName := range c.pvcNames {
			c.createPVC(pvcName)
		}
		//Wait for more than 5 min mins. This step makes sure that all the volume creation requests have been sent to csi controller pod
		logf.Log.Info("Added sleep for more then five min", "sleep time in minutes", c.delayTime)
		time.Sleep(time.Duration(c.delayTime) * time.Minute)
		for _, pvcName := range c.pvcNames {
			c.deletePVC(pvcName)
		}
	}
	c.waitForPvcDeletion()
	c.waitForPvDeletion()
	c.waitForMsvDeletion()
	c.waitForMspUsedSize(0)
	c.deleteSC()
}
func msnList() int {
	msnList, err := k8stest.GetMayastorNodeNames()
	Expect(err).ToNot(HaveOccurred())
	return len(msnList)
}

// verify msp used size
func (c *pvcCreateDeleteConfig) waitForMspUsedSize(size uint64) {
	// List Pools by CRDs
	crdPools, err := k8stest.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	for _, crdPool := range crdPools {
		err := checkPoolUsedSize(crdPool.Name, size)
		Expect(err).ShouldNot(HaveOccurred(), "failed to verify used size of pool %s error %v", crdPool.Name, err)
	}
}

// checkPoolUsedSize verify mayastor pool used size
func checkPoolUsedSize(poolName string, usedSize uint64) error {
	logf.Log.Info("Waiting for pool used size", "name", poolName)
	for ix := 0; ix < (timeoutSec+sleepTimeSec-1)/sleepTimeSec; ix++ {
		time.Sleep(time.Duration(sleepTimeSec) * time.Second)
		pool, err := k8stest.GetMsPool(poolName)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to get mayastor pool %s %v", poolName, err))
		if pool.Status.Used == usedSize {
			return nil
		}
	}
	return errors.Errorf("pool %s used size did not reconcile in %d seconds", poolName, timeoutSec)
}

// verify all msv to be deleted
func (c *pvcCreateDeleteConfig) waitForMsvDeletion() {
	logf.Log.Info("Waiting for mayastor volumes to be deleted", "storageClass", c.scName)
	// Wait for the msv to be deleted.
	Eventually(func() int {
		msv, err := k8stest.ListMsvs()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to list msv %v", err))
		return len(msv)
	},
		defTimeoutSecs, // timeout
		"6s",           // polling interval
	).Should(Equal(0), "all msv are not getting deleted")
}

// verify all pv to be deleted
func (c *pvcCreateDeleteConfig) waitForPvDeletion() {
	logf.Log.Info("Waiting for all PV to be deleted", "storageClass", c.scName)
	// Wait for the PV to be deleted.
	Eventually(func() bool {
		pvFound, err := k8stest.CheckForPVs()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to list pv %v", err))
		return pvFound
	},
		defTimeoutSecs, // timeout
		"6s",           // polling interval
	).Should(Equal(false), "all PV are not getting deleted")
}

// verify all pvc to be deleted
func (c *pvcCreateDeleteConfig) waitForPvcDeletion() {
	logf.Log.Info("Waiting for all PVC to be deleted", "storageClass", c.scName)
	// Wait for the PVC to be deleted.
	Eventually(func() bool {
		pvcFound, err := k8stest.CheckForPVCs()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to list pvc %v", err))
		return pvcFound
	},
		defTimeoutSecs, // timeout
		"6s",           // polling interval
	).Should(Equal(false), "all PVC are not getting deleted")
}
