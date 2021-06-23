package primitive_msp_state

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	"time"

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

	// 1) directives for all fio jobs
	podArgs = append(podArgs, "--")
	podArgs = append(podArgs, common.GetDefaultFioArguments()...)
	podArgs = append(podArgs, []string{
		"--time_based",
		fmt.Sprintf("--runtime=%d", int(c.duration.Seconds())),
		fmt.Sprintf("--thinktime=%d", int(c.thinkTime.Microseconds())),
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

// delete all fip pods
func (c *mspStateConfig) deleteFioPods() {

	// Delete the fio pod
	err := k8stest.DeletePod(c.fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete fio pod")

}

// check fio pods completion status
func (c *mspStateConfig) checkFioPodsComplete() {

	logf.Log.Info("Waiting for run to complete", "duration", c.duration, "timeout", c.timeout)
	tSecs := 0
	var phase coreV1.PodPhase
	var err error
	for {
		if tSecs > int(c.timeout.Seconds()) {
			break
		}
		time.Sleep(1 * time.Second)
		tSecs += 1
		phase, err = k8stest.CheckPodCompleted(c.fioPodName, common.NSDefault)
		Expect(err).To(BeNil(), "CheckPodComplete got error %s", err)
		if phase != coreV1.PodRunning {
			break
		}
	}
	Expect(phase == coreV1.PodSucceeded).To(BeTrue(), "fio pod phase is %s", phase)
	logf.Log.Info("fio completed", "duration", tSecs)
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
				verifyMspState(crPools[gPool.Name], gPool)
				verifyMspCapacity(crPools[gPool.Name], gPool)
				verifyMspUsedSpace(crPools[gPool.Name], gPool)
			}
		} else {
			logf.Log.Info("pools", "count", len(grpcPools), "error", err)
		}

	}

}

// verifyMspState will verify msp state
func verifyMspState(crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.State == grpcPool.State.String() {
		status = true
	}
	return status
}

// verifyMspCapacity will verify msp capacity
func verifyMspCapacity(crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.Capacity == int64(grpcPool.Capacity) {
		status = true
	}
	return status
}

// verifyMspCapacity will verify msp used size
func verifyMspUsedSpace(crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.Used == int64(grpcPool.Used) {
		status = true
	}
	return status
}
