package primitive_fault_injection

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"time"

	. "github.com/onsi/gomega"
)

const (
	sleepTime      = 2
	defTimeoutSecs = "90s"
	patchSleepTime = 10
	patchTimeout   = 240
)

type primitiveFaultInjectionConfig struct {
	protocol     common.ShareProto
	fsType       common.FileSystemType
	replicaCount int
	scName       string
	pvcName      string
	fioPodName   string
	pvcSize      int
	uuid         string
	nexusNodeIP  string
	nexusRep     string
	replicaIPs   []string
	duration     time.Duration
	thinkTime    time.Duration
	timeout      time.Duration
}

func generatePrimitiveFaultInjectionConfig(testName string) *primitiveFaultInjectionConfig {
	params := e2e_config.GetConfig().PrimitiveFaultInjection
	fioDuration, err := time.ParseDuration(params.Duration)
	Expect(err).ToNot(HaveOccurred(), "Duration configuration string format is invalid.")
	fioThinkTime, err := time.ParseDuration(params.ThinkTime)
	Expect(err).ToNot(HaveOccurred(), "Think time configuration string format is invalid.")
	fioTimeout, err := time.ParseDuration(params.Timeout)
	Expect(err).ToNot(HaveOccurred(), "Timeout configuration string format is invalid.")
	c := &primitiveFaultInjectionConfig{
		protocol:     common.ShareProtoNvmf,
		fsType:       common.Ext4FsType,
		pvcSize:      params.VolMb,
		replicaCount: params.Replicas,
		scName:       testName + "-sc",
		pvcName:      testName + "-pvc",
		fioPodName:   testName + "-fio",
		duration:     fioDuration,
		thinkTime:    fioThinkTime,
		timeout:      fioTimeout,
	}
	return c
}
