package uninstall

import (
	. "github.com/onsi/ginkgo"
	"mayastor-e2e/common/k8stest"
	"testing"
)

func TestTeardownSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	if k8stest.IsControlPlaneMoac() {
		k8stest.InitTesting(t, UninstallSuiteName, "uninstall")
	}
	if k8stest.IsControlPlaneMcp() {
		k8stest.InitTesting(t, MCPUninstallSuiteName, "uninstall")
	}
}

var _ = Describe("Mayastor setup", func() {
	It("should teardown using yamls", func() {
		TeardownMayastor()
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
