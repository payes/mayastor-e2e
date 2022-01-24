package uninstall

import (
	"mayastor-e2e/common/controlplane"
	"testing"

	. "github.com/onsi/ginkgo"

	"mayastor-e2e/common/k8sinstall"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"
)

func TestTeardownSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	switch controlplane.MajorVersion() {
	case 1:
		k8stest.InitTesting(t, k8sinstall.UninstallSuiteNameV1, "uninstall")
	}
}

var _ = Describe("Mayastor setup", func() {
	It("should teardown using yamls", func() {
		Expect(k8sinstall.TeardownMayastor()).ToNot(HaveOccurred(), "uninstall failed")
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
