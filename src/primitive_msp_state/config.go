package primitive_msp_state

import (
	"mayastor-e2e/common/e2e_config"
	"time"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type mspStateConfig struct {
	uuid              string
	replicaSize       int64
	timeout           time.Duration
	sleepTime         time.Duration
	poolCreateTimeout time.Duration
	poolDeleteTimeout time.Duration
	poolUsageTimeout  time.Duration
	iterations        int
}

func generateMspStateConfig() *mspStateConfig {
	params := e2e_config.GetConfig().PrimitiveMspState
	poolUsageTimeout, err := time.ParseDuration(params.PoolUsageTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "timeout configuration string format is invalid.")
	poolUsageSleepTime, err := time.ParseDuration(params.PoolUsageSleepTimeSecs)
	Expect(err).ToNot(HaveOccurred(), "Sleep time configuration string format is invalid.")
	mspCreateTimeout, err := time.ParseDuration(params.PoolCreateTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "Pool creation timeout configuration string format is invalid.")
	mspDeleteTimeout, err := time.ParseDuration(params.PoolDeleteTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "Pool deletion timeout configuration string format is invalid.")
	c := &mspStateConfig{
		replicaSize:       int64(params.ReplicaSize),
		uuid:              string(uuid.NewUUID()),
		timeout:           poolUsageTimeout,
		sleepTime:         poolUsageSleepTime,
		poolCreateTimeout: mspCreateTimeout,
		poolDeleteTimeout: mspDeleteTimeout,
		poolUsageTimeout:  mspDeleteTimeout,
		iterations:        params.IterationCount,
	}
	return c
}
