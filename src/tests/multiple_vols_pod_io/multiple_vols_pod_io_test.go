package multiple_vols_pod_io

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "120s"

func TestMultipleVolumeIO(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Multiple volumes single pod IO tests", "multiple_vols_pod_io")
}

func multipleVolumeIOTest(replicas int, volumeCount int, protocol common.ShareProto,
	volumeType common.VolumeType, volSize int, binding storageV1.VolumeBindingMode, local bool,
	timeout time.Duration, fioLoops int) {
	logf.Log.Info("MultipleVolumeIOTest", "replicas", replicas, "volumeCount", volumeCount, "protocol", protocol, "volumeType", volumeType, "binding", binding)
	scName := strings.ToLower(fmt.Sprintf("msv-repl-%d-%s-%s-%s", replicas, string(protocol), volumeType, binding))
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithReplicas(replicas).
		WithProtocol(protocol).
		WithNamespace(common.NSDefault).
		WithVolumeBindingMode(binding).
		WithLocal(local).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "failed to create storage class %s", scName)

	var volNames []string
	var volumes []coreV1.Volume
	var volMounts []coreV1.VolumeMount
	var volDevices []coreV1.VolumeDevice
	var volFioArgs [][]string

	// Create the volumes and associated bits
	for ix := 1; ix <= volumeCount; ix += 1 {
		volName := fmt.Sprintf("ms-vol-%s-%d", protocol, ix)
		uid, err := k8stest.MkPVC(volSize, volName, scName, volumeType, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", volName)
		logf.Log.Info("Volume", "uid", uid)
		volNames = append(volNames, volName)

		vol := coreV1.Volume{
			Name: fmt.Sprintf("ms-volume-%d", ix),
			VolumeSource: coreV1.VolumeSource{
				PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
					ClaimName: volName,
				},
			},
		}

		volumes = append(volumes, vol)

		vname := fmt.Sprintf("ms-volume-%d", ix)
		if volumeType == common.VolFileSystem {
			mount := coreV1.VolumeMount{
				Name:      vname,
				MountPath: fmt.Sprintf("/volume-%d", ix),
			}
			volMounts = append(volMounts, mount)
			volFioArgs = append(volFioArgs, []string{
				fmt.Sprintf("--name=%s", vname),
				fmt.Sprintf("--filename=/volume-%d/%s.test", ix, vname),
			})
		} else {
			device := coreV1.VolumeDevice{
				Name:       vname,
				DevicePath: fmt.Sprintf("/dev/sdm-%d", ix),
			}
			volDevices = append(volDevices, device)
			volFioArgs = append(volFioArgs, []string{
				fmt.Sprintf("--name=%s", vname),
				fmt.Sprintf("--filename=/dev/sdm-%d", ix),
			})
		}
	}

	logf.Log.Info("Volumes created")

	// Create the fio Pod
	fioPodName := "fio-multi-vol"

	// Construct argument list for fio to run a single instance of fio,
	// with multiple jobs, one for each volume.
	var podArgs []string

	// 1) directives for all fio jobs
	podArgs = append(podArgs, []string{"---", "fio"}...)
	podArgs = append(podArgs, common.GetDefaultFioArguments()...)

	if volumeType == common.VolFileSystem {
		// for FS play safe use filesize which is 75% of volume size
		podArgs = append(podArgs, fmt.Sprintf("--size=%dm", (volSize*75)/100))
	}

	if fioLoops != 0 {
		podArgs = append(podArgs, fmt.Sprintf("--loops=%d", fioLoops))
	}

	// 2) per volume directives
	for _, v := range volFioArgs {
		podArgs = append(podArgs, v...)
	}

	// e2e-fio commandline is
	logf.Log.Info(fmt.Sprintf("commandline: %s", strings.Join(podArgs[1:], " ")))
	podArgs = append(podArgs, "&")
	logf.Log.Info("pod", "args", podArgs)

	container := k8stest.MakeFioContainer(fioPodName, podArgs)
	podBuilder := k8stest.NewPodBuilder().
		WithName(fioPodName).
		WithNamespace(common.NSDefault).
		WithContainer(container).
		WithVolumes(volumes)

	if len(volDevices) != 0 {
		podBuilder.WithVolumeDevices(volDevices)
	}

	if len(volMounts) != 0 {
		podBuilder.WithVolumeMounts(volMounts)
	}

	podObj, err := podBuilder.Build()
	Expect(err).ToNot(HaveOccurred(), "failed to build fio test pod object")

	pod, err := k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", fioPodName)
	Expect(pod).ToNot(BeNil(), "got nil pointer to test pod")

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

	logf.Log.Info("Waiting for run to complete", "timeout", timeout)

	elapsedTime := 0
	const sleepTime = 2
	var phase coreV1.PodPhase
	for elapsedTime = 0; elapsedTime < int(timeout.Seconds()); elapsedTime += sleepTime {
		time.Sleep(sleepTime * time.Second)
		phase, err = k8stest.CheckPodCompleted(fioPodName, common.NSDefault)
		Expect(err).To(BeNil(), "CheckPodComplete got error %s", err)
		if phase != coreV1.PodRunning {
			break
		}
	}
	Expect(phase == coreV1.PodSucceeded).To(BeTrue(), "fio pod phase is %s", phase)
	logf.Log.Info("fio completed", "duration (secs)", elapsedTime)

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volumes
	for _, volName := range volNames {
		err = k8stest.RmPVC(volName, scName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", volName)
	}

	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

var _ = Describe("Mayastor Volume IO test", func() {

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

	cfg := e2e_config.GetConfig().MultipleVolumesPodIO
	timeout, err := time.ParseDuration(cfg.Timeout)
	Expect(err).ToNot(HaveOccurred(), "Timeout configuration string format is invalid.")

	logf.Log.Info("MultipleVolumeIO test", "configuration", cfg)

	It("should verify mayastor can process IO on multiple filesystem volumes with multiple replicas mounted on a single pod with immediate binding", func() {
		multipleVolumeIOTest(cfg.MultipleReplicaCount, cfg.VolumeCount, common.ShareProtoNvmf,
			common.VolFileSystem, cfg.VolumeSizeMb, storageV1.VolumeBindingImmediate, false,
			timeout, cfg.FioLoops)
	})

	It("should verify mayastor can process IO on multiple filesystem volumes with multiple replicas mounted on a single pod with late binding", func() {
		multipleVolumeIOTest(cfg.MultipleReplicaCount, cfg.VolumeCount, common.ShareProtoNvmf,
			common.VolFileSystem, cfg.VolumeSizeMb, storageV1.VolumeBindingWaitForFirstConsumer, true,
			timeout, cfg.FioLoops)
	})

	It("should verify mayastor can process IO on multiple raw block volumes with multiple replicas mounted on a single pod with immediate binding", func() {
		multipleVolumeIOTest(cfg.MultipleReplicaCount, cfg.VolumeCount, common.ShareProtoNvmf,
			common.VolRawBlock, cfg.VolumeSizeMb, storageV1.VolumeBindingImmediate, false,
			timeout, cfg.FioLoops)
	})

	It("should verify mayastor can process IO on multiple raw block volumes with multiple replicas mounted on a single pod with late binding", func() {
		multipleVolumeIOTest(cfg.MultipleReplicaCount, cfg.VolumeCount, common.ShareProtoNvmf,
			common.VolRawBlock, cfg.VolumeSizeMb, storageV1.VolumeBindingWaitForFirstConsumer, true,
			timeout, cfg.FioLoops)
	})

})

var _ = BeforeSuite(func(done Done) {
	err := k8stest.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnv %v", err)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	err := k8stest.TeardownTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)

})
