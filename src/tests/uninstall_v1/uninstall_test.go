package uninstall_v1

import (
	"testing"

	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/k8sinstall"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTeardownSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, k8sinstall.UninstallSuiteNameV1, "uninstall")
}

var _ = Describe("Mayastor setup", func() {
	It("should teardown using yamls", func() {
		Expect(controlplane.MajorVersion()).To(Equal(1), "Mayastor version should be 1")
		Expect(k8sinstall.TeardownMayastor()).ToNot(HaveOccurred(), "uninstall failed")
	})
})

var _ = BeforeSuite(func(done Done) {
	err := k8stest.SetupTestEnvBasic()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnvBasic %v", err)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	err := k8stest.TeardownTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)

})
