package install

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"mayastor-e2e/common/k8sinstall"
	"mayastor-e2e/common/k8stest"
	"testing"
)

func TestInstallSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	if k8stest.IsControlPlaneMoac() {
		k8stest.InitTesting(t, k8sinstall.InstallSuiteName, "install")
	}
	if k8stest.IsControlPlaneMcp() {
		k8stest.InitTesting(t, k8sinstall.MCPInstallSuiteName, "install")
	}
}

var _ = Describe("Mayastor setup", func() {
	It("should install using yaml files", func() {
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
