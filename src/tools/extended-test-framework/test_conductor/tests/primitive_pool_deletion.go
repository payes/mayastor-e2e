package tests

import (
	"fmt"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	"mayastor-e2e/tools/extended-test-framework/common/custom_resources"
	"mayastor-e2e/tools/extended-test-framework/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const sleepTimeSec = 10 // sleep time in seconds

func PrimitivePoolDeletionTest(testConductor *tc.TestConductor) error {
	var err error
	var combinederr error

	if err = SendTestRunToDo(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of test creation, error: %v", err)
	}

	duration, err := GetDuration(testConductor.Config.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse duration %v", err)
	}

	if err = SendTestRunStarted(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of test start, error: %v", err)
	}

	if err = tc.AddCommonWorkloads(
		testConductor.WorkloadMonitorClient,
		violations); err != nil {
		return fmt.Errorf("failed add common workloads, error: %v", err)
	}

	for ix := 0; ix < testConductor.Config.PrimitivePoolDeletion.Iterations; ix++ {
		if err = primitivePoolDeletion(testConductor); err != nil {
			return err
		}
	}

	// allow the test to run
	logf.Log.Info("Running test", "duration (s)", duration.Seconds())
	combinederr = MonitorCRs(testConductor, duration, "")

	if err = tc.DeleteWorkloads(testConductor.WorkloadMonitorClient); err != nil {
		logf.Log.Info("failed to delete all registered workloads", "error", err)
		combinederr = fmt.Errorf("%v: failed to delete all registered workloads, error: %v", combinederr, err)
	}

	return combinederr
}

func primitivePoolDeletion(testConductor *tc.TestConductor) error {
	logf.Log.Info("Primitive Mayastor Pool Deletion Test")
	//var combinederr error
	// List pools in the cluster
	params := testConductor.Config.PrimitivePoolDeletion
	pools, err := custom_resources.ListMsPools()
	if err != nil {
		return fmt.Errorf("failed to list mayastor pools via CRD, error: %v", err)
	}

	nodes, err := k8sclient.GetNodeLocs()

	var replicaCount int
	for _, node := range nodes {
		if !node.MayastorNode {
			continue
		}

		replicaUuid := uuid.NewUUID()
		if err != nil {
			return fmt.Errorf("failed to generate new UUID, error: %v", err)
		}
		var nodeAddrs []string
		nodeAddrs = append(nodeAddrs, node.IPAddress)
		mayastorPools, err := mayastorclient.ListPools(nodeAddrs)
		if err != nil {
			return fmt.Errorf("failed to list mayastor pools via gRPC, node: %v, error: %v", nodeAddrs, err)
		}

		for _, mayastorPool := range mayastorPools {
			if err = mayastorclient.CreateReplica(node.IPAddress,
				string(replicaUuid),
				uint64(params.ReplicaSize),
				mayastorPool.Name); err != nil {
				return fmt.Errorf("failed to create pool %s via gRPC, error: %v", mayastorPool.Name, err)
			}
			replicaCount++
		}
	}
	timeoutSec, err := time.ParseDuration(params.ReplicasTimeoutSecs)
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", params.ReplicasTimeoutSecs, err)
	}
	for ix := 0; ix < int(timeoutSec.Seconds())/sleepTimeSec; ix++ {
		var replicas []mayastorclient.MayastorReplica
		replicas, err = k8sclient.ListReplicasInCluster()
		if err != nil {
			logf.Log.Info("Failed to list msv", "error", err)
		} else if len(replicas) == replicaCount {
			break
		}
		time.Sleep(sleepTimeSec * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to match replica count to %d, error: %v", replicaCount, err)
	}

	timeoutSec, err = time.ParseDuration(params.PoolUsageTimeoutSecs)
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", params.PoolUsageTimeoutSecs, err)
	}
	for ix := 0; ix < int(timeoutSec.Seconds())/sleepTimeSec; ix++ {
		var poolUsage uint64
		poolUsage, err = k8sclient.GetPoolUsageInCluster()
		if err != nil {
			logf.Log.Info("Failed to get pool usage", "error", err)
		} else if poolUsage == 0 {
			break
		}
		time.Sleep(sleepTimeSec * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to get mayastor pool usage in cluster, error: %v", err)
	}

	timeoutSec, err = time.ParseDuration(params.PoolDeleteTimeoutSecs)
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", params.PoolDeleteTimeoutSecs, err)
	}
	var combinederr error
	for ix := 0; ix < int(timeoutSec.Seconds())/sleepTimeSec; ix++ {
		for _, pool := range pools {
			err = custom_resources.DeleteMsPool(pool.Name)
			if err != nil {
				combinederr = fmt.Errorf("%v: failed to delete pool %s, error: %v", combinederr, pool.Name, err)
			}
		}
		if combinederr != nil {
			logf.Log.Info("Failed to delete pools", "error", combinederr)
		} else {
			break
		}
		time.Sleep(sleepTimeSec * time.Second)
	}
	if combinederr != nil {
		return fmt.Errorf("failed to delete mayastor pools, error: %v", combinederr)
	}

	timeoutSec, err = time.ParseDuration(params.PoolListTimeoutSecs)
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", params.PoolUsageTimeoutSecs, err)
	}
	for ix := 0; ix < int(timeoutSec.Seconds())/sleepTimeSec; ix++ {
		var msp []v1alpha1.MayastorPool
		msp, err = custom_resources.ListMsPools()
		if err != nil {
			logf.Log.Info("Failed to list pools", "error", err)
		} else if len(msp) == 0 {
			break
		}
		time.Sleep(sleepTimeSec * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to list mayastor pools in cluster, error: %v", err)
	}

	// Restart mayastor pods
	err = k8stest.RestartMayastorPods(params.MayastorRestartTimeout)
	if err != nil {
		return fmt.Errorf("failed to restart mayastor pods, error: %v", err)
	}
	err = k8stest.WaitForMCPPath(params.DefTimeoutSecs)
	if err != nil {
		return fmt.Errorf("failed to start mayastor control plane components, error: %v", err)
	}
	err = k8stest.WaitForMayastorSockets(k8stest.GetMayastorNodeIPAddresses(), params.DefTimeoutSecs)
	if err != nil {
		return fmt.Errorf("failed to start socket to mayastor pod, error: %v", err)
	}

	timeoutSec, err = time.ParseDuration(params.PoolCreateTimeoutSecs)
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", params.PoolCreateTimeoutSecs, err)
	}
	for ix := 0; ix < int(timeoutSec.Seconds())/sleepTimeSec; ix++ {
		for _, pool := range pools {
			_, err = custom_resources.CreateMsPool(pool.Name, pool.Spec.Node, pool.Spec.Disks)
			if err != nil {
				combinederr = fmt.Errorf("%v: failed to create pool %s, error: %v", combinederr, pool.Name, err)
			}
		}
		if combinederr != nil {
			logf.Log.Info("Failed to create pools", "error", combinederr)
		} else {
			break
		}
		time.Sleep(sleepTimeSec * time.Second)
	}
	if combinederr != nil {
		return fmt.Errorf("failed to create mayastor pools, error: %v", combinederr)
	}

	time.Sleep(60 * time.Second)

	timeoutSec, err = time.ParseDuration(params.PoolUsageTimeoutSecs)
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", params.PoolUsageTimeoutSecs, err)
	}
	for ix := 0; ix < int(timeoutSec.Seconds())/sleepTimeSec; ix++ {
		var poolUsage uint64
		poolUsage, err = k8sclient.GetPoolUsageInCluster()
		if err != nil {
			logf.Log.Info("Failed to get pool usage", "error", err)
		} else if poolUsage == 0 {
			break
		}
		time.Sleep(sleepTimeSec * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to match mayastor pool usage to 0, error: %v", err)
	}

	timeoutSec, err = time.ParseDuration(params.ReplicasTimeoutSecs)
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", params.ReplicasTimeoutSecs, err)
	}
	for ix := 0; ix < int(timeoutSec.Seconds())/sleepTimeSec; ix++ {
		var replicas []mayastorclient.MayastorReplica
		replicas, err = k8sclient.ListReplicasInCluster()
		if err != nil {
			logf.Log.Info("Failed to list msv", "error", err)
		} else if len(replicas) == 0 {
			break
		}
		time.Sleep(sleepTimeSec * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to match replica count to 0, error: %v", err)
	}

	msps, err := custom_resources.ListMsPools()
	if err != nil {
		return fmt.Errorf("failed to list pools, error: %v", err)
	}

	err = compareMsps(pools, msps)
	if err != nil {
		return fmt.Errorf("failed while checking mayastor pool configuration, error: %v", err)
	}
	return nil
}

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
			return fmt.Errorf("pool not found")
		}
		if m.Status.Capacity != mspConfigAfter[i].Status.Capacity {
			return fmt.Errorf("failed due to capacity mismatch capacity before: %d capacity after: %d", m.Status.Capacity, mspConfigAfter[i].Status.Capacity)
		}
		if m.Status.Used != mspConfigAfter[i].Status.Used {
			return fmt.Errorf("failed due to pool usage mismatch usage before: %d usage after: %d", m.Status.Used, mspConfigAfter[i].Status.Used)
		}
		if m.Spec.Disks[0] != mspConfigAfter[i].Spec.Disks[0] {
			return fmt.Errorf("failed due to disk mismatch disk before: %s disk after: %s", m.Spec.Disks[0], mspConfigAfter[i].Spec.Disks[0])
		}
		if m.Spec.Node != mspConfigAfter[i].Spec.Node {
			return fmt.Errorf("failed due to node name mismatch node name before: %s node name after: %s", m.Spec.Node, mspConfigAfter[i].Spec.Node)
		}
	}
	return nil
}
