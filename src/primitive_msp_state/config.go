package primitive_msp_state

import (
	"mayastor-e2e/common/e2e_config"
	"time"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type mspStateConfig struct {
	uuid      string
	msvSize   int64
	timeout   time.Duration
	sleepTime time.Duration
}

func generateMspStateConfig(testName string, replicasCount int) *mspStateConfig {
	params := e2e_config.GetConfig().PrimitiveMspState
	poolUsageTimeout, err := time.ParseDuration(params.PoolUsageTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "timeout configuration string format is invalid.")
	poolUsageSleepTime, err := time.ParseDuration(params.PoolUsageSleepTimeSecs)
	Expect(err).ToNot(HaveOccurred(), "Sleep time configuration string format is invalid.")
	c := &mspStateConfig{
		msvSize:   int64(params.ReplicaSize),
		uuid:      string(uuid.NewUUID()),
		timeout:   poolUsageTimeout,
		sleepTime: poolUsageSleepTime,
	}
	return c
}
