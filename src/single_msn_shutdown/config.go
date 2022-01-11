package single_msn_shutdown

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/platform"
	"mayastor-e2e/common/platform/types"

	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

const (
	defTimeoutSecs   = 300 // in seconds
	defWaitTimeout   = "600s"
	volumeFileSizeMb = 100
	thinkTime        = 100000 // in milliseconds
	thinkTimeBlocks  = 10
	fioTimeoutSecs   = 300
)

type appConfig struct {
	pvcName              string
	uuid                 string
	deployName           string
	scName               string
	replicas             int
	pvcSize              int
	nodeName             string
	podName              string
	taskCompletionStatus string
	protocol             common.ShareProto
	fsType               common.FileSystemType
	volType              common.VolumeType
	volBindingMode       storageV1.VolumeBindingMode
}

type shutdownConfig struct {
	numMayastorInstances int
	platform             types.Platform
	config               []*appConfig
}

type TestType int

var poweredOffNode string

func generateConfig(testName string) *shutdownConfig {
	c := &shutdownConfig{
		numMayastorInstances: 3,
	}
	c.platform = platform.Create()
	Expect(c.platform).ToNot(BeNil())

	nodeLocs, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), err)
	for _, node := range nodeLocs {
		if !node.MayastorNode {
			continue
		}
		config := &appConfig{
			protocol:       common.ShareProtoNvmf,
			volType:        common.VolFileSystem,
			fsType:         common.Ext4FsType,
			volBindingMode: storageV1.VolumeBindingImmediate,
			replicas:       3,
			pvcSize:        1024, // In Mb
			scName:         testName + "-sc-" + node.NodeName,
			pvcName:        testName + "-pvc-" + node.NodeName,
			deployName:     testName + "-deploy-" + node.NodeName,
			nodeName:       node.NodeName,
		}
		c.config = append(c.config, config)
	}
	return c
}
