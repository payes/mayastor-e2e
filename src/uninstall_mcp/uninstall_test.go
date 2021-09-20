package uninstall_mcp

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/uninstall"
	"testing"
)

func TestTeardownSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, uninstall.MCPUninstallSuiteName, "uninstall")
}

var _ = Describe("Mayastor setup", func() {
	It("should teardown using yamls", func() {
		Expect(k8stest.IsControlPlaneMcp()).To(BeTrue(), "Control plane should be MCP")
		uninstall.TeardownMayastor()
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnvBasic()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
