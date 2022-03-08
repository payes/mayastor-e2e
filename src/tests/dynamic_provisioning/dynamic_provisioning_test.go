package dynamic_provisioning

import (
	"fmt"
	"strings"
	"testing"

	storageV1 "k8s.io/api/storage/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
)

func TestDynamicProvisioning(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Dynamic Provisioning Test", "dynamic_provisioning")
}

var defTimeoutSecs = "60s"

func dynamicProvisioningTest(protocol common.ShareProto, volumeType common.VolumeType, fsType common.FileSystemType, mode storageV1.VolumeBindingMode) {

	scName := strings.ToLower(fmt.Sprintf("dynamic-provisioning-%d-%s-%s", common.DefaultReplicaCount, string(protocol), volumeType))
	volName := strings.ToLower(fmt.Sprintf("dynamic-provisioning-%d-%s-%s", common.DefaultReplicaCount, string(protocol), volumeType))

	// Create storage class
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithReplicas(common.DefaultReplicaCount).
		WithVolumeBindingMode(mode).
		WithProtocol(protocol).
		WithFileSystemType(fsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	// Create PVC
	_, err = k8stest.MkPVC(common.DefaultVolumeSizeMb, volName, scName, volumeType, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", volName)
	// Create pod
	fioPodName := "fio-" + volName
	pod, err := k8stest.CreateFioPod(fioPodName, volName, common.VolFileSystem, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())
	Expect(pod).ToNot(BeNil())

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

	// List nexus in the cluster
	nexuses, err := k8stest.ListNexusesInCluster()
	if err != nil {
		logf.Log.Info("ResourceEachCheck: failed to retrieve list of nexuses")
	}
	logf.Log.Info("ResourceCheck:", "num nexuses", len(nexuses))
	Expect(len(nexuses) == 1).To(BeTrue(), "Nexus not found")

	// List replicas in the cluster
	replicas, err := k8stest.ListReplicasInCluster()
	if err != nil {
		logf.Log.Info("ResourceEachCheck: failed to retrieve list of replicas")
	}
	logf.Log.Info("ResourceCheck:", "num replicas", len(replicas))
	Expect(len(replicas) == common.DefaultReplicaCount).To(BeTrue(), "Replicas not found")

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	err = k8stest.RmPVC(volName, scName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", volName)
	// List nexus in cluster after pod and pvc deletion
	nexuses, err = k8stest.ListNexusesInCluster()
	if err != nil {
		logf.Log.Info("ResourceEachCheck: failed to retrieve list of nexuses")
	}
	logf.Log.Info("ResourceCheck:", "num nexuses", len(nexuses))
	Expect(len(nexuses)).To(BeZero(), "count of nexuses reported via mayastor client is %d", len(nexuses))

	// List replicas in cluster after pod and pvc deletion
	replicas, err = k8stest.ListReplicasInCluster()
	if err != nil {
		logf.Log.Info("ResourceEachCheck: failed to retrieve list of replicas")
	}
	logf.Log.Info("ResourceCheck:", "num replicas", len(replicas))
	Expect(len(replicas)).To(BeZero(), "count of replicas reported via mayastor client is %d", len(nexuses))

	// List mayastorvolumes
	msv, err := k8stest.ListMsvs()
	Expect(err).To(BeNil(), "Error while listing msv")
	Expect(len(msv) == 0).To(BeTrue(), "Msv is not nil %s", msv)

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

	It("Should verify Dynamic Provisioning ", func() {
		dynamicProvisioningTest(common.ShareProtoNvmf, common.VolFileSystem, common.XfsFsType, storageV1.VolumeBindingImmediate)
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
