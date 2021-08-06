package maximum_vols_io

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"time"

	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

const (
	defTimeoutSecs = 300
)

type maxVolConfig struct {
	protocol       common.ShareProto
	fsType         common.FileSystemType
	volType        common.VolumeType
	volBindingMode storageV1.VolumeBindingMode
	replicas       int
	volCountPerPod int
	podCount       int
	scName         string
	pvcNames       []string
	pvcSize        int
	fioPodNames    []string
	uuid           []string
	duration       time.Duration
	timeout        time.Duration
	thinkTime      time.Duration
}

func generateMaxVolConfig(testName string, replicasCount int) *maxVolConfig {
	params := e2e_config.GetConfig().MaximumVolsIO
	fioDuration, err := time.ParseDuration(params.Duration)
	Expect(err).ToNot(HaveOccurred(), "Duration configuration string format is invalid.")
	fioCheckTimeout, err := time.ParseDuration(params.Timeout)
	Expect(err).ToNot(HaveOccurred(), "Timeout configuration string format is invalid.")
	fioThinkTime, err := time.ParseDuration(params.ThinkTime)
	Expect(err).ToNot(HaveOccurred(), "Think time configuration string format is invalid.")
	c := &maxVolConfig{
		protocol:       common.ShareProtoNvmf,
		volType:        common.VolFileSystem,
		fsType:         common.Ext4FsType,
		volBindingMode: storageV1.VolumeBindingImmediate,
		pvcSize:        params.VolMb,
		volCountPerPod: params.VolumeCountPerPod,
		podCount:       params.PodCount,
		replicas:       replicasCount,
		scName:         testName + "-sc",
		duration:       fioDuration,
		timeout:        fioCheckTimeout,
		thinkTime:      fioThinkTime,
	}

	for ix := 0; ix < c.podCount; ix++ {
		//generate fio pod name
		fioPodName := fmt.Sprintf("%s-fio-%d", testName, ix)
		for jx := 0; jx < c.volCountPerPod; jx++ {
			//generate pvc name
			pvcName := fmt.Sprintf("%s-pvc-%d", fioPodName, jx)
			c.pvcNames = append(c.pvcNames, pvcName)
		}
		c.fioPodNames = append(c.fioPodNames, fioPodName)
	}

	return c
}
