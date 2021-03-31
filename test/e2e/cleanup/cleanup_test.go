package cleanup

import (
	"e2e-basic/common"
	rep "e2e-basic/common/reporter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestCleanUpCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Clean up cluster", rep.GetReporters("uninstall"))
}

var _ = Describe("Mayastor setup", func() {
	It("should clean up test artefacts in the cluster", func() {
		cleaned := common.CleanUp()
		//		Expect(cleaned).To(BeTrue())
		logf.Log.Info("", "cleaned", cleaned)
	})
})

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))
	common.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	common.TeardownTestEnv()
})
