package fsx_ext4_stress

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	// . "github.com/onsi/gomega"
)

const (
	podCompletionTimeout = 900
	sleepTime            = 2
	defTimeoutSecs       = "90s"
	patchSleepTime       = 10
	patchTimeout         = 240
)

type fsxExt4StressConfig struct {
	protocol          common.ShareProto
	fsType            common.FileSystemType
	replicaCount      int
	scName            string
	pvcName           string
	fsxPodName        string
	pvcSize           int
	uuid              string
	nexusNodeIP       string
	nexusRep          string
	nexusUuid         string
	replicaIPs        []string
	fileSystemType    string
	numberOfOperation int
	devicePath        string
}

func generateFsxExt4StressConfig(testName string) *fsxExt4StressConfig {
	params := e2e_config.GetConfig().FsxExt4Stress
	c := &fsxExt4StressConfig{
		protocol:          common.ShareProtoNvmf,
		fsType:            common.Ext4FsType,
		pvcSize:           params.VolMb,
		replicaCount:      params.Replicas,
		fileSystemType:    params.FileSystemType,
		numberOfOperation: params.NumberOfOperation,
		devicePath:        common.FsxBlockFileName,
		scName:            testName + "-sc",
		pvcName:           testName + "-pvc",
		fsxPodName:        testName + "-fsx",
	}
	return c
}
