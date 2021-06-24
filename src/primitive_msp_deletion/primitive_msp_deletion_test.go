package primitive_msp_deletion

import (
	"fmt"
	"testing"

	storageV1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
)

func TestPrimitiveMspDeletionTest(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Primitive Mayastor Pool Deletion Test", "primitive_msp_deletion")
}

func primitiveMspDeletionTest(protocol common.ShareProto, volumeType common.VolumeType, fsType common.FileSystemType, mode storageV1.VolumeBindingMode) {

	params := e2e_config.GetConfig().PrimitiveMspDelete

	// List pools in the cluster
	pools, err := custom_resources.ListMsPools()

	var replicaCount int
	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())
	for _, node := range nodes {
		if !node.MayastorNode {
			continue
		}
		replicaUuid := string(uuid.NewUUID())
		var nodeAddrs []string
		nodeAddrs = append(nodeAddrs, node.IPAddress)
		mayastorPools, err := mayastorclient.ListPools(nodeAddrs)
		Expect(err).ToNot(HaveOccurred())
		for _, mayastorPool := range mayastorPools {
			err = mayastorclient.CreateReplica(node.IPAddress, replicaUuid, uint64(params.ReplicaSize), mayastorPool.Name)
			Expect(err).ToNot(HaveOccurred(), "Replica creation failed for pool: %s", mayastorPool.Name)
			replicaCount++
		}
	}

	Eventually(func() int {
		replicas, err := k8stest.ListReplicasInCluster()
		Expect(err).ToNot(HaveOccurred(), "Failed to retrieve list of replicas")
		return len(replicas)
	},
		params.ReplicasTimeoutSecs, // timeout
		"1s",                       // polling interval
	).Should(Equal(replicaCount), "Failed while comparing replicas")

	Eventually(func() int {
		poolUsage, err := k8stest.GetPoolUsageInCluster()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to get pool usage %v", err))
		return int(poolUsage)
	},
		params.PoolUsageTimeoutSecs, // timeout
		"1s",                        // polling interval
	).ShouldNot(Equal(0), "Pool usage is 0")

	// Remove replicas from cluster if exist
	err = k8stest.RmReplicasInCluster()
	Expect(err).ToNot(HaveOccurred(), "failed to remove replica")

	Eventually(func() int {
		replicas, err := k8stest.ListReplicasInCluster()
		Expect(err).ToNot(HaveOccurred(), "Failed to retrieve list of replicas")
		return len(replicas)
	},
		params.ReplicasTimeoutSecs, // timeout
		"1s",                       // polling interval
	).Should(Equal(0), "Failed while comparing replicas")

	Eventually(func() int {
		poolUsage, err := k8stest.GetPoolUsageInCluster()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to get pool usage %v", err))
		return int(poolUsage)
	},
		params.PoolUsageTimeoutSecs, // timeout
		"1s",                        // polling interval
	).Should(Equal(0), "Pool usage is not 0")

	// Delete mayastorpool
	Eventually(func() error {
		for _, pool := range pools {
			err = custom_resources.DeleteMsPool(pool.Name)
			Expect(err).ToNot(HaveOccurred())
		}
		return nil
	},
		params.PoolDeleteTimeoutSecs, // timeout
		"1s",                         // polling interval
	).Should(BeNil(), "Failed to delete pool")

	// Restart mayastor pods
	err = k8stest.RestartMayastorPods(params.MayastorRestartTimeout)
	Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")

	// Create mayastorpools
	Eventually(func() error {
		for _, pool := range pools {
			_, err = custom_resources.CreateMsPool(pool.Name, pool.Spec.Node, pool.Spec.Disks)
			if err != nil {
				return err
			}
		}
		return nil
	},
		params.PoolCreateTimeoutSecs, // timeout
		"1s",                         // polling interval
	).Should(BeNil(), "Failed to create pool")

	// RestoreConfiguredPools (re)create pools as defined by the configuration.
	// As part of the tests we may modify the pools, in such test cases
	// the test should delete all pools and recreate the configured set of pools.
	err = k8stest.RestoreConfiguredPools()
	Expect(err).To(BeNil(), "Not all pools are online after restoration")

}

var _ = Describe("Primitive Mayatstor Pool deletion test", func() {

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

	It("Should verify mayastor pool deletion", func() {
		primitiveMspDeletionTest(common.ShareProtoNvmf, common.VolFileSystem, common.XfsFsType, storageV1.VolumeBindingImmediate)
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
