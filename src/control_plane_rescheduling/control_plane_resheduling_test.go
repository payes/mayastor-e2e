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

	params := e2e_config.GetConfig().BasicVolumeIO
	scName := fmt.Sprintf("reshedule-sc-%s", protocol)
	var volNames []string
	var fioPodNames []string
	deploymentName := "moac"
	var replicas int32
	logf.Log.Info("Test", "parameters", params)

	// Create storage class
	err := createStoragClass(scName, mode, common.NSDefault, params.Replicas, protocol, fsType)
	Expect(err).To(BeNil(), "Storage class creation failed")

	// Create volumes
	for ix := 1; ix <= e2e_config.GetConfig().ControlPlaneRescheduling.MayastorVolumeCount; ix += 1 {
		volName := fmt.Sprintf("reshedule-vol-%d", ix)
		volNames = append(volNames, volName)
		k8stest.MkPVC(common.DefaultVolumeSizeMb, volName, scName, volumeType, common.NSDefault)
	}

	// Check status of volume
	for _, volName := range volNames {
		err := checkVolumeStatus(volName)
		Expect(err).ToNot(HaveOccurred())
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
		waitPodComplete(fioPodName)
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

// createStorageClass creates storageclass object
// and creates storage class.
func createStoragClass(scName string, mode storageV1.VolumeBindingMode, namespace string, replicas int, protocol common.ShareProto, filetype common.FileSystemType) error {
	// Create storage class obj
	scObj, err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(namespace).
		WithReplicas(replicas).
		WithVolumeBindingMode(mode).
		WithProtocol(protocol).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating storage class definition %s", scName)
	if filetype != "" {
		scObj.Parameters[string(common.ScFsType)] = string(filetype)
	}
	// Create storage class
	err = k8stest.CreateSc(scObj)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	return err
}

// checkVolumeStatus confirms PVC has been created,
// wait for PVC to bound, wait for the PV to be provisioned
// and bound and also checks MSV to be provisioned and healthy
func checkVolumeStatus(volName string) error {
	// Confirm the PVC has been created.
	pvc, err := k8stest.GetPVC(volName, common.NSDefault)
	Expect(err).To(BeNil(), "PVC creation failed")
	Expect(pvc).ToNot(BeNil(), "PVC creation failed")

	// Wait for the PVC to be bound.
	Eventually(func() coreV1.PersistentVolumeClaimPhase {
		return k8stest.GetPvcStatusPhase(volName, common.NSDefault)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(coreV1.ClaimBound))

	// Refresh the PVC contents, so that we can get the PV name.
	pvc, err = k8stest.GetPVC(volName, common.NSDefault)
	Expect(err).To(BeNil(), "PVC content is nil")
	Expect(pvc).ToNot(BeNil(), "PVC content is nil")

	// Wait for the PV to be provisioned
	Eventually(func() *coreV1.PersistentVolume {
		pv, err := k8stest.GetPV(pvc.Spec.VolumeName)
		if err != nil {
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

	// Wait for the MSV to be provisioned
	Eventually(func() *k8stest.MayastorVolStatus {
		return k8stest.GetMSV(string(pvc.ObjectMeta.UID))
	},
		defTimeoutSecs, //timeout
		"1s",           // polling interval
	).Should(Not(BeNil()))

	// Wait for the MSV to be healthy
	Eventually(func() string {
		return k8stest.GetMsvState(string(pvc.ObjectMeta.UID))
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal("healthy"))

	return err
}

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

// waitPodComplete waits until all fio pods are in completed state
func waitPodComplete(fioPodName string) {
	const sleepTimeSecs = 3
	const timeoutSecs = 360
	var podPhase coreV1.PodPhase
	var err error

	logf.Log.Info("Waiting for pod to complete", "name", fioPodName)
	for ix := 0; ix < (timeoutSecs+sleepTimeSecs-1)/sleepTimeSecs; ix++ {
		time.Sleep(sleepTimeSecs * time.Second)
		podPhase, err = k8stest.CheckPodCompleted(fioPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to access pods status %s %v", fioPodName, err))
		if podPhase == coreV1.PodSucceeded {
			return
		}
		Expect(podPhase == coreV1.PodRunning).To(BeTrue(), fmt.Sprintf("Unexpected pod phase %v", podPhase))
	}
	Expect(podPhase == coreV1.PodSucceeded).To(BeTrue(), fmt.Sprintf("pod did not complete, phase %v", podPhase))
}
