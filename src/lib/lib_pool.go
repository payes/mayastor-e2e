package lib

import (
	"fmt"
	"mayastor-e2e/common/custom_resources"
	"time"

	"k8s.io/client-go/kubernetes"

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
	}
	return false
}

// CreateNamespace create the given namespace
func CreatePools(clientset kubernetes.Clientset, poolDevice string) error {
	mayastorNodes, err := GetMayastorNodeNames(clientset)
	if err == nil {
		for _, node := range mayastorNodes {
			fmt.Println("node ", node)
		}
	} else {
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
			fmt.Printf("no of pools: %d\n", len(pools))
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
