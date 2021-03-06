package k8sclient

import (
	"fmt"
	"mayastor-e2e/common/mayastorclient"
	"mayastor-e2e/tools/extended-test-framework/common/custom_resources"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func WaitForPoolCrd() bool {
	const timoSleepSecs = 5
	const timoSecs = 60
	for ix := 0; ix < timoSecs; ix += timoSleepSecs {
		_, err := custom_resources.ListMsPools()
		if err != nil {
			logf.Log.Info("WaitForPoolCrd", "error", err)
		} else {
			return true
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	return false
}

// CreateNamespace create the given namespace
func CreatePools(poolDevice string) error {
	mayastorNodes, err := GetMayastorNodeNames()
	if err != nil {
		return fmt.Errorf("failed to get nodes, err: %v", err)
	}

	numMayastorInstances := len(mayastorNodes)

	logf.Log.Info("Install", "# of mayastor instances", numMayastorInstances)

	if !WaitForPoolCrd() {
		return fmt.Errorf("timed out waiting for pool CRD")
	}

	for _, node := range mayastorNodes {
		_, err := custom_resources.CreateMsPool(node+"-pool", node, []string{poolDevice})
		if err != nil {
			return fmt.Errorf("failed to create pool, err: %v", err)
		}
	}
	// Wait for pools to be online
	const timoSecs = 120
	const timoSleepSecs = 10
	for ix := 0; ix < timoSecs/timoSleepSecs; ix++ {
		time.Sleep(timoSleepSecs * time.Second)

		pools, err := custom_resources.ListMsPools()
		if err != nil {
			return fmt.Errorf("failed to list pools, err: %v", err)
		}
		if len(pools) < numMayastorInstances {
			logf.Log.Info("Install", "no of pools", len(pools))
			continue
		}
		err = custom_resources.CheckAllMsPoolsAreOnline()
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("timed out waiting for pools to be online: %v", err)
	}
	return nil
}

// GetPoolUsageInCluster use mayastorclient to enumerate the set of pools and sum up the pool usage in the cluster
func GetPoolUsageInCluster() (uint64, error) {
	var poolUsage uint64
	pools, err := ListPoolsInCluster()
	if err == nil {
		for _, pool := range pools {
			poolUsage += pool.Used
		}
	}
	return poolUsage, err
}

// ListPoolsInCluster use mayastorclient to enumerate the set of mayastor pools present in the cluster
func ListPoolsInCluster() ([]mayastorclient.MayastorPool, error) {
	nodeAddrs, err := GetMayastorNodeIPs()
	if err == nil {
		return mayastorclient.ListPools(nodeAddrs)
	}
	return []mayastorclient.MayastorPool{}, err
}
