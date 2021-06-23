package primitive_msp_state

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"time"

	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

const (
	defTimeoutSecs = 240
)

type mspStateConfig struct {
	protocol       common.ShareProto
	fsType         common.FileSystemType
	volType        common.VolumeType
	volBindingMode storageV1.VolumeBindingMode
	replicas       int
	scName         string
	pvcName        string
	pvcSize        int
	fioPodName     string
	uuid           string
	duration       time.Duration
	timeout        time.Duration
	thinkTime      time.Duration
}

func generateMspStateConfig(testName string, replicasCount int) *mspStateConfig {
	params := e2e_config.GetConfig().MaximumVolsIO
	fioDuration, err := time.ParseDuration(params.Duration)
	Expect(err).ToNot(HaveOccurred(), "Duration configuration string format is invalid.")
	fioCheckTimeout, err := time.ParseDuration(params.Timeout)
	Expect(err).ToNot(HaveOccurred(), "Timeout configuration string format is invalid.")
	fioThinkTime, err := time.ParseDuration(params.ThinkTime)
	Expect(err).ToNot(HaveOccurred(), "Think time configuration string format is invalid.")
	c := &mspStateConfig{
		protocol:       common.ShareProtoNvmf,
		volType:        common.VolFileSystem,
		fsType:         common.Ext4FsType,
		volBindingMode: storageV1.VolumeBindingImmediate,
		pvcSize:        params.VolMb,
		replicas:       replicasCount,
		scName:         testName + "-sc",
		pvcName:        testName + "-pvc",
		fioPodName:     testName + "-fio-pod",
		duration:       fioDuration,
		timeout:        fioCheckTimeout,
		thinkTime:      fioThinkTime,
	}

	return c
}
