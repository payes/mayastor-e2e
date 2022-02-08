package k8stest

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/mayastorclient"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func WaitForMCPPath(timeout string) error {
	var err error
	timeSec, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", timeout, err)
	}
	for ix := 0; ix < int(timeSec.Seconds())/common.TimeSleepSecs; ix++ {
		time.Sleep(common.TimeSleepSecs * time.Second)
		// If this call goes through implies
		// REST, Core Agent and etcd pods are up and running
		_, err = ListMsvs()
		if err != nil {
			logf.Log.Info("Failed to list msv", "error", err)
		} else {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("one of the rest, core agent or etcd pods are not in running state, error: %v", err)
	}
	return nil
}

func WaitForMayastorSockets(addrs []string, timeout string) error {
	var err error
	timeSec, err := time.ParseDuration(timeout)
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", timeout, err)
	}

	for ix := 0; ix < int(timeSec.Seconds())/common.TimeSleepSecs; ix++ {
		time.Sleep(common.TimeSleepSecs * time.Second)
		// If this call goes through without an error implies
		// the listeners at the pod have started
		_, err = mayastorclient.ListReplicas(addrs)
		if err != nil {
			logf.Log.Info("Failed t list replicas", "address", addrs, "error", err)
		} else {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("failed to start listener at the pod, address: %s, error: %v", addrs, err)
	}
	return nil
}
