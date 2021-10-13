package rc_reconciliation

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/platform"
	"mayastor-e2e/common/platform/types"

	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

type Config struct {
	protocol             common.ShareProto
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

func generateConfig(testName string) *Config {
	c := &Config{
		protocol:             common.ShareProtoNvmf,
		volType:              common.VolRawBlock,
		volBindingMode:       storageV1.VolumeBindingImmediate,
		replicas:             2,
		scName:               testName + "-sc",
		pvcName:              testName + "-pvc",
		podName:              testName + "-pod",
		pvcSize:              2048, // In Mb
		deployName:           testName + "-deploy",
		numMayastorInstances: 3,
	}

	c.platform = platform.Create()
	Expect(c.platform).ToNot(BeNil())
	return c
}
