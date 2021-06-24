package primitive_msp_state

import (
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	msvSize = 1073741824 // in bytes
)

type mspStateConfig struct {
	uuid    string
	msvSize int64
}

func generateMspStateConfig(testName string, replicasCount int) *mspStateConfig {
	c := &mspStateConfig{
		msvSize: msvSize,
		uuid:    string(uuid.NewUUID()),
	}
	return c
}
