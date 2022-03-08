package unsupported_protocol

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	"mayastor-e2e/tests/primitive_fuzz_msv"
)

func TestPrimitiveFuzzMsv(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-1503", "MQ-1503")
}

var _ = Describe("Primitive Fuzz MSV Tests:", func() {

	BeforeEach(func() {
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Run Fuzz concurrently with large volume count test", func() {
		params := e2e_config.GetConfig().PrimitiveMsvFuzz
		c := primitive_fuzz_msv.GeneratePrimitiveMsvFuzzConfig("large-volume-fuzz-unsupported-protocol")
		c.Protocol = common.ShareProto(params.UnsupportedProtocol)
		c.VolumeCount = params.VolCount
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()

	})

})

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
