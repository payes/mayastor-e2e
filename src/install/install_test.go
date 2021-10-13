package install

import (
	"fmt"
	"testing"

	"mayastor-e2e/common/k8sinstall"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInstallSuite(t *testing.T) {
	major, err := k8sinstall.GetConfigMajorVersion()
	if err != nil {
		panic(err)
	}

	// Initialise test and set class and file names for reports
	switch major {
	case 0:
		k8stest.InitTesting(t, k8sinstall.InstallSuiteName, "install")
	case 1:
		k8stest.InitTesting(t, k8sinstall.InstallSuiteNameV1, "install")
	default:
		panic(fmt.Errorf("unsupported version %d", major))
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
