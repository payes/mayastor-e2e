package primitive_max_volumes_in_pool

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	// . "github.com/onsi/gomega"
)

// const (
// 	defTimeoutSecs = 240
// )

type primitiveMaxVolConfig struct {
	protocol    common.ShareProto
	fsType      common.FileSystemType
	replicas    int
	volumeCount int
	scName      string
	pvcNames    []string
	pvcSize     int
	uuid        []string
}

func generatePrimitiveMaxVolConfig(testName string, replicasCount int) *primitiveMaxVolConfig {
	params := e2e_config.GetConfig().PrimitiveMaxVolsInPool
	c := &primitiveMaxVolConfig{
		protocol:    common.ShareProtoNvmf,
		fsType:      common.Ext4FsType,
		pvcSize:     params.VolMb,
		volumeCount: params.VolumeCount,
		replicas:    replicasCount,
		scName:      testName + "-sc",
	}

	return c
}
