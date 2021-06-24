package primitive_msp_state

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	"mayastor-e2e/common/mayastorclient/grpc"

	coreV1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// createSC will create storageclass
func (c *mspStateConfig) createSC() {
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
func (c *mspStateConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createPVC will create pvc
func (c *mspStateConfig) createPVC() *mspStateConfig {
	// Create the volumes
	c.uuid = k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, common.VolFileSystem, common.NSDefault)
	return c
}

// deletePVC will delete all pvc
func (c *mspStateConfig) deletePVC() {
	k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
}

// getMsvDetails will set pools and msv capacity
func (c *mspStateConfig) getMsvDetails() *mspStateConfig {
	msv, err := custom_resources.GetMsVol(c.uuid)
	Expect(err).To(BeNil(), "failed to access volume %s, error=%v", c.uuid, err)
	c.msvSize = msv.Status.Size
	for _, replica := range msv.Status.Replicas {
		c.poolNames = append(c.poolNames, replica.Pool)
	}
	return c
}

// update pool and node address
func (c *mspStateConfig) getPoolAndNodeAddress() *mspStateConfig {
	nodes, err := k8stest.GetNodeLocs()
	if err != nil {
		logf.Log.Info("list nodes failed", "error", err)
	}
	// List Pools by CRDs
	crdPools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	var poolName, nodeName string
	for _, pool := range c.poolNames {
		for _, crdPool := range crdPools {
			if pool != crdPool.Name {
				poolName = crdPool.Name
				nodeName = crdPool.Spec.Node
			}
		}
	}
	c.newPoolName = poolName
	for _, node := range nodes {
		if !node.MayastorNode {
			continue
		}
		if node.NodeName == nodeName {
			c.nodeAddress = node.IPAddress
		}
	}
	return c
}

// verifyMspUsedSize will verify msp used size
func (c *mspStateConfig) verifyMspUsedSize() {
	for _, poolname := range c.poolNames {
		pool, err := custom_resources.GetMsPool(poolname)
		Expect(err).ToNot(HaveOccurred(), "GetMsPool failed %s", poolname)
		Expect(verifyMspUsedSizeValue(pool, c.msvSize)).Should(Equal(true))
	}
}

// createFioPods will create fio pods and run fio concurrently on all mounted volumes
func (c *mspStateConfig) createFioPods() {
	var volMounts []coreV1.VolumeMount
	var volDevices []coreV1.VolumeDevice
	var volFioArgs [][]string

	// fio pod container
	podContainer := coreV1.Container{
		Name:  c.fioPodName,
		Image: common.GetFioImage(),
		// Image:           "mayadata/e2e-fio",
		ImagePullPolicy: coreV1.PullAlways,
		Args:            []string{"sleep", "1000000"},
	}
	var volumes []coreV1.Volume
	// volume claim details
	volume := coreV1.Volume{
		Name: fmt.Sprintf("ms-volume-%s", c.pvcName),
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: c.pvcName,
			},
		},
	}
	volumes = append(volumes, volume)
	// volume mount or device
	if c.volType == common.VolFileSystem {
		mount := coreV1.VolumeMount{
			Name:      fmt.Sprintf("ms-volume-%s", c.pvcName),
			MountPath: fmt.Sprintf("/volume-%s", c.pvcName),
		}
		volMounts = append(volMounts, mount)
		volFioArgs = append(volFioArgs, []string{
			fmt.Sprintf("--filename=/volume-%s/fio-test-file", c.pvcName),
			fmt.Sprintf("--size=%dm", common.DefaultFioSizeMb),
		})
	} else {
		device := coreV1.VolumeDevice{
			Name:       fmt.Sprintf("ms-volume-%s", c.pvcName),
			DevicePath: fmt.Sprintf("/dev/sdm-%s", c.pvcName),
		}
		volDevices = append(volDevices, device)
		volFioArgs = append(volFioArgs, []string{
			fmt.Sprintf("--filename=/dev/sdm-%s", c.pvcName),
		})
	}

	podObj, err := k8stest.NewPodBuilder().
		WithName(c.fioPodName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolumes(volumes).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", c.fioPodName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

	switch c.volType {
	case common.VolFileSystem:
		podObj.Spec.Containers[0].VolumeMounts = volMounts
	case common.VolRawBlock:
		podObj.Spec.Containers[0].VolumeDevices = volDevices
	}

	// Construct argument list for fio to run a single instance of fio,
	// with multiple jobs, one for each volume.
	var podArgs []string

	// 1) directives for fio job
	podArgs = append(podArgs, "--")
	podArgs = append(podArgs, common.GetDefaultFioArguments()...)
	podArgs = append(podArgs, []string{
		"--time_based",
		fmt.Sprintf("--runtime=%d", c.duration),
		fmt.Sprintf("--thinktime=%d", c.thinkTime),
	}...,
	)

	// 2) per volume directives (filename, size, and testname)
	for ix, v := range volFioArgs {
		podArgs = append(podArgs, v...)
		podArgs = append(podArgs, fmt.Sprintf("--name=benchtest-%d", ix))
	}
	podArgs = append(podArgs, "&")

	logf.Log.Info("fio", "arguments", podArgs)
	podObj.Spec.Containers[0].Args = podArgs

	// Create first fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", c.fioPodName)
	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(c.fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
}

// delete all fio pods
func (c *mspStateConfig) deleteFioPods() {

	// Delete the fio pod
	err := k8stest.DeletePod(c.fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete fio pod")

}

// verifyMspCrdAndGrpcState verifies the msp details from grpc and crd
func verifyMspCrdAndGrpcState() {

	nodes, err := k8stest.GetNodeLocs()
	if err != nil {
		logf.Log.Info("list nodes failed", "error", err)
		return
	}

	// List Pools by CRDs
	crdPools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	crPools := map[string]v1alpha1.MayastorPool{}
	for _, crdPool := range crdPools {
		crPools[crdPool.Name] = crdPool
	}

	for _, node := range nodes {
		if !node.MayastorNode {
			continue
		}
		addrs := []string{node.IPAddress}
		grpcPools, err := mayastorclient.ListPools(addrs)
		Expect(err).ToNot(HaveOccurred(), "failed to list pools via grpc")

		if err == nil && len(grpcPools) != 0 {
			for _, gPool := range grpcPools {
				Expect(verifyMspState(crPools[gPool.Name], gPool)).Should(Equal(true))
				Expect(verifyMspCapacity(crPools[gPool.Name], gPool)).Should(Equal(true))
				Expect(verifyMspUsedSpace(crPools[gPool.Name], gPool)).Should(Equal(true))
			}
		} else {
			logf.Log.Info("pools", "count", len(grpcPools), "error", err)
		}
	}
}

// verifyMspState will verify msp state via crd  and grpc
// gRPC report msp status as "POOL_UNKNOWN","POOL_ONLINE","POOL_DEGRADED","POOL_FAULTED"
// CRD report msp status as "unknown", "online", "degraded", "faulted"
// CRDs report as online
func verifyMspState(crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.State == grpcStateToCrdstate(grpcPool.State) {
		status = true
	}
	return status
}

// verifyMspCapacity will verify msp capacity via crd  and grpc
func verifyMspCapacity(crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.Capacity == int64(grpcPool.Capacity) {
		status = true
	}
	return status
}

// verifyMspUsedSpace will verify msp used size via crd  and grpc
func verifyMspUsedSpace(crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.Used == int64(grpcPool.Used) {
		status = true
	}
	return status
}

// verifyMspUsedSize will verify msp used size
func verifyMspUsedSizeValue(crPool v1alpha1.MayastorPool, size int64) bool {
	var status bool
	if crPool.Status.Used == size {
		status = true
	} else {
		logf.Log.Info("Pool", "name", crPool.Name, "Used", crPool.Status.Used, "Expected Used", size)
	}
	return status
}

// update replica
func (c *mspStateConfig) updateReplica() {
	err := mayastorclient.CreateReplica(c.nodeAddress, c.uuid, uint64(c.msvSize), c.newPoolName)
	Expect(err).ToNot(HaveOccurred(), "failed to update replica")

}

// verifyMspUsedSize will verify msp used size
func (c *mspStateConfig) verifyNewlyAddedPoolUsedSize() {
	pool, err := custom_resources.GetMsPool(c.newPoolName)
	Expect(err).ToNot(HaveOccurred(), "GetMsPool failed %s", c.newPoolName)
	Expect(verifyMspUsedSizeValue(pool, c.msvSize)).Should(Equal(true))
}

func grpcStateToCrdstate(mspState grpc.PoolState) string {
	switch mspState {
	case 0:
		return "unknown"
	case 1:
		return "online"
	case 2:
		return "degraded"
	default:
		return "faulted"
	}
}
