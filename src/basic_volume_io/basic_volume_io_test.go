// JIRA: CAS-505
// JIRA: CAS-506
package basic_volume_io

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

func TestBasicVolumeIO(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Basic volume IO tests, NVMe-oF TCP and iSCSI", "basic_volume_io")
}

var _ = Describe("Mayastor Volume IO test", func() {

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

	It("should verify an NVMe-oF TCP volume can process IO on a Filesystem volume with immediate binding", func() {
		BasicVolumeIOTest(common.ShareProtoNvmf, common.VolFileSystem, storageV1.VolumeBindingImmediate)
	})

	It("should verify an NVMe-oF TCP volume can process IO on a Raw Block volume with immediate binding", func() {
		BasicVolumeIOTest(common.ShareProtoNvmf, common.VolRawBlock, storageV1.VolumeBindingImmediate)
	})

	It("should verify an NVMe-oF TCP volume can process IO on a Filesystem volume with delayed binding", func() {
		BasicVolumeIOTest(common.ShareProtoNvmf, common.VolFileSystem, storageV1.VolumeBindingWaitForFirstConsumer)
	})

	It("should verify an NVMe-oF TCP volume can process IO on a Raw Block volume with delayed binding", func() {
		BasicVolumeIOTest(common.ShareProtoNvmf, common.VolRawBlock, storageV1.VolumeBindingWaitForFirstConsumer)
	})

})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
