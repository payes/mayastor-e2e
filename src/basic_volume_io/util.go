package basic_volume_io

import (
	"fmt"
	"strings"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "120s"

func BasicVolumeIOTest(protocol common.ShareProto, volumeType common.VolumeType, mode storageV1.VolumeBindingMode) {
	params := e2e_config.GetConfig().BasicVolumeIO
	log.Log.Info("Test", "parameters", params)
	scName := strings.ToLower(fmt.Sprintf("basic-vol-io-repl-%d-%s-%s-%s", common.DefaultReplicaCount, string(protocol), volumeType, mode))
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithReplicas(common.DefaultReplicaCount).
		WithProtocol(protocol).
		WithNamespace(common.NSDefault).
		WithVolumeBindingMode(mode).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "failed to create storage class %s", scName)

	volName := strings.ToLower(fmt.Sprintf("basic-vol-io-repl-%d-%s-%s-%s", common.DefaultReplicaCount, string(protocol), volumeType, mode))

	// Create the volume
	uid := k8stest.MkPVC(params.VolSizeMb, volName, scName, volumeType, common.NSDefault)
	log.Log.Info("Volume", "uid", uid)

	// Create the fio Pod
	fioPodName := "fio-" + volName

	var args = []string{
		"--",
	}
	switch volumeType {
	case common.VolFileSystem:
		args = append(args, fmt.Sprintf("--filename=%s", common.FioFsFilename))
		args = append(args, fmt.Sprintf("--size=%dm", params.FsVolSizeMb))
	case common.VolRawBlock:
		args = append(args, fmt.Sprintf("--filename=%s", common.FioBlockFilename))
	}
	args = append(args, common.GetFioArgs()...)
	log.Log.Info("fio", "arguments", args)

	// fio pod container
	podContainer := k8stest.MakeFioContainer(fioPodName, args)

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
		WithName(fioPodName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(volumeType).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", fioPodName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

	// Create fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", fioPodName)

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	log.Log.Info("fio test pod is running.")

	msvc_err := k8stest.MsvConsistencyCheck(uid)
	Expect(msvc_err).ToNot(HaveOccurred(), "%v", msvc_err)

	log.Log.Info("Waiting for run to complete", "timeout", params.FioTimeout)
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
	log.Log.Info("fio completed", "duration", tSecs)

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}
