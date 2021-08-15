package primitive_device_retirement

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/platform"
	"mayastor-e2e/common/platform/types"

	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

type primitiveDeviceRetirementConfig struct {
	protocol             common.ShareProto
	fsType               common.FileSystemType
	volType              common.VolumeType
	volBindingMode       storageV1.VolumeBindingMode
	replicas             int
	replicaIPs           []string
	scName               string
	pvcName              string
	podName              string
	pvcSize              int
	deployName           string
	numMayastorInstances int
	platform             types.Platform
}

type TestType int

func generatePrimitiveDeviceRetirementConfig(testName string) *primitiveDeviceRetirementConfig {
	c := &primitiveDeviceRetirementConfig{
		protocol:             common.ShareProtoNvmf,
		fsType:               common.Ext4FsType,
		volType:              common.VolRawBlock,
		volBindingMode:       storageV1.VolumeBindingImmediate,
		replicas:             3,
		scName:               testName + "-sc",
		pvcName:              testName + "-pvc",
		podName:              testName + "-pod",
		pvcSize:              1024, // In Mb
		deployName:           testName + "-deploy",
		numMayastorInstances: 3,
	}

	c.platform = platform.Create()
	Expect(c.platform).ToNot(BeNil())
	return c
}
