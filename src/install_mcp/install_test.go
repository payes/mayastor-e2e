package install_mcp

import (
	. "github.com/onsi/ginkgo"
	"mayastor-e2e/common/k8sinstall"
	"mayastor-e2e/common/k8stest"
	"testing"

	. "github.com/onsi/gomega"
)

func TestInstallSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, k8sinstall.MCPInstallSuiteName, "install")
}

var _ = Describe("Mayastor control plane setup", func() {
	It("should install using yaml files", func() {
		Expect(k8stest.IsControlPlaneMcp()).To(BeTrue(), "Control plane should be MCP")
		Expect(k8sinstall.InstallMayastor()).ToNot(HaveOccurred(), "install failed")
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnvBasic()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	k8stest.TeardownTestEnvNoCleanup()
})
