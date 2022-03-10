package basic_volume_io_iscsi

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	basicVolIO "mayastor-e2e/tests/basic_volume_io"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
)

func TestBasicVolumeIOIscsi(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-1656", "MQ-1656")
}

var _ = Describe("Basic Mayastor Volume IO test iSCSI:", func() {

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

	It("should verify an iSCSI volume can process IO on a Filesystem volume with immediate binding", func() {
		basicVolIO.BasicVolumeIOTest(common.ShareProtoIscsi, common.VolFileSystem, storageV1.VolumeBindingImmediate)
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
