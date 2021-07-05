package primitive_max_volumes_in_pool

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"

	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func generatePrimitiveMaxVolConfig(testName string) *primitiveMaxVolConfig {
	params := e2e_config.GetConfig().PrimitiveMaxVolsInPool
	c := &primitiveMaxVolConfig{
		protocol:    common.ShareProtoNvmf,
		fsType:      common.Ext4FsType,
		pvcSize:     params.VolMb,
		volumeCount: params.VolumeCountPerPool,
		replicas:    params.Replicas,
		scName:      testName + "-sc",
	}
	// List Pools by CRDs
	pools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	for _, pool := range pools {
		nodeName := pool.Spec.Node
		for ix := 0; ix < c.volumeCount; ix++ {
			//generate pvc name
			pvcName := fmt.Sprintf("%s-%s-pvc-%d", testName, nodeName, ix)
			c.pvcNames = append(c.pvcNames, pvcName)

			//pvc size
			volSizeMbStr := fmt.Sprintf("%dMi", c.pvcSize)
			// VolumeMode: Filesystem
			var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem
			//node selector
			nodeSelector := map[string]string{
				"kubernetes.io/hostname": nodeName,
			}
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
					Selector: &metaV1.LabelSelector{
						MatchLabels: nodeSelector,
					},
				},
			}
			c.optsList = append(c.optsList, opts)
		}
	}
	c.createErrs = make([]error, params.VolumeCountPerPool*len(pools))
	c.deleteErrs = make([]error, params.VolumeCountPerPool*len(pools))
	c.uuid = make([]string, params.VolumeCountPerPool*len(pools))
	return c
}
