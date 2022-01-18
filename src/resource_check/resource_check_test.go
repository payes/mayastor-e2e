package resource_check

import (
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Check that there are no artefacts left over from
// running a 3rd party test.
func resourceCheck() {

	err := k8stest.ResourceCheck()
	if err != nil {
		logf.Log.Info("Failed resource check.", "error", err)
	}

	Expect(err).ToNot(HaveOccurred())
}

func TestResourceCheck(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Resource Check Suite", "resource_check")
}

var _ = Describe("Mayastor resource check", func() {
	It("should have no resources allocated", func() {
		resourceCheck()
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
