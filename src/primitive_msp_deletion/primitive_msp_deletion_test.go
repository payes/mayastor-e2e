package primitive_msp_deletion

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestPrimitiveMspDeletionTest(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Primitive Mayastor Pool Deletion Test", "primitive_msp_deletion")
}

func testMspDeletion() {
	for ix := 0; ix < e2e_config.GetConfig().PrimitiveMspDelete.Iterations; ix++ {
		primitiveMspDeletionTest()
	}
}

func primitiveMspDeletionTest() {

	params := e2e_config.GetConfig().PrimitiveMspDelete

	logf.Log.Info("Primitive MayastorPool Deletion Test", "parameters", params)

	// List pools in the cluster
	pools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "Failed to list pools")

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

	// Wait for the custom resources to actually be removed
	Eventually(func() int {
		pl, err := custom_resources.ListMsPools()
		if err != nil {
			return -1
		}
		return len(pl)
	},
		"360s", // timeout
		"2s",   // poll interval
	).Should(BeIdenticalTo(0))

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

	time.Sleep(60 * time.Second)

	Eventually(func() int {
		poolUsage, err := k8stest.GetPoolUsageInCluster()
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Failed to get pool usage %v", err))
		return int(poolUsage)
	},
		params.PoolUsageTimeoutSecs, // timeout
		"1s",                        // polling interval
	).Should(Equal(0), "Pool usage is not 0")

	Eventually(func() int {
		replicas, err := k8stest.ListReplicasInCluster()
		Expect(err).ToNot(HaveOccurred(), "Failed to retrieve list of replicas")
		return len(replicas)
	},
		params.ReplicasTimeoutSecs, // timeout
		"1s",                       // polling interval
	).Should(Equal(0), "Failed while checking replicas")

	msps, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "Failed to list pools")

	err = compareMsps(pools, msps)
	Expect(err).ToNot(HaveOccurred(), "Failed while checking mayastor pool configuration")
}

var _ = Describe("Primitive Mayatstor Pool deletion test", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// TODO workaround for MQ-1536
		k8stest.WorkaroundForMQ1536()
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should verify mayastor pool deletion", func() {
		testMspDeletion()
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

func compareMsps(mspListBefore []v1alpha1.MayastorPool, mspListAfter []v1alpha1.MayastorPool) error {
	mspConfigBefore := make(map[string]v1alpha1.MayastorPool)
	mspConfigAfter := make(map[string]v1alpha1.MayastorPool)

	for _, msp := range mspListBefore {
		mspConfigBefore[msp.Name] = msp
	}

	for _, msp := range mspListAfter {
		mspConfigAfter[msp.Name] = msp
	}
	for i, m := range mspConfigBefore {
		_, ok := mspConfigAfter[i]
		if !ok {
			return fmt.Errorf("Pool not found")
		}
		if m.Status.Capacity != mspConfigAfter[i].Status.Capacity {
			return fmt.Errorf("Failed due to capacity mismatch capacity before: %d capacity after: %d", m.Status.Capacity, mspConfigAfter[i].Status.Capacity)
		}
		if m.Status.Used != mspConfigAfter[i].Status.Used {
			return fmt.Errorf("Failed due to pool usage mismatch usage before: %d usage after: %d", m.Status.Used, mspConfigAfter[i].Status.Used)
		}
		if m.Spec.Disks[0] != mspConfigAfter[i].Spec.Disks[0] {
			return fmt.Errorf("Failed due to disk mismatch disk before: %s disk after: %s", m.Spec.Disks[0], mspConfigAfter[i].Spec.Disks[0])
		}
		if m.Spec.Node != mspConfigAfter[i].Spec.Node {
			return fmt.Errorf("Failed due to node name mismatch node name before: %s node name after: %s", m.Spec.Node, mspConfigAfter[i].Spec.Node)
		}
	}
	return nil
}
