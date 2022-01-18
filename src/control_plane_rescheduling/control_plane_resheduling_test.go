package control_plane_rescheduling

import (
	"fmt"
	"testing"
	"time"

	storageV1 "k8s.io/api/storage/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	coreV1 "k8s.io/api/core/v1"
)

func TestControlPlaneRescheduling(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Control Plane Rescheduling Test", "control_plane_rescheduling")
}

var defTimeoutSecs = "60s"

func controlPlaneReschedulingTest(protocol common.ShareProto, volumeType common.VolumeType, fsType common.FileSystemType, mode storageV1.VolumeBindingMode) {
	const sleepTimeSecs = 3
	const timeoutSecs = 360
	scName := fmt.Sprintf("reshedule-sc-%s", protocol)
	var volNames []string
	var fioPodNames []string
	deploymentName := "moac"
	var replicas int32

	// Create storage class
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithReplicas(common.DefaultReplicaCount).
		WithVolumeBindingMode(mode).
		WithProtocol(protocol).
		WithFileSystemType(fsType).
		BuildAndCreate()

	Expect(err).To(BeNil(), "Storage class creation failed")

	// Create volumes
	for ix := 1; ix <= e2e_config.GetConfig().ControlPlaneRescheduling.MayastorVolumeCount; ix += 1 {
		volName := fmt.Sprintf("reshedule-vol-%d", ix)
		volNames = append(volNames, volName)
		k8stest.MkPVC(common.DefaultVolumeSizeMb, volName, scName, volumeType, common.NSDefault)
	}

	// Create pod
	for ix, volName := range volNames {
		fioPodName := fmt.Sprintf("fio-%s-%d", volName, ix)
		fioPodNames = append(fioPodNames, fioPodName)
		err := createFioPod(fioPodName, volName)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Creating fio pod %s %v", fioPodName, err))
	}

	// Wait untill all fio pods are in running state
	for _, fioPodName := range fioPodNames {
		// Wait for the fio Pod to transition to running
		Eventually(func() bool {
			return k8stest.IsPodRunning(fioPodName, common.NSDefault)
		},
			defTimeoutSecs,
			"1s",
		).Should(Equal(true))
	}

	replicas = 0
	// Scale down moac deployment to 0 replicas
	k8stest.SetDeploymentReplication(deploymentName, e2e_config.GetConfig().Platform.MayastorNamespace, &replicas)

	var moacPodName []string
	// Check presence of moac pod
	Eventually(func() bool {
		moacPodName, _ = k8stest.GetMoacPodName()
		return len(moacPodName) == 0
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	logf.Log.Info("Moac pod is removed")

	// After scaling down moac deployment sleep for 10 sec.
	time.Sleep(10 * time.Second)

	// Check if all fio pods are in running state or not
	for _, fioPodName := range fioPodNames {
		// Wait for the fio Pod to transition to running
		Eventually(func() bool {
			return k8stest.IsPodRunning(fioPodName, common.NSDefault)
		},
			defTimeoutSecs,
			"1s",
		).Should(Equal(true))
	}

	replicas = 1
	// Scale up moac deployment to 1 replicas
	k8stest.SetDeploymentReplication(deploymentName, e2e_config.GetConfig().Platform.MayastorNamespace, &replicas)

	// Check presence of moac pod
	Eventually(func() bool {
		moacPodName, _ = k8stest.GetMoacPodName()
		return len(moacPodName) == 1
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	logf.Log.Info("Moac pod is now present")

	Eventually(func() bool {
		podName := moacPodName[0]
		return k8stest.IsPodRunning(podName, common.NSMayastor())
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	logf.Log.Info("Moac pod is in running state")

	// Wait for fio pods to get into completed state
	for _, fioPodName := range fioPodNames {
		err = k8stest.WaitPodComplete(fioPodName, sleepTimeSecs, timeoutSecs)
		Expect(err).ToNot(HaveOccurred())
	}

	// Cleanup of resources
	for ix, fioPodName := range fioPodNames {

		// Delete the fio pod
		err := k8stest.DeletePod(fioPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred())

		// Delete the volume
		k8stest.RmPVC(volNames[ix], scName, common.NSDefault)

	}

	// Delete the storage class
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

var _ = Describe("Control Plane Rescheduling Test", func() {

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

	It("Should verify Control Plane Rescheduling ", func() {
		controlPlaneReschedulingTest(common.ShareProtoNvmf, common.VolFileSystem, common.XfsFsType, storageV1.VolumeBindingImmediate)
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
	durationSecs := 60
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
		Image: common.GetFioImage(),
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
