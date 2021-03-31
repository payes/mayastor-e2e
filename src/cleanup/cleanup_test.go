package cleanup

import (
	"testing"

	"mayastor-e2e/common/k8stest"
	rep "mayastor-e2e/common/reporter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestCleanUpCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Clean up cluster", rep.GetReporters("uninstall"))
}

var _ = Describe("Mayastor setup", func() {
	It("should clean up test artefacts in the cluster", func() {
		cleaned := k8stest.CleanUp()
		//		Expect(cleaned).To(BeTrue())
		logf.Log.Info("", "cleaned", cleaned)
	})
})

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
