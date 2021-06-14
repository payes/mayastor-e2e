package expand_msp_disk

import (
	"fmt"
	"testing"
	"time"

	errors "github.com/pkg/errors"

	agent "mayastor-e2e/common/e2e-agent"
	"mayastor-e2e/common/e2e_config"

	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/k8stest"
)

func TestDiskPartitioning(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Expand MSP Disk", "expand_msp_disk")
}

var defTimeoutSecs = "60s"

func diskPartitioningTest(protocol common.ShareProto, volumeType common.VolumeType, fsType common.FileSystemType, mode storageV1.VolumeBindingMode, replica int) {
	const timoSecs = 360
	const timoSleepSecs = 3
	var expandMspDisk = e2e_config.GetConfig().ExpandMspDisk

	scName := fmt.Sprintf("expand-msp-disk-sc-%d", replica)
	volName := fmt.Sprintf("expand-msp-disk-vol-%d", replica)
	// List Pools
	pools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools failed")

	logf.Log.Info("Deleting Mayastor Pool")
	poolsDeleted := k8stest.DeleteAllPools()
	Expect(poolsDeleted).To(BeTrue())

	// List nodes
	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "Getting nodes failed")

	time.Sleep(10 * time.Second)

	// Create Pool using partitioned disk i.e /dev/sdb1
	command := "parted --script " + expandMspDisk.DiskPath + " mklabel gpt mkpart primary ext4 " + expandMspDisk.PartitionStartSize + " " + expandMspDisk.PartitionEndSize
	for _, node := range nodes {
		if node.MayastorNode && !node.MasterNode {
			err = agent.DiskPartition(node.IPAddress, command)
			Expect(err).ToNot(HaveOccurred(), "Disk Partitioning failed for node %s:", node.NodeName)
		}
	}
	time.Sleep(10 * time.Second)
	// Create pool using partitioned disk
	logf.Log.Info("Creating Mayastor Pool")
	for _, pool := range pools {
		diskPath := make([]string, 1)
		diskPath[0] = pool.Spec.Disks[0] + "1"
		_, err = custom_resources.CreateMsPool(pool.Name, pool.Spec.Node, diskPath)
		Expect(err).ToNot(HaveOccurred(), "Pool creation failed for node %s:", pool.Spec.Node)
	}

	// Wait for pools to be online
	for ix := 0; ix < timoSecs/timoSleepSecs; ix++ {
		time.Sleep(timoSleepSecs * time.Second)
		err := custom_resources.CheckAllMsPoolsAreOnline()
		if err == nil {
			break
		}
	}
	Expect(err).To(BeNil(), "One or more pools are offline")

	// Storage capacity of pool before resizing the disk
	pools, err = custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools failed")
	capacityBeforeDiskResize := map[string]int64{}
	for _, pool := range pools {
		capacityBeforeDiskResize[pool.Name] = pool.Status.Capacity
	}
	// Create storageclass
	err = createStorageClass(scName, mode, common.NSDefault, replica, protocol)
	Expect(err).To(BeNil(), "Storage class creation failed")

	// Create PVC
	k8stest.MkPVC(common.DefaultVolumeSizeMb, volName, scName, volumeType, common.NSDefault)

	// Create fio-pod
	fioPodName := fmt.Sprintf("fio-%s", volName)
	err = createFioPod(fioPodName, volName, expandMspDisk.Duration, expandMspDisk.VolSizeMb)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Creating fio pod %s %v", fioPodName, err))

	// Check if fio pod is in running state or not
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

	// Resize the partitioned disk
	command = "parted " + expandMspDisk.DiskPath + " resizepart 1 " + expandMspDisk.ResizePartitionDisk
	for _, node := range nodes {
		if node.MayastorNode && !node.MasterNode {
			err = agent.DiskPartition(node.IPAddress, command)
			Expect(err).ToNot(HaveOccurred(), "Disk Resizing failed for node %s:", node.NodeName)
		}
	}

	// Wait for pools to be online
	for ix := 0; ix < timoSecs/timoSleepSecs; ix++ {
		time.Sleep(timoSleepSecs * time.Second)
		err := custom_resources.CheckAllMsPoolsAreOnline()
		if err == nil {
			break
		}
	}
	capacityAfterDiskResize := map[string]int64{}

	// Store capacity of pools after
	// resizing the partitioned disk
	pools, err = custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools failed")
	for _, pool := range pools {
		capacityAfterDiskResize[pool.Name] = pool.Status.Capacity
	}

	// Check capacity of pool is equal or not after resizing the partitioned disk.
	// This check ensures that we have chosen a long enough runtime for fio,
	// so that it runs over the duration of the test.
	err = compareDiskSize(capacityBeforeDiskResize, capacityAfterDiskResize)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// Check if fio pod is in running state or not
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

	// Wait for fio pod to get into completed state
	err = k8stest.WaitPodComplete(fioPodName, timoSleepSecs, timoSecs)
	Expect(err).ToNot(HaveOccurred())
	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	// Delete the storage class
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)

	// Delete mayastorpool
	logf.Log.Info("Deleting Mayastor Pool")
	for _, pool := range pools {
		err := custom_resources.DeleteMsPool(pool.Name)
		Expect(err).ToNot(HaveOccurred())
	}
	// Added sleep so that partitioned disk
	// can be deleted easily
	time.Sleep(10 * time.Second)

	// Delete partitioned disk
	command = "parted " + expandMspDisk.DiskPath + " rm 1"
	for _, node := range nodes {
		if node.MayastorNode && !node.MasterNode {
			err = agent.DiskPartition(node.IPAddress, command)
			Expect(err).ToNot(HaveOccurred(), "Disk deletetion failed for node %s:", node.NodeName)
		}
	}

	// RestoreConfiguredPools (re)create pools as defined by the configuration.
	// As part of the tests we may modify the pools, in such test cases
	// the test should delete all pools and recreate the configured set of pools.
	err = k8stest.RestoreConfiguredPools()
	Expect(err).To(BeNil(), "Not all pools are online after restoration")
}

