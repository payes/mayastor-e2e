package primitive_max_volumes_in_pool

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/common/k8stest"
	"sync"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// createSC will create storageclass
func (c *primitiveMaxVolConfig) createSC() {
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
func (c *primitiveMaxVolConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createPVCs will create pvc
func (c *primitiveMaxVolConfig) createPVCs() *primitiveMaxVolConfig {
	// Create the volumes
	for i := 0; i < len(c.pvcNames); i++ {
		pvc, err := k8stest.CreatePVC(&c.optsList[i], common.NSDefault)
		c.createErrs[i] = err
		c.uuid[i] = string(pvc.UID)
	}
	return c
}

// createVolumes will create volumes
func (c *primitiveMaxVolConfig) createVolumes() *primitiveMaxVolConfig {
	// Create the volumes
	var wg sync.WaitGroup
	wg.Add(len(c.pvcNames))
	for i := 0; i < len(c.pvcNames); i++ {
		go k8stest.CreatePvc(&c.optsList[i], &c.createErrs[i], &c.uuid[i], &wg)
	}
	wg.Wait()

	logf.Log.Info("Finished calling the create methods for all PVC candidates.")

	return c
}

// removeVolumes will remove volumes
func (c *primitiveMaxVolConfig) removeVolumes() *primitiveMaxVolConfig {
	// Create the volumes
	var wg sync.WaitGroup
	wg.Add(len(c.pvcNames))
	for i := 0; i < len(c.pvcNames); i++ {
		go k8stest.DeletePvc(c.pvcNames[i], common.NSDefault, &c.createErrs[i], &wg)
	}
	wg.Wait()
	logf.Log.Info("Finished calling the delete methods for all PVC candidates.")

	return c
}

// verify msp used size
func (c *primitiveMaxVolConfig) verifyMspUsedSize(size int64) {
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

// deletePVC will delete all pvc
func (c *primitiveMaxVolConfig) deletePVC() {
	for _, pvc := range c.pvcNames {
		k8stest.RmPVC(pvc, c.scName, common.NSDefault)
	}
}

// Check that all volumes have been created successfully,
// that none of them are in the pending state,
// that all of them have the right size
func (c *primitiveMaxVolConfig) verifyVolumesCreation() {
	for ix := 0; ix < len(c.pvcNames); ix++ {
		// Confirm that the PVC has been created
		Expect(c.createErrs[ix]).To(BeNil(), "failed to create PVC %s", c.pvcNames[ix])

		namespace := common.NSDefault
		volName := c.pvcNames[ix]
		pvc, getPvcErr := k8stest.GetPVC(volName, namespace)
		Expect(getPvcErr).To(BeNil(), "Failed to get PVC %s", volName)
		Expect(pvc).ToNot(BeNil())

		// Wait for the PVC to be bound.
		Eventually(func() coreV1.PersistentVolumeClaimPhase {
			return k8stest.GetPvcStatusPhase(volName, namespace)
		},
			defTimeoutSecs, // timeout
			"1s",           // polling interval
		).Should(Equal(coreV1.ClaimBound))

		// Refresh the PVC contents, so that we can get the PV name.
		pvc, getPvcErr = k8stest.GetPVC(volName, namespace)
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
func (c *primitiveMaxVolConfig) verifyVolumesDeletion() {
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
