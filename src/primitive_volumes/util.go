package primitive_volumes

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/common/k8stest"
	"sync"
	"time"

	coreV1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
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
	namespace := common.NSDefault
	for ix := 0; ix < len(c.pvcNames); ix++ {
		// Confirm that the PVC has been created
		Expect(c.createErrs[ix]).To(BeNil(), "failed to create PVC %s", c.pvcNames[ix])
		volName := c.pvcNames[ix]
		pvc, getPvcErr := k8stest.GetPVC(volName, namespace)
		Expect(getPvcErr).To(BeNil(), "Failed to get PVC %s", volName)
		Expect(pvc).ToNot(BeNil())
	}

	elapsedTime := 0
	const sleepTime = 10
	allBound := false
	volBindMap := make(map[string]bool)
	for ; !allBound && elapsedTime < 300; elapsedTime += sleepTime {
		allBound = true
		for ix := 0; ix < len(c.pvcNames); ix++ {
			volName := c.pvcNames[ix]
			bound := coreV1.ClaimBound == k8stest.GetPvcStatusPhase(volName, namespace)
			allBound = allBound && bound
			volBindMap[volName] = bound
		}
		time.Sleep(sleepTime * time.Second)
	}
	logf.Log.Info("PVCs bind status", "elapsed (secs)", elapsedTime, "allBound", allBound)
	if !allBound {
		for volName, bound := range volBindMap {
			if !bound {
				logf.Log.Info("", "unbound vol", volName)
			}
		}
	}
	Expect(allBound).To(BeTrue(), "all pvcs were not bound")

	for ix := 0; ix < len(c.pvcNames); ix++ {
		volName := c.pvcNames[ix]
		// Refresh the PVC contents, so that we can get the PV name.
		pvc, getPvcErr := k8stest.GetPVC(volName, namespace)
		Expect(getPvcErr).To(BeNil())
		Expect(pvc).ToNot(BeNil())
		logf.Log.Info("Created", "volume", pvc.Spec.VolumeName, "uuid", pvc.ObjectMeta.UID)

		// Wait for the PV to be provisioned
		Eventually(func() *coreV1.PersistentVolume {
			pv, getPvErr := k8stest.GetPV(pvc.Spec.VolumeName)
			if getPvErr != nil {
				return nil
			}
			return pv

		},
			defTimeoutSecs, // timeout
			"1s",           // polling interval
		).Should(Not(BeNil()))

		// Wait for the PV to be bound.
		Eventually(func() coreV1.PersistentVolumePhase {
			return k8stest.GetPvStatusPhase(pvc.Spec.VolumeName)
		},
			defTimeoutSecs, // timeout
			"1s",           // polling interval
		).Should(Equal(coreV1.VolumeBound))

		Eventually(func() *v1alpha1.MayastorVolume {
			return k8stest.GetMSV(string(pvc.ObjectMeta.UID))
		},
			defTimeoutSecs,
			"1s",
		).Should(Not(BeNil()))

		logf.Log.Info("Created", "volume", volName, "uuid", pvc.ObjectMeta.UID, "storageClass", c.scName)
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
		// Wait for the MSV to be deleted.
		Eventually(func() bool {
			return custom_resources.IsMsVolDeleted(c.uuid[ix])
		},
			defTimeoutSecs, // timeout
			"1s",           // polling interval
		).Should(Equal(true))
	}
}

func (c *pvcConcurrentConfig) pvcConcurrentTest() {
	c.createStorageClass()
	var wg sync.WaitGroup
	for ix := 0; ix < c.iterations; ix++ {
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
	}
	c.deleteSC()
	c.waitForMspUsedSize(0)
}

func (c *pvcConcurrentConfig) pvcSerialTest() {
	c.createStorageClass()
	for ix := 0; ix < c.iterations; ix++ {
		for _, pvcName := range c.pvcNames {
			c.createSerialPVC(pvcName)
		}
		for _, pvcName := range c.pvcNames {
			c.deleteSerialPVC(pvcName)
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
func (c *pvcConcurrentConfig) waitForMspUsedSize(size int64) {
	// List Pools by CRDs
	crdPools, err := custom_resources.ListMsPools()
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
		pool, err := custom_resources.GetMsPool(poolName)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to get mayastor pool %s %v", poolName, err))
		if pool.Status.Used == usedSize {
			return nil
		}
	}
	return errors.Errorf("pool %s used size did not reconcile in %d seconds", poolName, timeoutSec)
}
