package primitive_msp_deletion

import (
	"fmt"
	"strings"
	"testing"

	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
)

func TestPrimitiveMspDeletionTest(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Primitive Mayastorpool Deletion Test", "primitive_msp_deletion")
}

var defTimeoutSecs = "60s"

func primitiveMspDeletionTest(protocol common.ShareProto, volumeType common.VolumeType, fsType common.FileSystemType, mode storageV1.VolumeBindingMode) {

	params := e2e_config.GetConfig().PrimitiveMspDelete

	scName := strings.ToLower(fmt.Sprintf("primitive-ms-deletion-sc-%s-%s", string(protocol), volumeType))
	volName := strings.ToLower(fmt.Sprintf("primitive-ms-deletion-%s-%s", string(protocol), volumeType))

	// Create storage class
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithReplicas(params.Replicas).
		WithVolumeBindingMode(mode).
		WithProtocol(protocol).
		WithFileSystemType(fsType).
		BuildAndCreate()

	Expect(err).To(BeNil(), "Storage class creation failed")

	// Create volume
	k8stest.MkPVC(common.DefaultVolumeSizeMb, volName, scName, volumeType, common.NSDefault)

	// Create fio pod
	fioPodName := fmt.Sprintf("fio-%s", volName)
	err = createFioPod(fioPodName, volName)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Creating fio pod %s %v", fioPodName, err))

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		params.FioPodTimeoutSecs,
		"1s",
	).Should(Equal(true))

	// Wait for fio pod to transition to completed state
	err = k8stest.WaitPodComplete(fioPodName, params.SleepTimeSecs, params.TimeoutSecs)
	Expect(err).ToNot(HaveOccurred())

	// Check for pool usage
	Eventually(func() int {
		poolUsage, err := k8stest.GetPoolUsageInCluster()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to get pool usage %v", err))
		return int(poolUsage)
	},
		params.PoolUsageTimeoutSecs, // timeout
		"1s",                        // polling interval
	).ShouldNot(Equal(0), "Pool usage is 0")

	// List replicas in the cluster
	replicas, err := k8stest.ListReplicasInCluster()
	if err != nil {
		logf.Log.Info("ResourceEachCheck: failed to retrieve list of replicas")
	}
	logf.Log.Info("ResourceCheck:", "num replicas", len(replicas))
	Expect(len(replicas) == params.Replicas).To(BeTrue(), "Replicas not found")

	// List pools in the cluster
	pools, err := custom_resources.ListMsPools()

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	// Delete the storage class
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)

	Eventually(func() int {
		poolUsage, err := k8stest.GetPoolUsageInCluster()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to get pool usage %v", err))
		return int(poolUsage)
	},
		params.PoolUsageTimeoutSecs, // timeout
		"1s",                        // polling interval
	).Should(Equal(0), "Pool usage is not 0")

	// Delete mayastorpool
	Eventually(func() error {
		for _, pool := range pools {
			err = custom_resources.DeleteMsPool(pool.Name)
			if err != nil {
				return err
			}
		}
		return nil
	},
		params.PoolDeleteTimeoutSecs, // timeout
		"1s",                         // polling interval
	).Should((BeNil()))

	// Restart mayastor pods
	err = k8stest.RestartMayastorPods(params.MayastorRestartTimeout)
	Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")

	// Create mayastorpools
	Eventually(func() error {
		for _, pool := range pools {
			_, err = custom_resources.CreateMsPool(pool.Name, pool.Spec.Node, pool.Spec.Disks)
			if err != nil {
				return err
			}
		}
		return nil
	},
		params.PoolCreateTimeoutSecs, // timeout
		"1s",                         // polling interval
	).Should((BeNil()))

	// RestoreConfiguredPools (re)create pools as defined by the configuration.
	// As part of the tests we may modify the pools, in such test cases
	// the test should delete all pools and recreate the configured set of pools.
	err = k8stest.RestoreConfiguredPools()
	Expect(err).To(BeNil(), "Not all pools are online after restoration")

}

var _ = Describe("Primitive mayastorpool deletion test", func() {

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

	It("Should verify mayastor pool deletion", func() {
		primitiveMspDeletionTest(common.ShareProtoNvmf, common.VolFileSystem, common.XfsFsType, storageV1.VolumeBindingImmediate)
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

// createFioPod created fio pod obj and create fio pod
func createFioPod(podName string, volName string) error {
	durationSecs := 30
	volumeFileSizeMb := 50
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
		return nil
	}
	return err
}
