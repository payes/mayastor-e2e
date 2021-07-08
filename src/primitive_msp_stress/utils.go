package primitive_msp_stress

import (
	"mayastor-e2e/common/custom_resources"
	agent "mayastor-e2e/common/e2e-agent"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	"mayastor-e2e/common/mayastorclient/grpc"
	"strconv"

	. "github.com/onsi/gomega"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	defTimeoutSecs    = "60s"
	GiB               = int64(1024 * 1024 * 1024)
	MiB               = int64(1024 * 1024)
	EstimatedMetaSize = 100 * MiB
	PoolCapacity      int64
	DiskPath          string
)

func CreateDeletePools(nodeList map[string]k8stest.NodeLocation, poolSuffix string, iter int, diskPath string, capacity int64) {
	var err error
	for i := 0; i < iter; i++ {

		// Create mayastorpools
		for _, node := range nodeList {
			poolName := node.NodeName + poolSuffix

			logf.Log.Info("Creating msp", "poolName", poolName)
			_, err = custom_resources.CreateMsPool(poolName, node.NodeName, []string{diskPath})
			Expect(err).To(BeNil(), "Failed to create pool")

			// Duplicate pool creation should fail
			logf.Log.Info("Creating duplicate msp", "poolName", poolName)
			_, err = custom_resources.CreateMsPool(poolName, node.NodeName, []string{diskPath})
			Expect(err).NotTo(BeNil(), "Duplicate pool created")

			logf.Log.Info("Creating fuzz msp", "poolName", poolName+"-fusspool-wrong-node")
			_, err = custom_resources.CreateMsPool(poolName+"-fusspool-wrong-node", "fuzzNode", []string{diskPath})
			Expect(err).To(BeNil(), "Failed to create pool")

			logf.Log.Info("Creating fuzz msp", "poolName", poolName+"-fusspool-wrong-disk")
			_, err = custom_resources.CreateMsPool(poolName+"-fusspool-wrong-disk", node.NodeName, []string{"/dev" + "fuzzPath"})
			Expect(err).To(BeNil(), "Failed to create pool")

			logf.Log.Info("Verifying msps creation")
			Eventually(func() bool {
				return verifyPoolCreated(node.IPAddress, poolName, capacity)
			},
				defTimeoutSecs, // timeout
				"5s",           // polling interval
			).Should(Equal(true))
		}

		// Delete mayastorpools
		logf.Log.Info("Deleting msps")
		for _, node := range nodeList {
			poolName := node.NodeName + poolSuffix

			err = custom_resources.DeleteMsPool(poolName)
			Expect(err).To(BeNil(), "Failed to delete pool")

			err = custom_resources.DeleteMsPool(poolName + "-fusspool-wrong-node")
			Expect(err).To(BeNil(), "Failed to delete pool")
			err = custom_resources.DeleteMsPool(poolName + "-fusspool-wrong-disk")
			Expect(err).To(BeNil(), "Failed to delete pool")

			logf.Log.Info("Verifying msps deletion")
			Eventually(func() bool {
				return verifyPoolDeleted(node.IPAddress, poolName)
			},
				defTimeoutSecs, // timeout
				"5s",           // polling interval
			).Should(Equal(true))
		}

	}
}

func verifyPoolCreated(nodeAddr, poolName string, capacity int64) bool {
	grpcPool, err := mayastorclient.GetPool(poolName, nodeAddr)
	Expect(err).ToNot(HaveOccurred(), "failed to list pools via grpc")

	crdPool, err := custom_resources.GetMsPool(poolName)
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")

	if ok := (grpcPool.State == grpc.PoolState_POOL_ONLINE && crdPool.Status.State == "online"); !ok {
		logf.Log.Info("Failed to verify state", "Expected State", "PoolState_POOL_ONLINE", "grpcPool.State", grpcPool.State, "crdPool.Status.State", crdPool.Status.State)
		return false
	}
	if ok := (CapacityRange(int64(grpcPool.Capacity), capacity) && CapacityRange(crdPool.Status.Capacity, capacity)); !ok {
		logf.Log.Info("Failed to verify capacity", "Expected capacity", capacity, "grpcPool.Capacity", grpcPool.Capacity, "crdPool.Status.Capacity", crdPool.Status.Capacity)
		return false
	}
	if ok := (int64(grpcPool.Used) == 0 && crdPool.Status.Used == 0); !ok {
		logf.Log.Info("Failed to verify used space", "Expected used space", "0", "grpcPool.Used", grpcPool.Used, "crdPool.Status.Used", crdPool.Status.Used)
		return false
	}
	return true
}

func CapacityRange(actual, expected int64) bool {
	if (actual <= expected) || (actual >= expected-EstimatedMetaSize) {
		return true
	}
	return false
}

func verifyPoolDeleted(nodeAddr, poolName string) bool {
	_, err := mayastorclient.GetPool(poolName, nodeAddr)
	if !k8serrors.IsNotFound(err) {
		return false
	}
	_, err = custom_resources.GetMsPool(poolName)
	return k8serrors.IsNotFound(err)
}

func createDiskPartitions(addr string, count int, partitionSizeInGiB int, diskPath string) {
	start := 1
	command := "parted --script " + diskPath + " mklabel gpt"
	logf.Log.Info("Labelling disk before partitioning", "addr", addr, "command", command)
	err := agent.DiskPartition(addr, command)
	Expect(err).ToNot(HaveOccurred(), "Failed to label disk on node %s: ", addr)
	for i := 0; i < count; i++ {
		command := "parted --script " + diskPath + " mkpart primary ext4 " + strconv.Itoa(start) + "GiB" + " " + strconv.Itoa(start+partitionSizeInGiB) + "GiB"
		start += partitionSizeInGiB
		logf.Log.Info("Creating partition", "addr", addr, "command", command)
		err := agent.DiskPartition(addr, command)
		Expect(err).ToNot(HaveOccurred(), "Disk Partitioning failed for node %s: ", addr)
	}
}

func deleteDiskPartitions(addr string, count int, diskPath string) {
	for i := 0; i < count; i++ {
		command := "parted --script " + diskPath + " rm " + strconv.Itoa(i+1)
		logf.Log.Info("Deleting partition", "addr", addr, "command", command)
		err := agent.DiskPartition(addr, command)
		Expect(err).ToNot(HaveOccurred(), "Failed to delete disk Partition on node %s: ", addr)
	}
}
