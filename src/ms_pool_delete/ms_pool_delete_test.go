package ms_pool_delete

import (
	"fmt"
	"strings"
	"testing"

	storageV1 "k8s.io/api/storage/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/k8stest"
)

func TestPooldeletion(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Pool Deletion Test", "ms_pool_delete")
}

var defTimeoutSecs = "60s"

func pooldeletionTest(protocol common.ShareProto, volumeType common.VolumeType, fsType common.FileSystemType, mode storageV1.VolumeBindingMode) {

	scName := strings.ToLower(fmt.Sprintf("pool-deletion-%d-%s-%s", common.DefaultReplicaCount, string(protocol), volumeType))
	volName := strings.ToLower(fmt.Sprintf("pool-deletion-%d-%s-%s", common.DefaultReplicaCount, string(protocol), volumeType))

	// Create storage class
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithReplicas(common.DefaultReplicaCount).
		WithVolumeBindingMode(mode).
		WithProtocol(protocol).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	// Create the volume
	uid := k8stest.MkPVC(
		common.LargeClaimSizeMb,
		volName,
		scName,
		volumeType,
		common.NSDefault,
	)

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

	// Get pool name from mayastorvolume
	replicas, err := k8stest.GetMsvReplicas(uid)
	Expect(err).ToNot(HaveOccurred(), "Failed to get pool name")

	var poolName string
	for _, replica := range replicas {
		poolName = replica.Pool
		break
	}

	// Delete pool
	err = custom_resources.DeleteMsPool(poolName)
	Expect(err).ToNot(HaveOccurred())

	// Get pool
	pool, err := custom_resources.GetMsPool(poolName)
	Expect(err).ToNot(HaveOccurred())
	Expect(pool).ToNot(BeNil())

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	// Delete the storage class
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)

	// RestoreConfiguredPools (re)create pools as defined by the configuration.
	// As part of the tests we may modify the pools, in such test cases
	// the test should delete all pools and recreate the configured set of pools.
	err = k8stest.RestoreConfiguredPools()
	Expect(err).To(BeNil(), "Not all pools are online after restoration")

}

var _ = Describe("Pool deletion check test", func() {

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

	It("Should verify mayastorpool deletion ", func() {
		pooldeletionTest(common.ShareProtoNvmf, common.VolFileSystem, common.XfsFsType, storageV1.VolumeBindingImmediate)
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
