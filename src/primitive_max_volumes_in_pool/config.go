package primitive_max_volumes_in_pool

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// . "github.com/onsi/gomega"
)

var defTimeoutSecs = "90s"

const (
	timeoutSec   = 90
	sleepTimeSec = 2
)

type primitiveMaxVolConfig struct {
	protocol    common.ShareProto
	fsType      common.FileSystemType
	replicas    int
	volumeCount int
	scName      string
	pvcNames    []string
	pvcSize     int
	uuid        []string
	createErrs  []error
	deleteErrs  []error
	optsList    []coreV1.PersistentVolumeClaim
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
		createErrs:  make([]error, params.VolumeCount),
		deleteErrs:  make([]error, params.VolumeCount),
		uuid:        make([]string, params.VolumeCount),
	}
	for ix := 0; ix < c.volumeCount; ix++ {
		//generate pvc name
		pvcName := fmt.Sprintf("%s-pvc-%d", testName, ix)
		volSizeMbStr := fmt.Sprintf("%dMi", c.pvcSize)
		// VolumeMode: Filesystem
		var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem
		opts := coreV1.PersistentVolumeClaim{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      pvcName,
				Namespace: common.NSDefault,
			},
			Spec: coreV1.PersistentVolumeClaimSpec{
				StorageClassName: &c.scName,
				AccessModes:      []coreV1.PersistentVolumeAccessMode{coreV1.ReadWriteOnce},
				Resources: coreV1.ResourceRequirements{
					Requests: coreV1.ResourceList{
						coreV1.ResourceStorage: resource.MustParse(volSizeMbStr),
					},
				},
				VolumeMode: &fileSystemVolumeMode,
			},
		}

		c.pvcNames = append(c.pvcNames, pvcName)
		c.optsList = append(c.optsList, opts)
	}

	return c
}
