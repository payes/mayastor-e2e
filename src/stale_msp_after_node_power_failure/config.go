package stale_msp_after_node_power_failure

import (
	"mayastor-e2e/common/platform"
	"mayastor-e2e/common/platform/types"

	. "github.com/onsi/gomega"
)

const (
	defTimeoutSecs = 120 // in seconds
	defWaitTimeout = "600s"
)

var msDeployment = []string{
	"msp-operator",
	"rest",
	"csi-controller",
}

type nodepowerfailureConfig struct {
	platform types.Platform
}

func generateNodePowerFailureConfig(testName string) *nodepowerfailureConfig {
	c := &nodepowerfailureConfig{}

	c.platform = platform.Create()
	Expect(c.platform).ToNot(BeNil())
	return c
}
