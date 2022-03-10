package etcd_inaccessibility

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"strings"

	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

type inaccessibleEtcdTestConfig struct {
	protocol       common.ShareProto
	volType        common.VolumeType
	volBindingMode storageV1.VolumeBindingMode
	replicas       int
	scName         string
	pvcName        string
	podName        string
	pvcSize        int
}

type TestType int

var (
	defTimeoutSecs       = "600s"
	detachedDeviceAtNode string
	poolDevice           string
)

func generateInaccessibleEtcdTestConfig(testName string) *inaccessibleEtcdTestConfig {
	c := &inaccessibleEtcdTestConfig{
		protocol:       common.ShareProtoNvmf,
		volType:        common.VolRawBlock,
		volBindingMode: storageV1.VolumeBindingImmediate,
		replicas:       2,
		scName:         testName + "-sc",
		pvcName:        testName + "-pvc",
		podName:        testName + "-pod",
		pvcSize:        2048, // In MB
	}
	e2eCfg := e2e_config.GetConfig()
	poolDevice = e2eCfg.PoolDevice
	Expect(strings.HasPrefix(poolDevice, "/dev/")).To(BeTrue(), "unexpected pool spec %s", poolDevice)
	poolDevice = poolDevice[5:]

	return c
}
