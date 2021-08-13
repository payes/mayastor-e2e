package primitive_fuzz_msv

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"

	"mayastor-e2e/common/k8stest"
	"sync"
	"time"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

// serialFuzzMsvTest creates/deletes sc,pvc,fio-pod in sequence
func (c *PrimitiveMsvFuzzConfig) serialFuzzMsvTest() {
	c.createSC()
	c.createPvcSerial()
	c.verifyVolumesCreation()
	c.createFioPods()
	c.deleteFioPods()
	c.deletePVC()
	c.deleteSC()
	c.waitForMspUsedSize(0)
}

// concurrentMsvFuzz creates/deletes sc,pvc,fio-pod in concurrent
func (c *PrimitiveMsvFuzzConfig) ConcurrentMsvFuzz() {
	c.createSC()
	c.createPvcParallel()
	c.verifyVolumesCreation()
	c.createFioPods()
	c.deleteFioPods()
	c.deletePvcParallel()
	c.verifyVolumesDeletion()
	c.deleteSC()
	c.waitForMspUsedSize(0)
}

// createSC will create storageclass
func (c *PrimitiveMsvFuzzConfig) createSC() {
	err := k8stest.NewScBuilder().
		WithName(c.ScName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.Protocol).
		WithReplicas(c.Replicas).
		WithFileSystemType(c.FsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.ScName)
}

// deleteSC will delete storageclass
func (c *PrimitiveMsvFuzzConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.ScName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.ScName)
}

// createPvcSerial will create pvc in serial
func (c *PrimitiveMsvFuzzConfig) createPvcSerial() *PrimitiveMsvFuzzConfig {
	// Create the volumes
	for i := 0; i < len(c.PvcNames); i++ {
		pvc, err := k8stest.CreatePVC(&c.OptsList[i], common.NSDefault)
		c.CreateErrs[i] = err
		c.Uuid[i] = string(pvc.UID)
	}
	return c
}

// verify deletion of pvc and corresponding msv
func (c *PrimitiveMsvFuzzConfig) verifyVolumesDeletion() {
	for ix := 0; ix < len(c.PvcNames); ix++ {
		// Confirm that the PVC has been created
		Expect(c.DeleteErrs[ix]).To(BeNil(), "failed to delete PVC %s", c.PvcNames[ix])

		// Confirm the PVC has been deleted.
		pvc, _ := k8stest.GetPVC(c.PvcNames[ix], common.NSDefault)
		Expect(pvc).ToNot(BeNil())

		// Wait for the PVC to be deleted.
		Eventually(func() bool {
			return k8stest.IsPVCDeleted(c.PvcNames[ix], common.NSDefault)
		},
			defTimeoutSecs, // timeout
			"1s",           // polling interval
		).Should(Equal(true))

		// Wait for the PV to be deleted.
		Eventually(func() bool {
			// This check is required here because it will check for pv name
			// when pvc is in pending state at that time we will not
			// get pv name inside pvc spec i.e pvc.Spec.VolumeName
			if pvc.Spec.VolumeName != "" {
				return k8stest.IsPVDeleted(pvc.Spec.VolumeName)
			}
			return true
		},
			"360s", // timeout
			"1s",   // polling interval
		).Should(Equal(true))
		// Wait for the MSV to be deleted.
		Eventually(func() bool {
			return custom_resources.IsMsVolDeleted(c.Uuid[ix])
		},
			"360s", // timeout
			"1s",   // polling interval
		).Should(Equal(true))

	}
}

// deletePVC will delete all pvc
func (c *PrimitiveMsvFuzzConfig) deletePVC() {
	for _, pvc := range c.PvcNames {
		k8stest.RmPVC(pvc, c.ScName, common.NSDefault)
	}
}

// concurrentMsvOperationInIteration will create/delete multiple volumes
// in concurrent. It creates and deletes volumes in Iterations.
func (c *PrimitiveMsvFuzzConfig) concurrentMsvOperationInIteration() {
	for ix := 0; ix < c.Iterations; ix++ {
		c.ConcurrentMsvFuzz()
	}
}

