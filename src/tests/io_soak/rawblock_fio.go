package io_soak

import (
	"fmt"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
)

// IO soak raw block fio  job

type FioRawBlockSoakJob struct {
	volName  string
	scName   string
	podName  string
	id       int
	duration int
	volUUID  string
}

func (job FioRawBlockSoakJob) makeVolume() {
	var err error
	job.volUUID, err = k8stest.MkPVC(common.DefaultVolumeSizeMb, job.volName, job.scName, common.VolRawBlock, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", job.volName)
}

func (job FioRawBlockSoakJob) removeVolume() {
	err := k8stest.RmPVC(job.volName, job.scName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", job.volName)
}

func (job FioRawBlockSoakJob) makeTestPod(selector map[string]string) (*coreV1.Pod, error) {
	var z64 int64 = 0
	var vTrue bool = true

	sc := coreV1.SecurityContext{
		Privileged:               &vTrue,
		RunAsUser:                &z64,
		AllowPrivilegeEscalation: &vTrue,
	}

	pod := k8stest.CreateFioPodDef(job.podName, job.volName, common.VolRawBlock, common.NSDefault)
	pod.Spec.NodeSelector = selector

	e2eCfg := e2e_config.GetConfig()

	args := []string{
		"--",
		fmt.Sprintf("--startdelay=%d", e2eCfg.IOSoakTest.FioStartDelay),
		"--time_based",
		fmt.Sprintf("--runtime=%d", job.duration),
		fmt.Sprintf("--filename=%s", common.FioBlockFilename),
		fmt.Sprintf("--thinktime=%d", GetThinkTime(job.id)),
		fmt.Sprintf("--thinktime_blocks=%d", GetThinkTimeBlocks(job.id)),
	}
	args = append(args, GetIOSoakFioArgs()...)
	pod.Spec.Containers[0].Args = args
	pod.Spec.Containers[0].SecurityContext = &sc

	pod, err := k8stest.CreatePod(pod, common.NSDefault)
	return pod, err
}

func (job FioRawBlockSoakJob) removeTestPod() error {
	return k8stest.DeletePod(job.podName, common.NSDefault)
}

func (job FioRawBlockSoakJob) getPodName() string {
	return job.podName
}

func (job FioRawBlockSoakJob) describe() string {
	return fmt.Sprintf("pod: %s, vol: %s, volUUID: %s", job.podName, job.podName, job.volUUID)
}

func MakeFioRawBlockJob(scName string, id int, duration time.Duration) FioRawBlockSoakJob {
	nm := fmt.Sprintf("fio-rawblock-%s-%d", scName, id)
	return FioRawBlockSoakJob{
		volName:  nm,
		scName:   scName,
		podName:  nm,
		id:       id,
		duration: int(duration.Seconds()),
	}
}
