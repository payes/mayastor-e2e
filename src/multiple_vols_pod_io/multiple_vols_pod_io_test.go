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

func multipleVolumeIOTest(replicas int, volumeCount int, protocol common.ShareProto, volumeType common.VolumeType, binding storageV1.VolumeBindingMode, duration time.Duration, timeout time.Duration) {
	logf.Log.Info("MultipleVolumeIOTest", "replicas", replicas, "volumeCount", volumeCount, "protocol", protocol, "volumeType", volumeType, "binding", binding)
	scName := strings.ToLower(fmt.Sprintf("msv-repl-%d-%s-%s-%s", replicas, string(protocol), volumeType, binding))
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithReplicas(replicas).
		WithProtocol(protocol).
		WithNamespace(common.NSDefault).
		WithVolumeBindingMode(binding).
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
		uid := k8stest.MkPVC(common.DefaultVolumeSizeMb, volName, scName, volumeType, common.NSDefault)
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

		if volumeType == common.VolFileSystem {
			mount := coreV1.VolumeMount{
				Name:      fmt.Sprintf("ms-volume-%d", ix),
				MountPath: fmt.Sprintf("/volume-%d", ix),
			}
			volMounts = append(volMounts, mount)
			volFioArgs = append(volFioArgs, []string{
				fmt.Sprintf("--filename=/volume-%d/fio-test-file", ix),
				fmt.Sprintf("--size=%dm", common.DefaultFioSizeMb),
			})
		} else {
			device := coreV1.VolumeDevice{
				Name:       fmt.Sprintf("ms-volume-%d", ix),
				DevicePath: fmt.Sprintf("/dev/sdm-%d", ix),
			}
			volDevices = append(volDevices, device)
			volFioArgs = append(volFioArgs, []string{
				fmt.Sprintf("--filename=/dev/sdm-%d", ix),
			})
		}
	}

	logf.Log.Info("Volumes created")

	// Create the fio Pod
	fioPodName := "fio-multi-vol"

	pod := k8stest.CreateFioPodDef(fioPodName, "aa", volumeType, common.NSDefault)
	pod.Spec.Volumes = volumes
	switch volumeType {
	case common.VolFileSystem:
		pod.Spec.Containers[0].VolumeMounts = volMounts
	case common.VolRawBlock:
		pod.Spec.Containers[0].VolumeDevices = volDevices
	}

	// Construct argument list for fio to run a single instance of fio,
	// with multiple jobs, one for each volume.
	var podArgs []string

	// 1) directives for all fio jobs
	podArgs = append(podArgs, "--")
	podArgs = append(podArgs, common.GetDefaultFioArguments()...)
	podArgs = append(podArgs, []string{
		"--time_based",
		fmt.Sprintf("--runtime=%d", int(duration.Seconds())),
	}...,
	)

	// 2) per volume directives (filename, size, and testname)
	for ix, v := range volFioArgs {
		podArgs = append(podArgs, v...)
		podArgs = append(podArgs, fmt.Sprintf("--name=benchtest-%d", ix))
	}
	podArgs = append(podArgs, "&")

	logf.Log.Info("fio", "arguments", podArgs)
	pod.Spec.Containers[0].Args = podArgs

	pod, err = k8stest.CreatePod(pod, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())
	Expect(pod).ToNot(BeNil())

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

	logf.Log.Info("Waiting for run to complete", "duration", duration, "timeout", timeout)
	tSecs := 0
	var phase coreV1.PodPhase
	for {
		if tSecs > int(timeout.Seconds()) {
			break
		}
		time.Sleep(1 * time.Second)
		tSecs += 1
		phase, err = k8stest.CheckPodCompleted(fioPodName, common.NSDefault)
		Expect(err).To(BeNil(), "CheckPodComplete got error %s", err)
		if phase != coreV1.PodRunning {
			break
		}
	}
	Expect(phase == coreV1.PodSucceeded).To(BeTrue(), "fio pod phase is %s", phase)
	logf.Log.Info("fio completed", "duration", tSecs)

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volumes
	for _, volName := range volNames {
		k8stest.RmPVC(volName, scName, common.NSDefault)
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
	duration, err := time.ParseDuration(cfg.Duration)
	Expect(err).ToNot(HaveOccurred(), "Duration configuration string format is invalid.")
	timeout, err := time.ParseDuration(cfg.Timeout)
	Expect(err).ToNot(HaveOccurred(), "Timeout configuration string format is invalid.")

	logf.Log.Info("MultipleVolumeIO test", "configuration", cfg)

	It("should verify mayastor can process IO on multiple filesystem volumes with multiple replicas mounted on a single pod with immediate binding", func() {
		multipleVolumeIOTest(cfg.MultipleReplicaCount, cfg.VolumeCount, common.ShareProtoNvmf, common.VolFileSystem, storageV1.VolumeBindingImmediate, duration, timeout)
	})

	It("should verify mayastor can process IO on multiple raw block volumes with multiple replicas mounted on a single pod with immediate binding", func() {
		multipleVolumeIOTest(cfg.MultipleReplicaCount, cfg.VolumeCount, common.ShareProtoNvmf, common.VolRawBlock, storageV1.VolumeBindingImmediate, duration, timeout)
	})

	It("should verify mayastor can process IO on multiple filesystem volumes with multiple replicas mounted on a single pod with late binding", func() {
		multipleVolumeIOTest(cfg.MultipleReplicaCount, cfg.VolumeCount, common.ShareProtoNvmf, common.VolFileSystem, storageV1.VolumeBindingWaitForFirstConsumer, duration, timeout)
	})

	It("should verify mayastor can process IO on multiple raw block volumes with multiple replicas mounted on a single pod with late binding", func() {
		multipleVolumeIOTest(cfg.MultipleReplicaCount, cfg.VolumeCount, common.ShareProtoNvmf, common.VolRawBlock, storageV1.VolumeBindingWaitForFirstConsumer, duration, timeout)
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
