package single_msn_shutdown

import (
	"testing"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/platform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSingleMsnShutdown(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Single msn shutdown test", "single_msn_shutdown")
}

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})

var _ = Describe("Mayastor single msn shutdown test", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if len(poweredOffNode) != 0 {
			platform := platform.Create()
			_ = platform.PowerOnNode(poweredOffNode)
		}
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})
	if !common.IsControlPlaneMcp() {
		It("should verify single non moac msn shutdown test", func() {
			c := generateConfig("single-non-moac-msn-shutdown")
			c.nonMoacNodeShutdownTest()
		})
		It("should verify single moac msn shutdown test", func() {
			c := generateConfig("single-moac-msn-shutdown")
			c.moacNodeShutdownTest()
		})
	} else {
		It("should verify single non core-agent msn shutdown test", func() {
			c := generateConfig("single-non-core-agent-msn-shutdown")
			c.nonCoreAgentNodeShutdownTest()
		})
		It("should verify single core-agent msn shutdown test", func() {
			c := generateConfig("single-core-agent-msn-shutdown")
			c.coreAgentNodeShutdownTest()
		})
	}
})
