package primitive_fuzz_msv

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

var defTimeoutSecs = "180s"

const (
	timeoutSec   = 90
	sleepTimeSec = 2
)

type PrimitiveMsvFuzzConfig struct {
	Protocol    common.ShareProto
	FsType      common.FileSystemType
	Replicas    int
	VolumeCount int
	PvcScName   string
	ScName      string
	TestName    string
	PvcNames    []string
	FioPodNames []string
	PvcSize     int
	Uuid        []string
	CreateErrs  []error
	DeleteErrs  []error
	Iterations  int
	OptsList    []coreV1.PersistentVolumeClaim
}

func GeneratePrimitiveMsvFuzzConfig(testName string) *PrimitiveMsvFuzzConfig {
	params := e2e_config.GetConfig().PrimitiveMsvFuzz
	c := &PrimitiveMsvFuzzConfig{
		Protocol:    common.ShareProtoNvmf,
		FsType:      common.Ext4FsType,
		PvcSize:     params.VolMb,
		VolumeCount: params.VolumeCountPerPool,
		Replicas:    params.Replicas,
		Iterations:  params.Iterations,
		ScName:      testName + "-sc",
		PvcScName:   testName + "-sc",
		TestName:    testName,
	}
	return c
}
func (c *PrimitiveMsvFuzzConfig) GeneratePVCSpec() *PrimitiveMsvFuzzConfig {
	// List Pools by CRDs
	pools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	for _, pool := range pools {
		nodeName := pool.Spec.Node
		for ix := 0; ix < c.VolumeCount; ix++ {
			//generate pvc name
			pvcName := fmt.Sprintf("%s-%s-pvc-%d", c.TestName, nodeName, ix)
			c.PvcNames = append(c.PvcNames, pvcName)
			//generate fio pod name
			fioPodName := fmt.Sprintf("%s-%s-fio-%d", c.TestName, nodeName, ix)
			c.FioPodNames = append(c.FioPodNames, fioPodName)
			//pvc size
			volSizeMbStr := fmt.Sprintf("%dMi", c.PvcSize)
			// VolumeMode: Filesystem
			var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem
			Opts := coreV1.PersistentVolumeClaim{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      pvcName,
					Namespace: common.NSDefault,
				},
				Spec: coreV1.PersistentVolumeClaimSpec{
					StorageClassName: &c.PvcScName,
					AccessModes:      []coreV1.PersistentVolumeAccessMode{coreV1.ReadWriteOnce},
					Resources: coreV1.ResourceRequirements{
						Requests: coreV1.ResourceList{
							coreV1.ResourceStorage: resource.MustParse(volSizeMbStr),
						},
					},
					VolumeMode: &fileSystemVolumeMode,
				},
			}
			c.OptsList = append(c.OptsList, Opts)
		}
	}
	c.CreateErrs = make([]error, c.VolumeCount*len(pools))
	c.DeleteErrs = make([]error, c.VolumeCount*len(pools))
	c.Uuid = make([]string, c.VolumeCount*len(pools))
	return c
}
