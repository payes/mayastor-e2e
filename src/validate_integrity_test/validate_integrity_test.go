package validate_integrity_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"mayastor-e2e/common"
	client "mayastor-e2e/common/e2e-agent"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "120s"

var fioWriteParams = []string{
	"--name=benchtest",
	"--numjobs=1",
	"--direct=1",
	"--rw=randwrite",
	"--do_verify=0",
	"--ioengine=libaio",
	"--bs=4k",
	"--iodepth=16",
	"--verify=crc32",
}

var fioVerifyParams = []string{
	"--name=benchtest",
	"--numjobs=1",
	"--direct=1",
	"--rw=randread",
	"--ioengine=libaio",
	"--bs=4k",
	"--iodepth=16",
	"--verify=crc32",
	"--verify_fatal=1",
	"--verify_async=2",
}

func TestIntegrityTest(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Integrity test validation test", "validate_integrity_test")
}

func blankCentreBlock(nodeIP string, device string, sizeMB int) {
	blockSize := 4096
	sizeBlocks := (sizeMB * 1024 * 1024) / blockSize
	seekParam := fmt.Sprintf("seek=%d", sizeBlocks/2)
	blockSizeParam := fmt.Sprintf("bs=%d", blockSize)

	cmdArgs := []string{
		"dd",
		"if=/dev/zero",
		"of=/host" + device,
		"count=1",
		"oflag=direct",
		seekParam,
		blockSizeParam,
	}
	args := strings.Join(cmdArgs, " ")
	logf.Log.Info("Executing", "cmd", args)

	_, err := client.Exec(nodeIP, args)
	if err != nil {
		logf.Log.Info("Running agent failed", "error", err)
	}
	_, err = client.Exec(nodeIP, "sync")
	if err != nil {
		logf.Log.Info("Running agent failed", "error", err)
	}
}

func createFioPod(fioPodName string, volumeName string, volumeType common.VolumeType, verify bool) {

	var args = []string{
		"--",
	}
	args = append(args, fmt.Sprintf("--filename=%s", common.FioBlockFilename))

	if verify == true {
		args = append(args, fioVerifyParams...)
	} else {
		args = append(args, fioWriteParams...)
	}
	logf.Log.Info("fio", "arguments", args)

	// fio pod container
	podContainer := coreV1.Container{
		Name:            fioPodName,
		Image:           common.GetFioImage(),
		ImagePullPolicy: coreV1.PullAlways,
		Args:            args,
	}

	// volume claim details
	volume := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: volumeName,
			},
		},
	}

	podObj, err := k8stest.NewPodBuilder().
		WithName(fioPodName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(volumeType).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", fioPodName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", fioPodName)

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	logf.Log.Info("fio test pod is running.")
}

func validateCorruptionTest(corrupt bool) {
	protocol := common.ShareProtoNvmf
	volumeType := common.VolRawBlock

	params := e2e_config.GetConfig().ValidateIntegrityTest
	logf.Log.Info("Test", "parameters", params)

	nodeList, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	scName := "sc-validate-integrity-test"

	scObj, err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithProtocol(protocol).
		WithReplicas(params.Replicas).
		WithLocal(false).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating storage class definition %s", scName)

	err = k8stest.CreateSc(scObj)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	volumeName := "pvc-validate-integrity-test"

	// Create the volume
	uid := k8stest.MkPVC(params.VolMb, volumeName, scName, volumeType, common.NSDefault)
	logf.Log.Info("Volume", "uid", uid)

	// Create the fio Pod, The first time is just to write a verification pattern to the volume
	fioPodName := "fio-write-" + volumeName
	createFioPod(fioPodName, volumeName, volumeType, false)

	logf.Log.Info("Waiting for run to complete", "timeout", params.FioTimeout)
	tSecs := 0
	var phase coreV1.PodPhase
	for {
		if tSecs > params.FioTimeout {
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

	// (optionally) modify one block in the middle of each target devices
	if corrupt == true {
		for _, node := range nodeList {
			if node.MayastorNode {
				blankCentreBlock(node.IPAddress, params.Device, params.VolMb)
			}
		}
	}

	// Create a new fio Pod, This time just to verify the data.
	fioPodName = "fio-verify-" + volumeName
	createFioPod(fioPodName, volumeName, volumeType, true)

	logf.Log.Info("Waiting for run to complete", "timeout", params.FioTimeout)
	tSecs = 0

	for {
		if tSecs > params.FioTimeout {
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

	// if corruption is detected, the pod will be "failed"
	if corrupt {
		Expect(phase == coreV1.PodFailed).To(BeTrue(), "fio pod phase is %s", phase)
	} else {
		Expect(phase == coreV1.PodSucceeded).To(BeTrue(), "fio pod phase is %s", phase)
	}
	logf.Log.Info("fio completed", "duration", tSecs, "phase", string(phase))

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(volumeName, scName, common.NSDefault)

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

	It("should verify testing with no data corruption", func() {
		validateCorruptionTest(false)
	})

	It("should verify with data corruption", func() {
		validateCorruptionTest(true)
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	k8stest.TeardownTestEnv()
})
