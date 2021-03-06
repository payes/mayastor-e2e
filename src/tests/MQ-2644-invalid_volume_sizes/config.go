package MQ_2644_invalid_volume_sizes

import (
	storageV1 "k8s.io/api/storage/v1"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
)

const (
	defTimeoutSecs = 3600
)

type pvcConfig struct {
	protocol       common.ShareProto
	fsType         common.FileSystemType
	volType        common.VolumeType
	volBindingMode storageV1.VolumeBindingMode
	replicas       int
	iterations     int
	scName         string
	pvcName        string
	pvcSizeMB      int
	volumeCount    int
	delayTime      int
	testName       string
}

func generatePvc(testName string, replicas int, volSizeMB int) *pvcConfig {
	params := e2e_config.GetConfig().PvcCreateDelete
	c := &pvcConfig{
		protocol:       common.ShareProtoNvmf,
		volType:        common.VolFileSystem,
		fsType:         common.Ext4FsType,
		volBindingMode: storageV1.VolumeBindingImmediate,
		pvcSizeMB:      volSizeMB,
		iterations:     params.Iterations,
		replicas:       replicas,
		scName:         testName + "-sc",
		pvcName:        testName + "-pvc",
		delayTime:      3,
		volumeCount:    1,
		testName:       testName,
	}
	return c
}