// waitForMspUsedSize verify msp used size
func (c *PrimitiveMsvFuzzConfig) waitForMspUsedSize(size int64) {
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
	for ix := 0; ix < (300+sleepTimeSec-1)/sleepTimeSec; ix++ {
		time.Sleep(time.Duration(sleepTimeSec) * time.Second)
		pool, err := custom_resources.GetMsPool(poolName)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to get mayastor pool %s %v", poolName, err))
		if pool.Status.Used == usedSize {
			return nil
		}
	}
	return errors.Errorf("pool %s used size did not reconcile in %d seconds", poolName, timeoutSec)
}

// createPvcParallel will create volumes in parallet
func (c *PrimitiveMsvFuzzConfig) createPvcParallel() *PrimitiveMsvFuzzConfig {
	// Create the volumes
	var wg sync.WaitGroup
	wg.Add(len(c.PvcNames))
	for i := 0; i < len(c.PvcNames); i++ {
		go k8stest.CreatePvc(&c.OptsList[i], &c.CreateErrs[i], &c.Uuid[i], &wg)
	}
	wg.Wait()

	logf.Log.Info("Finished calling the create methods for all PVC candidates.")

	return c
}

// deletePvcParallel will remove volumes
func (c *PrimitiveMsvFuzzConfig) deletePvcParallel() *PrimitiveMsvFuzzConfig {
	// Create the volumes
	var wg sync.WaitGroup
	wg.Add(len(c.PvcNames))
	for i := 0; i < len(c.PvcNames); i++ {
		go k8stest.DeletePvc(c.PvcNames[i], common.NSDefault, &c.DeleteErrs[i], &wg)
	}
	wg.Wait()
	logf.Log.Info("Finished calling the delete methods for all PVC candidates.")

	return c
}

// createFioPods will create fio pods and runs fio on all mounted volumes.
func (c *PrimitiveMsvFuzzConfig) createFioPods() {
	durationSecs := 60
	volumeFileSizeMb := 50

	for ix, podName := range c.FioPodNames {
		args := []string{
			"--",
			"--time_based",
			fmt.Sprintf("--runtime=%d", durationSecs),
			fmt.Sprintf("--filename=%s", common.FioFsFilename),
			fmt.Sprintf("--size=%dm", volumeFileSizeMb),
		}
		fioArgs := append(args, common.GetFioArgs()...)

		// fio pod container
		podContainer := coreV1.Container{
			Name:  podName,
			Image: "mayadata/e2e-fio",
			Args:  fioArgs,
		}

		// volume claim details
		volume := coreV1.Volume{
			Name: "ms-volume",
			VolumeSource: coreV1.VolumeSource{
				PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
					ClaimName: c.PvcNames[ix],
				},
			},
		}

		podObj, err := k8stest.NewPodBuilder().
			WithName(podName).
			WithNamespace(common.NSDefault).
			WithRestartPolicy(coreV1.RestartPolicyNever).
			WithContainer(podContainer).
			WithVolume(volume).
			WithVolumeDeviceOrMount(common.VolFileSystem).Build()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Generating fio pod definition %s", podName))
		Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

		// Create fio pod
		_, err = k8stest.CreatePod(podObj, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", podName)

		// Wait for the fio Pod to transition to error
		Eventually(func() bool {
			return k8stest.IsPodRunning(podName, common.NSDefault)
		},
			defTimeoutSecs,
			"1s",
		).Should(Equal(false))
	}
}

// delete all fio pods
func (c *PrimitiveMsvFuzzConfig) deleteFioPods() {
	for _, podName := range c.FioPodNames {
		// Delete the fio pod
		err := k8stest.DeletePod(podName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "failed to delete fio pod")
	}
}

// Check that all volumes have been created successfully,
// that none of them are in the pending state,
// that all of them have the right size
func (c *PrimitiveMsvFuzzConfig) verifyVolumesCreation() {
	for ix := 0; ix < len(c.PvcNames); ix++ {
		// Confirm that the PVC has been created
		Expect(c.CreateErrs[ix]).To(BeNil(), "failed to create PVC %s", c.PvcNames[ix])

		namespace := common.NSDefault
		volName := c.PvcNames[ix]
		// Wait for the PVC to be bound.
		Eventually(func() *coreV1.PersistentVolumeClaim {
			pvc, _ := k8stest.GetPVC(volName, namespace)
			return pvc
		},
			defTimeoutSecs, // timeout
			"1s",           // polling interval
		).Should(Not(BeNil()))
	}
}
