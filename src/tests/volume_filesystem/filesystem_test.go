package volume_filesystem

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	//	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestVolumeFilesystem(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Volume Filesystem Test", "volume_filesystem")
}

var defTimeoutSecs = "60s"

func volumeFilesytemTest(protocol common.ShareProto, volumeType common.VolumeType, fsType common.FileSystemType) {

	scName := strings.ToLower(fmt.Sprintf("volume-filesystem-repl-%d-%s-%s", common.DefaultReplicaCount, string(protocol), volumeType))
	volName := strings.ToLower(fmt.Sprintf("volume-filesystem-repl-%d-%s-%s", common.DefaultReplicaCount, string(protocol), volumeType))

	// Create storage class obj
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithReplicas(common.DefaultReplicaCount).
		WithProtocol(protocol).
		WithFileSystemType(fsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	// Create PVC
	_, err = k8stest.MkPVC(common.DefaultVolumeSizeMb, volName, scName, volumeType, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", volName)
	// Confirm the PVC has been created.
	pvc, getPvcErr := k8stest.GetPVC(volName, common.NSDefault)
	Expect(getPvcErr).To(BeNil(), "PVC creation failed")
	Expect(pvc).ToNot(BeNil(), "PVC creation failed")

	// Wait for the PVC to be bound.
	Eventually(func() coreV1.PersistentVolumeClaimPhase {
		phase, err := k8stest.GetPvcStatusPhase(volName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "failed to get pvc %s phase", volName)
		return phase
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(coreV1.ClaimBound))

	// Refresh the PVC contents, so that we can get the PV name.
	pvc, getPvcErr = k8stest.GetPVC(volName, common.NSDefault)
	Expect(getPvcErr).To(BeNil(), "PVC content is nil")
	Expect(pvc).ToNot(BeNil(), "PVC content is nil")

	// Check for Volumemode in PV
	pv, getPvErr := k8stest.GetPV(pvc.Spec.VolumeName)
	Expect(getPvErr).To(BeNil(), "Error pv is nil")
	Expect(pv).ToNot(BeNil(), "Error pv is nil")

	Expect(coreV1.PersistentVolumeFilesystem == *pv.Spec.VolumeMode).To(BeTrue(), "Volume type is %s", *pv.Spec.VolumeMode)

	// TODO: Add check for default Filesystem type i.e ext4
	// after having clarification on it
	// if fsType == "" {
	// 	Expect(pv.Spec.CSI.FSType == string(common.Ext4FsType)).To(BeTrue(), "Filesystem type is %s", pv.Spec.CSI.FSType)
	// }

	if fsType != "" {
		Expect(string(fsType) == pv.Spec.CSI.FSType).To(BeTrue(), "Filesystem type is %s", pv.Spec.CSI.FSType)
	}

	// Delete the volume
	err = k8stest.RmPVC(volName, scName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", volName)

	// Delete the storage class
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

var _ = Describe("Volume FileSystem check test", func() {

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

	It("Should verify VolumeType FileSystem and FileSystemType xfs ", func() {
		volumeFilesytemTest(common.ShareProtoNvmf, common.VolFileSystem, common.XfsFsType)
	})

	It("Should verify VolumeType FileSystem and FileSystemType ext4", func() {
		volumeFilesytemTest(common.ShareProtoNvmf, common.VolFileSystem, common.Ext4FsType)
	})
	// TODO:: Add It clause for default filesystem type after having clarification
	// It("Should verify VolumeType FileSystem and FileSystemType None", func() {
	// 	volumeFilesytemTest(common.ShareProtoNvmf, common.VolFileSystem, common.NoneFsType)
	// })

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
