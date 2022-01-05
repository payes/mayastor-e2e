package pvc_create_delete

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"

	storageV1 "k8s.io/api/storage/v1"
)

const (
	timeoutSec     = 180
	sleepTimeSec   = 2
	defTimeoutSecs = 1800
)

type pvcCreateDeleteConfig struct {
	protocol       common.ShareProto
	fsType         common.FileSystemType
	volType        common.VolumeType
	volBindingMode storageV1.VolumeBindingMode
	replicas       int
	iterations     int
	scName         string
	pvcNames       []string
	pvcSizeMB      int
	volumeCount    int
	delayTime      int
}

func generatePvcCreateDeleteConfig(testName string, volCount int) *pvcCreateDeleteConfig {
	params := e2e_config.GetConfig().PvcCreateDelete
	c := &pvcCreateDeleteConfig{
		protocol:       common.ShareProtoNvmf,
		volType:        common.VolFileSystem,
		fsType:         common.Ext4FsType,
		volBindingMode: storageV1.VolumeBindingImmediate,
		pvcSizeMB:      params.VolSize,
		iterations:     params.Iterations,
		replicas:       params.Replicas,
		scName:         testName + "-sc",
		delayTime:      params.DelayTime,
	}
	c.volumeCount = volCount * params.VolumeMultiplier

	for ix := 0; ix < c.volumeCount; ix++ {
		//generate pvc name
		pvcName := fmt.Sprintf("%s-pvc-%d", testName, ix)
		c.pvcNames = append(c.pvcNames, pvcName)
	}
	return c
}
