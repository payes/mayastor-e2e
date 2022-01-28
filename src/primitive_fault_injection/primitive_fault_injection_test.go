package primitive_fault_injection

import (
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPrimitiveFaultInjection(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-1499", "MQ-1499")
}

var _ = BeforeSuite(func(done Done) {
	err := k8stest.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnv %v", err)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	err := k8stest.TeardownTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)

})

var _ = Describe("Primitive fault injection tests:", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})
	It("should verify IO when fault is injected into block device", func() {
		c := generatePrimitiveFaultInjectionConfig("primitive-fault-injection")
		c.faultInjectionTest()
	})

})

func (c *primitiveFaultInjectionConfig) faultInjectionTest() {
	c.createSC()
	c.createPVC()
	c.createFio()
	c.verifyVolumeStateOverGrpcAndCrd()
	c.verifyUninterruptedIO()
	c.getNexusDetail()
	c.faultNexusChild()
	c.verifyFaultedReplica()
	c.verifyMsvStatus()
	c.verifyUninterruptedIO()
	c.verifyVolumeStateOverGrpcAndCrd()
	c.waitForFioPodCompletion()
	c.dataIntegrityCheck()
	c.deleteFio()
	c.deletePVC()
	c.deleteSC()
}
