package k8sclient

import (
	"fmt"
	"mayastor-e2e/common/mayastorclient"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// ListReplicasInCluster use mayastorclient to enumerate the set of mayastor replicas present in the cluster
func ListReplicasInCluster() ([]mayastorclient.MayastorReplica, error) {
	nodeAddrs, err := GetMayastorNodeIPs()
	if err == nil {
		return mayastorclient.ListReplicas(nodeAddrs)
	}
	return []mayastorclient.MayastorReplica{}, err
}

// RmReplicasInCluster use mayastorclient to remove mayastor replicas present in the cluster
func RmReplicasInCluster() error {
	nodeAddrs, err := GetMayastorNodeIPs()
	if err == nil {
		return mayastorclient.RmNodeReplicas(nodeAddrs)
	}
	return err
}

func WaitForMayastorSockets(addrs []string, timeout string) error {
	var err error
	timeoutSec, err := time.ParseDuration(timeout)
	const sleepTimeSec = 10 // time in seconds
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", timeout, err)
	}

	for ix := 0; ix < int(timeoutSec.Seconds())/sleepTimeSec; ix++ {
		// If this call goes through without an error implies
		// the listeners at the pod have started
		_, err = mayastorclient.ListReplicas(addrs)
		if err != nil {
			logf.Log.Info("Failed t list replicas", "address", addrs, "error", err)
		} else {
			break
		}
		time.Sleep(sleepTimeSec * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to start listener at the pod, address: %s, error: %v", addrs, err)
	}
	return nil
}
