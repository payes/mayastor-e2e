package primitive_fuzz_msv

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
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

	It("Run Fuzz in serial create/delete multiple volumes", func() {
		params := e2e_config.GetConfig().PrimitiveMsvFuzz
		// Fuzz test with invalid replica
		c := GeneratePrimitiveMsvFuzzConfig("serial-msv-fuzz-invalid-replica")
		c.Replicas = params.InvalidReplicaCount
		c.GeneratePVCSpec()
		c.serialFuzzMsvTest()
		// Fuzz test with unsupported protocol
		c = GeneratePrimitiveMsvFuzzConfig("serial-msv-fuzz-unsupported-protocol")
		c.Protocol = common.ShareProto(params.UnsupportedProtocol)
		c.GeneratePVCSpec()
		c.serialFuzzMsvTest()
		// Fuzz test with unsupported filesystem
		c = GeneratePrimitiveMsvFuzzConfig("serial-msv-fuzz-unsupported-fstype")
		c.FsType = common.FileSystemType(params.UnsupportedFsType)
		c.GeneratePVCSpec()
		c.serialFuzzMsvTest()
		// Fuzz test with incorrect storage class name
		c = GeneratePrimitiveMsvFuzzConfig("serial-msv-fuzz-incorrect-sc-name")
		c.PvcScName = params.IncorrectScName
		c.GeneratePVCSpec()
		c.serialFuzzMsvTest()
		// Fuzz test with invalid pvc size
		c = GeneratePrimitiveMsvFuzzConfig("serial-msv-fuzz-invalid-pvc-size")
		c.PvcSize = params.LargePvcSize
		c.GeneratePVCSpec()
		c.serialFuzzMsvTest()
	})

	It("Run Fuzz concurrently with serial create/delete multiple volumes", func() {
		params := e2e_config.GetConfig().PrimitiveMsvFuzz
		// Fuzz test with invalid replica
		c := GeneratePrimitiveMsvFuzzConfig("iterative-fuzz-invalid-replica")
		c.Replicas = params.InvalidReplicaCount
		c.GeneratePVCSpec()
		c.concurrentMsvOperationInIteration()
		// Fuzz test with unsupported protocol
		c = GeneratePrimitiveMsvFuzzConfig("iterative-fuzz-unsupported-protocol")
		c.Protocol = common.ShareProto(params.UnsupportedProtocol)
		c.GeneratePVCSpec()
		c.concurrentMsvOperationInIteration()
		// Fuzz test with unsupported filesystem
		c = GeneratePrimitiveMsvFuzzConfig("iterative-fuzz-unsupported-fstype")
		c.FsType = common.FileSystemType(params.UnsupportedFsType)
		c.GeneratePVCSpec()
		c.concurrentMsvOperationInIteration()
		// Fuzz test with incorrect storage class name
		c = GeneratePrimitiveMsvFuzzConfig("iterative-fuzz-incorrect-sc-name")
		c.PvcScName = params.IncorrectScName
		c.GeneratePVCSpec()
		c.concurrentMsvOperationInIteration()
		// Fuzz test with invalid pvc size
		c = GeneratePrimitiveMsvFuzzConfig("iterative-fuzz-invalid-pvc-name")
		c.PvcSize = params.LargePvcSize
		c.GeneratePVCSpec()
		c.concurrentMsvOperationInIteration()
	})

	It("Run Fuzz concurrently with concurrent create/delete", func() {
		params := e2e_config.GetConfig().PrimitiveMsvFuzz
		// Fuzz test with invalid replica
		c := GeneratePrimitiveMsvFuzzConfig("concurrent-fuzz-invalid-replica")
		c.Replicas = params.InvalidReplicaCount
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()
		// Fuzz test with unsupported protocol
		c = GeneratePrimitiveMsvFuzzConfig("concurrent-fuzz-unsupported-protocol")
		c.Protocol = common.ShareProto(params.UnsupportedProtocol)
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()
		// Fuzz test with unsupported filesystem
		c = GeneratePrimitiveMsvFuzzConfig("concurrent-fuzz-unsupported-fstype")
		c.FsType = common.FileSystemType(params.UnsupportedFsType)
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()
		// Fuzz test with incorrect storage class name
		c = GeneratePrimitiveMsvFuzzConfig("concurrent-fuzz-incorrect-sc-name")
		c.PvcScName = params.IncorrectScName
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()
		// Fuzz test with invalid pvc size
		c = GeneratePrimitiveMsvFuzzConfig("concurrent-fuzz-invalid-pvc-size")
		c.PvcSize = params.LargePvcSize
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()
	})
	It("Run Fuzz concurrently with large volume count test", func() {
		params := e2e_config.GetConfig().PrimitiveMsvFuzz
		// Fuzz test with invalid replica
		c := GeneratePrimitiveMsvFuzzConfig("large-volume-fuzz-invalid-replica")
		c.Replicas = params.InvalidReplicaCount
		c.VolumeCount = params.VolCount
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()
		// Fuzz test with unsupported protocol
		c = GeneratePrimitiveMsvFuzzConfig("large-volume-fuzz-unsupported-protocol")
		c.Protocol = common.ShareProto(params.UnsupportedProtocol)
		c.VolumeCount = params.VolCount
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()
		// Fuzz test with unsupported filesystem
		c = GeneratePrimitiveMsvFuzzConfig("large-volume-fuzz-unsupported-fstype")
		c.FsType = common.FileSystemType(params.UnsupportedFsType)
		c.VolumeCount = params.VolCount
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()
		// Fuzz test with incorrect storage class name
		c = GeneratePrimitiveMsvFuzzConfig("large-volume-fuzz-incorrect-sc-name")
		c.PvcScName = params.IncorrectScName
		c.VolumeCount = params.VolCount
		c.GeneratePVCSpec()
		c.ConcurrentMsvFuzz()
		// Fuzz test with invalid pvc size
		c = GeneratePrimitiveMsvFuzzConfig("large-volume-fuzz-invalid-pvc-size")
		c.PvcSize = params.LargePvcSize
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
