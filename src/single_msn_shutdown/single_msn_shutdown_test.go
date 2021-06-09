package single_msn_shutdown

import (
	"testing"
	"time"

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
	It("should verify single msn shutdown test", func() {
		c := generateConfig("single-msn-shutdown")
		c.nodeShutdownTest()
	})
})

func (c *shutdownConfig) nodeShutdownTest() {

	var conf *appConfig

	moacNodeName, err := k8stest.GetMoacNodeName()
	Expect(err).ToNot(HaveOccurred())
	Expect(moacNodeName).ToNot(BeEmpty(), "moac pod not found in running state")

	// Create SC, PVC and Application Deployment
	for _, config := range c.config {
		config.createSC()
		config.uuid = config.createPVC()
		config.createDeployment()
		if config.nodeName != moacNodeName {
			conf = config
		}
	}

	// Power off one non moac node on which application is running
	Expect(c.platform.PowerOffNode(conf.nodeName)).ToNot(HaveOccurred(), "PowerOffNode")
	poweredOffNode = conf.nodeName
	time.Sleep(2 * time.Minute)
	verifyNodeNotReady(conf.nodeName)

	verifyMayastorComponentStates(c.numMayastorInstances - 1)

	for _, config := range c.config {
		if config.nodeName == conf.nodeName {
			continue
		}
		// Verify mayastor pods at the other nodes are still running
		config.verifyApplicationPodRunning(true)
	}

	// Poweron the node for other tests to proceed
	Expect(c.platform.PowerOnNode(conf.nodeName)).ToNot(HaveOccurred(), "PowerOnNode")
	poweredOffNode = ""
	verifyNodesReady()
	// Delete deployment, PVC and SC
	for _, config := range c.config {
		config.deleteDeployment()
		config.deletePVC()
		config.deleteSC()
	}

	verifyMayastorComponentStates(c.numMayastorInstances)
	err = k8stest.RestartMayastor(240, 240, 240)
	Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")
}
