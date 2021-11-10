package primitive_msp_state

import (
	"fmt"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"

	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// verifyMspUsedSize will verify msp used size
func (c *mspStateConfig) verifyMspUsedSize(size int64) {
	// List Pools by CRDs
	crdPools, err := k8stest.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	for _, crdPool := range crdPools {
		err := c.checkPoolUsedSize(crdPool.Name, size)
		Expect(err).ShouldNot(HaveOccurred(), "failed to verify used size of pool %s error %v", crdPool.Name, err)
	}
}

func getPoolCrs() map[string]v1alpha1.MayastorPool {
	// List Pools by CRDs
	crdPools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	crPools := map[string]v1alpha1.MayastorPool{}
	for _, crdPool := range crdPools {
		crPools[crdPool.Name] = crdPool
	}
	return crPools
}

func getPoolsGrpc() []mayastorclient.MayastorPool {
	addrs := k8stest.GetMayastorNodeIPAddresses()
	pools, err := mayastorclient.ListPools(addrs)
	Expect(err).ToNot(HaveOccurred(), "failed to list pools via grpc")
	return pools
}

// verifyMspCrdAndGrpcState verifies the msp details from grpc and crd
func verifyMspCrdAndGrpcState() {
	grpcPools := getPoolsGrpc()
	crPools := getPoolCrs()
	Expect(len(grpcPools)).To(Equal(len(crPools)))

	logf.Log.Info("verifyMspCrdAndGrpcState", "pool count", len(grpcPools))
	Eventually(func() bool {
		grpcPools = getPoolsGrpc()
		crPools = getPoolCrs()
		Expect(len(grpcPools)).To(Equal(len(crPools)))

		for _, gPool := range grpcPools {
			_, ok := crPools[gPool.Name]
			Expect(ok).To(BeTrue(), "pool %s not found in custom resource pools", gPool.Name)
		}

		res := true
		for _, gPool := range grpcPools {
			res = res && verifyMspState(gPool.Name, crPools[gPool.Name], gPool)
			res = res && verifyMspCapacity(gPool.Name, crPools[gPool.Name], gPool)
			res = res && verifyMspUsedSpace(gPool.Name, crPools[gPool.Name], gPool)
		}
		return res
	},
		"180s", // timeout
		"5s",   // polling interval
	).Should(BeTrue())
}

// verifyMspState will verify msp state via crd  and grpc
// gRPC report msp status as "POOL_UNKNOWN","POOL_ONLINE","POOL_DEGRADED","POOL_FAULTED"
// CRD report msp status as "pending", "online", "degraded", "faulted" and "offline"
//pool state can be offline in CRDs but there is no such state in gRPC
func verifyMspState(poolName string, crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.State == controlplane.MspGrpcStateToCrdState(int(grpcPool.State)) {
		status = true
	} else {
		logf.Log.Info("verifyMspState",
			"pool", poolName,
			"CR", crPool.Status.State,
			"gRPC (int)", controlplane.MspGrpcStateToCrdState(int(grpcPool.State)),
			"gRPC", grpcPool.State,
		)
	}
	return status
}

// verifyMspCapacity will verify msp capacity via crd  and grpc
func verifyMspCapacity(poolName string, crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.Capacity == int64(grpcPool.Capacity) {
		status = true
	} else {
		logf.Log.Info("verifyMspState",
			"pool", poolName,
			"CR", crPool.Status.Capacity,
			"gRPC", grpcPool.Capacity,
		)
	}
	return status
}

// verifyMspUsedSpace will verify msp used size via crd  and grpc
func verifyMspUsedSpace(poolName string, crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.Used == int64(grpcPool.Used) {
		status = true
	} else {
		logf.Log.Info("verifyMspUsedSpace", "pool", poolName, "CR", crPool.Status.Used, "gRPC", grpcPool.Used)
	}
	return status
}

// create replicas
func (c *mspStateConfig) createReplica() {
	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())
	for _, node := range nodes {
		if !node.MayastorNode {
			continue
		}
		var nodeAddrs []string
		nodeAddrs = append(nodeAddrs, node.IPAddress)
		mayastorPools, err := mayastorclient.ListPools(nodeAddrs)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(mayastorPools)).ToNot(BeZero(), "Invalid number of pool on node %s", node.NodeName)
		for _, mayastorPool := range mayastorPools {
			Eventually(func() error {
				err = mayastorclient.CreateReplica(node.IPAddress, c.uuid, uint64(c.replicaSize), mayastorPool.Name)
				Expect(err).ToNot(HaveOccurred(), "Failed to create replica by gRPC")
				return nil
			},
				c.poolCreateTimeout.Seconds(), // timeout
				"1s",                          // polling interval
			).Should(BeNil(), "Failed to delete pool")

		}

	}

}

// remove replicas
func removeReplica() {
	err := k8stest.RmReplicasInCluster()
	Expect(err).ToNot(HaveOccurred(), "Failed to remove replicas from cluster")
}

// checkPoolUsedSize verify mayastor pool used size
func (c *mspStateConfig) checkPoolUsedSize(poolName string, replicaSize int64) error {
	timeoutSecs := int(c.poolUsageTimeout.Seconds())
	sleepTimeSecs := int(c.sleepTime.Seconds())
	logf.Log.Info("Waiting for pool used size", "name", poolName, "Expected Used size", replicaSize)
	for ix := 0; ix < (timeoutSecs+sleepTimeSecs-1)/sleepTimeSecs; ix++ {
		time.Sleep(time.Duration(sleepTimeSecs) * time.Second)
		pool, err := k8stest.GetMsPool(poolName)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to get mayastor pool %s %v", poolName, err))
		if pool.Status.Used == replicaSize {
			return nil
		}
	}
	return errors.Errorf("pool %s used size did not reconcile in %d seconds", poolName, timeoutSecs)
}
