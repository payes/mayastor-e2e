package basic_volume_io

import (
	"fmt"
	"github.com/onsi/gomega"
	v12 "k8s.io/api/core/v1"
	"k8s.io/api/storage/v1"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"time"
)

var defTimeoutSecs = "120s"

func BasicVolumeIOTest(protocol common.ShareProto, volumeType common.VolumeType, mode v1.VolumeBindingMode) {
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
	gomega.Expect(err).ToNot(gomega.HaveOccurred(), "failed to create storage class %s", scName)

	volName := strings.ToLower(fmt.Sprintf("basic-vol-io-repl-%d-%s-%s-%s", common.DefaultReplicaCount, string(protocol), volumeType, mode))

	// Create the volume
	uid := k8stest.MkPVC(params.VolSizeMb, volName, scName, volumeType, common.NSDefault)
	log.Log.Info("Volume", "uid", uid)

	// Create the fio Pod
	fioPodName := "fio-" + volName
	pod := k8stest.CreateFioPodDef(fioPodName, volName, volumeType, common.NSDefault)
	gomega.Expect(pod).ToNot(gomega.BeNil())

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
	pod.Spec.Containers[0].Args = args

	pod, err = k8stest.CreatePod(pod, common.NSDefault)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	gomega.Expect(pod).ToNot(gomega.BeNil())

	// Wait for the fio Pod to transition to running
	gomega.Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(gomega.Equal(true))
	log.Log.Info("fio test pod is running.")

	log.Log.Info("Waiting for run to complete", "timeout", params.FioTimeout)
	tSecs := 0
	var phase v12.PodPhase
	for {
		if tSecs > params.FioTimeout {
			break
		}
		time.Sleep(1 * time.Second)
		tSecs += 1
		phase, err = k8stest.CheckPodCompleted(fioPodName, common.NSDefault)
		gomega.Expect(err).To(gomega.BeNil(), "CheckPodComplete got error %s", err)
		if phase != v12.PodRunning {
			break
		}
	}
	gomega.Expect(phase == v12.PodSucceeded).To(gomega.BeTrue(), "fio pod phase is %s", phase)
	log.Log.Info("fio completed", "duration", tSecs)

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	err = k8stest.RmStorageClass(scName)
	gomega.Expect(err).ToNot(gomega.HaveOccurred(), "Deleting storage class %s", scName)
}
