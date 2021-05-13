package cleanup

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"mayastor-e2e/common/k8stest"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// This is run as a test but is really a utility to restore
// the cluster to usable state and restart mayastor.
func TestRestartMayastor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clean up cluster, and restart Mayastor")
}

var _ = Describe("Mayastor utility: restart mayastor", func() {
	It("should clean up test artefacts in the cluster, and restart mayastor", func() {
		err := k8stest.RestartMayastor(120, 120, 120)
		Expect(err).ToNot(HaveOccurred(), "Restart failed %v", err)
		logf.Log.Info("Mayastor has been restarted")
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
