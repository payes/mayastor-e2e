package xfstests

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	// . "github.com/onsi/gomega"
)

const (
	podCompletionTimeout = 900
	defTimeoutSecs       = "90s"
	sleepTime            = 2
)

type xfsTestConfig struct {
	protocol          common.ShareProto
	fsType            common.FileSystemType
	replicaCount      int
	scName            string
	pvcNames          []string
	xfstestPodName    string
	pvcSize           int
	uuids             []string
	fileSystemType    string
	numberOfOperation int
	devicePaths       []string
}

func generateXFSTestsConfig(testName string) *xfsTestConfig {
	params := e2e_config.GetConfig().XFSTests
	c := &xfsTestConfig{
		protocol:          common.ShareProtoNvmf,
		fsType:            common.Ext4FsType,
		pvcSize:           params.VolMb,
		replicaCount:      params.Replicas,
		fileSystemType:    params.FileSystemType,
		numberOfOperation: params.NumberOfOperation,
		devicePaths:       common.XFSTestsBlockFilenames,
		scName:            testName + "-sc",
		pvcNames:          []string{testName + "-pvc-test", testName + "-pvc-scratch"},
		xfstestPodName:    testName,
	}
	return c
}
