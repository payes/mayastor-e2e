package install

import (
	. "github.com/onsi/ginkgo"
	"mayastor-e2e/common/k8stest"
	"testing"
)

func TestInstallSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	if k8stest.IsControlPlaneMoac() {
		k8stest.InitTesting(t, "Basic Install Suite", "install")
	}
	if k8stest.IsControlPlaneMcp() {
		k8stest.InitTesting(t, "Basic Install Suite (mayastor control plane)", "install")
	}
}

var _ = Describe("Mayastor setup", func() {
	It("should install using yaml files", func() {
		InstallMayastor()
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
