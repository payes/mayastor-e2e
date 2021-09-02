package stale_msp_after_node_power_failure

import (
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"
)

func (c *nodepowerfailureConfig) verifyNodeNotReady(nodeName string) {
	Eventually(func() bool {
		readyStatus, err := k8stest.IsNodeReady(nodeName, nil)
		Expect(err).ToNot(HaveOccurred())
		return readyStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(false))
}

func (c *nodepowerfailureConfig) verifyMayastorComponentStates() {
	Eventually(func() bool {
		nodeList, err := custom_resources.ListMsNodes()
		Expect(err).ToNot(HaveOccurred(), "ListMsNodes")
		for _, node := range nodeList {
			if node.Status != "online" {
				return false
			}
		}
		ready, err := k8stest.MayastorReady(3, 540)
		Expect(err).ToNot(HaveOccurred())
		return ready
	}, defTimeoutSecs, 5,
	).Should(Equal(true))
}
