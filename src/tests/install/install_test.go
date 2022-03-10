package install

import (
	"mayastor-e2e/common/controlplane"
	"testing"

	"mayastor-e2e/common/k8sinstall"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInstallSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	switch controlplane.MajorVersion() {
	case 1:
		k8stest.InitTesting(t, k8sinstall.InstallSuiteNameV1, "install")
	}
}

var _ = Describe("Mayastor setup", func() {
	It("should install using yaml files", func() {
		Expect(k8sinstall.InstallMayastor()).ToNot(HaveOccurred(), "install failed")
	})
})

var _ = BeforeSuite(func(done Done) {
	err := k8stest.SetupTestEnvBasic()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnvBasic %v", err)
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := k8stest.TeardownTestEnvNoCleanup()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)

})
