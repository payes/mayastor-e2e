package pvc_create_delete

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
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

// createSerialPVC will create pvc in serial
func (c *pvcCreateDeleteConfig) createPVC(pvcName string) {
	// Create the volumes
	k8stest.MkPVC(c.pvcSize, pvcName, c.scName, common.VolFileSystem, common.NSDefault)
}

// deleteSerialPVC will delete pvc in serial
func (c *pvcCreateDeleteConfig) deletePVC(pvcName string) {
	// Create the volumes
	k8stest.RmPVC(pvcName, c.scName, common.NSDefault)
}

func (c *pvcCreateDeleteConfig) pvcCreateDeleteTest() {
	c.createStorageClass()
	for ix := 0; ix < c.iterations; ix++ {
		for _, pvcName := range c.pvcNames {
			c.createPVC(pvcName)
		}
		for _, pvcName := range c.pvcNames {
			c.deletePVC(pvcName)
		}
	}
	c.deleteSC()
	c.waitForMspUsedSize(0)
}
func msnList() int {
	msnList, err := k8stest.GetMayastorNodeNames()
	Expect(err).ToNot(HaveOccurred())
	return len(msnList)
}

// verify msp used size
func (c *pvcCreateDeleteConfig) waitForMspUsedSize(size int64) {
	// List Pools by CRDs
	crdPools, err := k8stest.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	for _, crdPool := range crdPools {
		err := checkPoolUsedSize(crdPool.Name, size)
		Expect(err).ShouldNot(HaveOccurred(), "failed to verify used size of pool %s error %v", crdPool.Name, err)
	}
}

// checkPoolUsedSize verify mayastor pool used size
func checkPoolUsedSize(poolName string, usedSize int64) error {
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