var _ = Describe("Expand MSP disk test", func() {

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

	It("Should verify expansion of MSP disk with multiple replicas", func() {
		diskPartitioningTest(common.ShareProtoNvmf, common.VolFileSystem, common.XfsFsType, storageV1.VolumeBindingImmediate, 3)
	})

})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})

// createStorageClass creates storageclass object
// and creates storage class.
func createStorageClass(scName string, mode storageV1.VolumeBindingMode, namespace string, replicas int, protocol common.ShareProto) error {
	// Create storage class
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(namespace).
		WithReplicas(replicas).
		WithVolumeBindingMode(mode).
		WithProtocol(protocol).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	return err
}

// createFioPod created fio pod obj and create fio pod
func createFioPod(podName string, volName string, durationSecs string, volumeFileSizeMb string) error {
	args := []string{
		"--",
		"--time_based",
		fmt.Sprintf("--runtime=%s", durationSecs),
		fmt.Sprintf("--filename=%s", common.FioFsFilename),
		fmt.Sprintf("--size=%sm", volumeFileSizeMb),
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
				ClaimName: volName,
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

	if err != nil {
		return err
	}
	return nil
}

// compareDiskSize compares disk
func compareDiskSize(capacityBeforeDiskResize map[string]int64, capacityAfterDiskResize map[string]int64) error {
	for poolName, poolCapacity := range capacityBeforeDiskResize {
		if capacityAfterDiskResize[poolName] != poolCapacity {
			return errors.Errorf("Capacity of pool: %s changed, capacity before disk partition : %d and capacity after disk partition : %d", poolName, poolCapacity, capacityAfterDiskResize[poolName])
		}
	}
	return nil
}
