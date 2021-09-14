package lib

import (
	"fmt"
	"mayastor-e2e/common/custom_resources"

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
		return err
	}

	numMayastorInstances := len(mayastorNodes)

	logf.Log.Info("Install", "# of mayastor instances", numMayastorInstances)

	if !WaitForPoolCrd() {
		return fmt.Errorf("timed out waiting for pool CRD")
	}

	for _, node := range mayastorNodes {
		_, err := custom_resources.CreateMsPool(node+"-pool", node, []string{poolDevice})
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
