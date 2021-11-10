package node_shutdown

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/platform"
	"mayastor-e2e/common/platform/types"

	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

const (
	defTimeoutSecs   = 800  // in seconds
	durationSecs     = 600  // in seconds
	volumeFileSizeMb = 250  // in Mb
	thinkTime        = 1000 // in milliseconds
)

type shutdownConfig struct {
	protocol             common.ShareProto
	fsType               common.FileSystemType
	volType              common.VolumeType
	volBindingMode       storageV1.VolumeBindingMode
	replicas             int
	scName               string
	pvcName              string
	podName              string
	pvcSize              int
	deployName           string
	numMayastorInstances int
	platform             types.Platform
}

type TestType int

func generateShutdownConfig(testName string) *shutdownConfig {
	c := &shutdownConfig{
		protocol:             common.ShareProtoNvmf,
		volType:              common.VolFileSystem,
		fsType:               common.Ext4FsType,
		volBindingMode:       storageV1.VolumeBindingImmediate,
		replicas:             3,
		pvcSize:              2048, // In Mb
		scName:               testName + "-sc",
		pvcName:              testName + "-pvc",
		deployName:           testName + "-deploy",
		podName:              testName + "-pod",
		numMayastorInstances: 3,
	}

	c.platform = platform.Create()
	Expect(c.platform).ToNot(BeNil())
	return c
}
