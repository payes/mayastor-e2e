package install_v1

import (
	"mayastor-e2e/common/k8sinstall"
	"mayastor-e2e/common/k8stest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInstallSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, k8sinstall.InstallSuiteNameV1, "install")
}

var _ = Describe("Mayastor control plane setup", func() {
	It("should install using yaml files", func() {
		major, err := k8sinstall.GetConfigMajorVersion()
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		Expect(major).To(Equal(1), "Version should be 1")
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
