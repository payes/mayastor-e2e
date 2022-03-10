package patch_install

import (
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Patch the mayastor installation on the cluster under test.
// We deliberately call out to kubectl, rather than constructing the client-go
// objects, so that we can verify the local deploy yaml files are correct.
func patchMayastor() {
	e2eCfg := e2e_config.GetConfig()

	Expect(e2eCfg.ImageTag).ToNot(BeEmpty(),
		"mayastor image tag not defined")
	Expect(e2eCfg.PoolDevice != "").To(BeTrue(),
		"configuration error pools are not defined.")

	imageTag := e2eCfg.ImageTag
	registry := e2eCfg.Registry

	err := k8stest.MayastorDsPatch(registry, imageTag, e2e_config.GetConfig().Platform.MayastorNamespace)
	Expect(err).ToNot(HaveOccurred(), "Patching mayastor daemonset failed")

	err = k8stest.MayastorCsiPatch(registry, imageTag, e2e_config.GetConfig().Platform.MayastorNamespace)
	Expect(err).ToNot(HaveOccurred(), "Patching mayastor CSI daemonset failed")

	err = k8stest.RestartMayastor(240, 120, 120)
	Expect(err).ToNot(HaveOccurred(), "Restarting mayastor failed")

	ready, err := k8stest.MayastorReady(2, 540)
	Expect(err).ToNot(HaveOccurred())
	Expect(ready).To(BeTrue())
}

func TestPatchSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Basic Patch Suite", "patch")
}

var _ = Describe("Mayastor setup", func() {
	It("should patch mayastor installation", func() {
		patchMayastor()
	})
})

var _ = BeforeSuite(func(done Done) {
	err := k8stest.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnv %v", err)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := k8stest.TeardownTestEnvNoCleanup()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)

})
