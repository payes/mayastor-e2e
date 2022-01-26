package last_replica

import (
	"fmt"
	"strings"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"

	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

type testContext struct {
	poolDev        string
	protocol       common.ShareProto
	volumeType     common.VolumeType
	volBindingMode storageV1.VolumeBindingMode
	replicaCount   int
	scName         string
	pvcName        string
	podName        string
	pvcSize        int
	fioRuntime     int
	fioTimeout     int
	blipSeconds    int
	fioArgs        []string
	volUid         string
}

func makeTestContext(name string) *testContext {
	ctx := &testContext{
		poolDev:        e2e_config.GetConfig().PoolDevice,
		protocol:       common.ShareProtoNvmf,
		volumeType:     common.VolRawBlock,
		volBindingMode: storageV1.VolumeBindingImmediate,
		replicaCount:   1,
		scName:         name + "-sc",
		pvcName:        name + "-pvc",
		podName:        name + "-pod",
		pvcSize:        2048, // In MB
		fioRuntime:     30,
		fioTimeout:     60,
		blipSeconds:    0,
	}
	Expect(strings.HasPrefix(ctx.poolDev, "/dev/")).To(BeTrue(), "unexpected pool device")
	ctx.poolDev = e2e_config.GetConfig().PoolDevice[5:]
	ctx.fioArgs = []string{
		"--",
		"--time_based",
		fmt.Sprintf("--runtime=%d", ctx.fioRuntime),
	}
	switch ctx.volumeType {
	case common.VolFileSystem:
		ctx.fioArgs = append(ctx.fioArgs, fmt.Sprintf("--filename=%s", common.FioFsFilename))
		ctx.fioArgs = append(ctx.fioArgs, fmt.Sprintf("--size=%dm", ctx.pvcSize))
	case common.VolRawBlock:
		ctx.fioArgs = append(ctx.fioArgs, fmt.Sprintf("--filename=%s", common.FioBlockFilename))
	}
	ctx.fioArgs = append(ctx.fioArgs, common.GetFioArgs()...)
	return ctx
}
