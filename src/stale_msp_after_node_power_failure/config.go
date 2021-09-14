package stale_msp_after_node_power_failure

import (
	"mayastor-e2e/common/platform"
	"mayastor-e2e/common/platform/types"

	. "github.com/onsi/gomega"
)

const (
	defTimeoutSecs = 120 // in seconds
)

type nodepowerfailureConfig struct {
	platform types.Platform
}

func generateNodePowerFailureConfig(testName string) *nodepowerfailureConfig {
	c := &nodepowerfailureConfig{}

	c.platform = platform.Create()
	Expect(c.platform).ToNot(BeNil())
	return c
}
