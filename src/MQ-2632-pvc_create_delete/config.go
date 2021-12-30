package pvc_create_delete

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"

	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeoutSec   = 180
	sleepTimeSec = 2
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
	pvcSize        int
	optsList       []coreV1.PersistentVolumeClaim
	volumeCount    int
}

func generatePvcCreateDeleteConfig(testName string, volCount int) *pvcCreateDeleteConfig {
	params := e2e_config.GetConfig().PvcCreateDelete
	c := &pvcCreateDeleteConfig{
		protocol:       common.ShareProtoNvmf,
		volType:        common.VolFileSystem,
		fsType:         common.Ext4FsType,
		volBindingMode: storageV1.VolumeBindingImmediate,
		pvcSize:        params.VolSize,
		iterations:     params.Iterations,
		replicas:       params.Replicas,
		scName:         testName + "-sc",
	}
	c.volumeCount = volCount * params.VolumeMultiplier

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
