package primitive_msp_state

import (
	"fmt"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	"mayastor-e2e/common/mayastorclient/grpc"

	"time"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// verifyMspUsedSize will verify msp used size
func (c *mspStateConfig) verifyMspUsedSize() {
	// List Pools by CRDs
	crdPools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	for _, crdPool := range crdPools {
		err := c.checkPoolUsedSize(crdPool.Name)
		Expect(err).ShouldNot(HaveOccurred(), "failed to verify used size of pool %s error %v", crdPool.Name, err)
	}
}

// verifyMspCrdAndGrpcState verifies the msp details from grpc and crd
func verifyMspCrdAndGrpcState() {

	nodes, err := k8stest.GetNodeLocs()
	if err != nil {
		logf.Log.Info("list nodes failed", "error", err)
		return
	}

	// List Pools by CRDs
	crdPools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	crPools := map[string]v1alpha1.MayastorPool{}
	for _, crdPool := range crdPools {
		crPools[crdPool.Name] = crdPool
	}

	for _, node := range nodes {
		if !node.MayastorNode {
			continue
		}
		addrs := []string{node.IPAddress}
		grpcPools, err := mayastorclient.ListPools(addrs)
		Expect(err).ToNot(HaveOccurred(), "failed to list pools via grpc")

		if err == nil && len(grpcPools) != 0 {
			for _, gPool := range grpcPools {
				Expect(verifyMspState(crPools[gPool.Name], gPool)).Should(Equal(true))
				Expect(verifyMspCapacity(crPools[gPool.Name], gPool)).Should(Equal(true))
				Expect(verifyMspUsedSpace(crPools[gPool.Name], gPool)).Should(Equal(true))
			}
		} else {
			logf.Log.Info("pools", "count", len(grpcPools), "error", err)
		}
	}
}

// verifyMspState will verify msp state via crd  and grpc
// gRPC report msp status as "POOL_UNKNOWN","POOL_ONLINE","POOL_DEGRADED","POOL_FAULTED"
// CRD report msp status as "unknown", "online", "degraded", "faulted"
// CRDs report as online
func verifyMspState(crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.State == grpcStateToCrdstate(grpcPool.State) {
		status = true
	}
	return status
}

// verifyMspCapacity will verify msp capacity via crd  and grpc
func verifyMspCapacity(crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.Capacity == int64(grpcPool.Capacity) {
		status = true
	}
	return status
}

// verifyMspUsedSpace will verify msp used size via crd  and grpc
func verifyMspUsedSpace(crPool v1alpha1.MayastorPool,
	grpcPool mayastorclient.MayastorPool) bool {
	var status bool
	if crPool.Status.Used == int64(grpcPool.Used) {
		status = true
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
		for _, mayastorPool := range mayastorPools {
			err = mayastorclient.CreateReplica(node.IPAddress, c.uuid, uint64(c.msvSize), mayastorPool.Name)
			Expect(err).ToNot(HaveOccurred(), "Failed to create replica by gRPC")
		}
	}

}

// remove replicas
func (c *mspStateConfig) removeReplica() {
	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())
	var address []string
	for _, node := range nodes {
		if !node.MayastorNode {
			continue
		}
		logf.Log.Info("", "node", node)
		address = append(address, node.IPAddress)
	}
	err = mayastorclient.RmReplicas(address)
	Expect(err).ToNot(HaveOccurred(), "failed to remove replicas")
}

// WaitPodComplete waits until pod is in completed state
func (c *mspStateConfig) checkPoolUsedSize(poolName string) error {
	timeoutSecs := int(c.timeout.Seconds())
	sleepTimeSecs := int(c.sleepTime.Seconds())
	logf.Log.Info("Waiting for pool used size", "name", poolName)
	for ix := 0; ix < (timeoutSecs+sleepTimeSecs-1)/sleepTimeSecs; ix++ {
		time.Sleep(time.Duration(sleepTimeSecs) * time.Second)
		pool, err := custom_resources.GetMsPool(poolName)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to get mayastor pool %s %v", poolName, err))
		if pool.Status.Used == c.msvSize {
			return nil
		}
	}
	return errors.Errorf("pool %s used size did not reconcile in %d seconds", poolName, timeoutSecs)
}

func grpcStateToCrdstate(mspState grpc.PoolState) string {
	switch mspState {
	case 0:
		return "unknown"
	case 1:
		return "online"
	case 2:
		return "degraded"
	default:
		return "faulted"
	}
}
