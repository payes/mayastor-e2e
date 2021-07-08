package primitive_msp_stress

import (
	"strconv"
	"sync"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
)

func TestMspStressTest(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Primitive MSP stress test", "primitive_msp_stress")
}

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()
	// List pools in the cluster
	err := k8stest.RestoreConfiguredPools()
	Expect(err).To(BeNil(), "Not all pools are online after restoration")
	pools, err := custom_resources.ListMsPools()
	Expect(err).To(BeNil(), "Failed to list pools")

	// Delete mayastorpool
	for _, pool := range pools {
		if PoolCapacity == 0 {
			PoolCapacity = pool.Status.Capacity
		} else {
			// All disks are expected to be of same size for this test
			Expect(pool.Status.Capacity).To(Equal(PoolCapacity), "Pool Capacities do not match")
		}
		if DiskPath == "" {
			// The test picks up only one disk from the initially created pools
			DiskPath = pool.Spec.Disks[0]
		} else {
			Expect(pool.Spec.Disks[0]).To(Equal(DiskPath), "Disk paths do not match")
		}
		err = custom_resources.DeleteMsPool(pool.Name)
		Expect(err).To(BeNil(), "Failed to delete pool")
	}
	close(done)
}, 60)

var _ = AfterSuite(func() {
	// RestoreConfiguredPools (re)create pools as defined by the configuration.
	// As part of the tests we may modify the pools, in such test cases
	// the test should delete all pools and recreate the configured set of pools.
	err := k8stest.RestoreConfiguredPools()
	Expect(err).To(BeNil(), "Not all pools are online after restoration")

	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})

var _ = Describe("Primitive MSP Stress Test", func() {

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

	It("Should verify Sequential Create Delete Pool", func() {
		SequentialCreateDeletePoolTest()
	})
	It("Should verify Concurrent Create Delete Pool Test", func() {
		ConcurrentCreateDeletePoolTest()
	})
	It("Should verify Concurrent Create Delete Pool On Single Node Test", func() {
		ConcurrentCreateDeletePoolOnSingleNodeTest()
	})

})

func SequentialCreateDeletePoolTest() {
	params := e2e_config.GetConfig().PrimitiveMspStressTest

	nodeList, err := k8stest.GetNodeLocsMap()
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	for _, node := range nodeList {
		if !node.MayastorNode {
			delete(nodeList, node.NodeName)
			break
		}
	}
	CreateDeletePools(nodeList, "-pool", params.Iterations, DiskPath, PoolCapacity)
}

func ConcurrentCreateDeletePoolTest() {
	params := e2e_config.GetConfig().PrimitiveMspStressTest
	wg := sync.WaitGroup{}
	nodeList, err := k8stest.GetNodeLocsMap()
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	for _, node := range nodeList {
		if !node.MayastorNode {
			continue
		}
		wg.Add(1)
		go func(node k8stest.NodeLocation) {
			defer GinkgoRecover()
			CreateDeletePools(map[string]k8stest.NodeLocation{node.NodeName: node}, "-pool", params.Iterations, DiskPath, PoolCapacity)
			wg.Done()
		}(node)
	}
	wg.Wait()

}

func ConcurrentCreateDeletePoolOnSingleNodeTest() {
	params := e2e_config.GetConfig().PrimitiveMspStressTest
	Expect(GiB*int64(params.PartitionSizeInGiB*params.PartitionCount) < PoolCapacity).To(Equal(true), "Total of partition sizes exceeding pool size")
	wg := sync.WaitGroup{}
	nodeList, err := k8stest.GetNodeLocsMap()
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	for _, node := range nodeList {
		if !node.MayastorNode {
			continue
		}
		createDiskPartitions(node.IPAddress, params.PartitionCount, params.PartitionSizeInGiB, DiskPath)
		for i := 0; i < params.PartitionCount; i++ {
			wg.Add(1)
			diskPath := DiskPath + strconv.Itoa(i+1)
			poolSuffix := "-pool-" + strconv.Itoa(i)
			go func(node k8stest.NodeLocation) {
				defer GinkgoRecover()
				CreateDeletePools(map[string]k8stest.NodeLocation{node.NodeName: node}, poolSuffix, 1, diskPath, GiB*int64(params.PartitionSizeInGiB))
				wg.Done()
			}(node)
		}
		wg.Wait()
		deleteDiskPartitions(node.IPAddress, params.PartitionCount, DiskPath)
		break
	}
}
