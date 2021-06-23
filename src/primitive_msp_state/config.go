package primitive_msp_state

import (
	"mayastor-e2e/common"

	// . "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

const (
	defTimeoutSecs = 240
	fioDuration    = 30
	fioTimeout     = 100
	fioThinkTime   = 1000
	pvcSize        = 1024
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
	poolNames      []string
	msvSize        int64
	duration       int64
	timeout        int64
	thinkTime      int64
}

func generateMspStateConfig(testName string, replicasCount int) *mspStateConfig {
	c := &mspStateConfig{
		protocol:       common.ShareProtoNvmf,
		volType:        common.VolFileSystem,
		fsType:         common.Ext4FsType,
		volBindingMode: storageV1.VolumeBindingImmediate,
		pvcSize:        pvcSize,
		replicas:       replicasCount,
		scName:         testName + "-sc",
		pvcName:        testName + "-pvc",
		fioPodName:     testName + "-fio-pod",
		duration:       fioDuration,
		timeout:        fioTimeout,
		thinkTime:      fioThinkTime,
	}

	return c
}
