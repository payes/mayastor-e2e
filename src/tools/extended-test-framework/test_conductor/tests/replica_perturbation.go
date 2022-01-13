package tests

import (
	"fmt"
	e2eagent "mayastor-e2e/common/e2e-agent"
	"mayastor-e2e/tools/extended-test-framework/common/custom_resources"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
	"mayastor-e2e/tools/extended-test-framework/common/mini_mcp_client"
	"strings"
	"sync"
	"time"

	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

const testName = "replica-perturbation"

func getOnlyVolume(ms_ip string) (string, string, string, error) {
	uuid, status, nexus_node, err := mini_mcp_client.GetOnlyVolume(ms_ip)
	return uuid, status, nexus_node, err
}

func getVolumeStatus(ms_ip string, uuid string) (string, error) {
	status, err := mini_mcp_client.GetVolumeStatus(ms_ip, uuid)
	return status, err
}

func waitForVolumeStatus(ms_ip string, uuid string, wantedState string) error {
	for i := 0; ; i++ {
		state, err := getVolumeStatus(ms_ip, uuid)
		if err != nil {
			return fmt.Errorf("failed to get nexus state, error: %v", err)
		}
		if state == wantedState {
			break
		}
		if i == 30 {
			return fmt.Errorf("timed out waiting for nexus to be %s", wantedState)
		}
		logf.Log.Info("volume not ready", "current state", state, "wanted state", wantedState)
		time.Sleep(10 * time.Second)
	}
	logf.Log.Info("volume now in wanted state", "state", wantedState)
	return nil
}

func checkDeviceState(nodeIp string, poolDevice string, state string) error {
	cmd := "cat /sys/block/" + poolDevice + "/device/state"
	res, _ := e2eagent.Exec(nodeIp, cmd)
	res = strings.TrimRight(res, "\n")
	if res != state {
		return fmt.Errorf("unexpected state: expected %s, got %s", state, res)
	} else {
		return nil
	}
}

func setReplicas(ms_ip string, uuid string, replicas int) error {
	err := mini_mcp_client.SetVolumeReplicas(ms_ip, uuid, replicas)
	if err != nil {
		return fmt.Errorf("failed to set replicas to %d, err: %v", replicas, err)
	}
	return err
	//return waitForReplicas(ms_ip, uuid, replicas)
}

func waitForReplicas(ms_ip string, uuid string, replicas int) error {
	for i := 0; i < 100; i++ {
		reps, err := mini_mcp_client.GetVolumeReplicas(ms_ip, uuid)
		if err != nil {
			return fmt.Errorf("failed to get replicas, err: %v", err)
		}
		logf.Log.Info("wait for replicas", "wanted replicas", replicas, "got replicas", reps)
		if reps == replicas {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("wait for replicas to be %d, timed out", replicas)
}

func offlineDevice(index int, nodecount int) (string, string, string, error) {
	var err error
	var nodeIp string
	var poolDevice string
	var nodeName string

	// list the pools
	pools, err := custom_resources.ListMsPools()
	if err != nil {
		return nodeIp, poolDevice, nodeName, fmt.Errorf("failed to list pools, err: %v", err)
	}
	if len(pools) < nodecount {
		return nodeIp, poolDevice, nodeName, fmt.Errorf("Expected %d ips, found %d", nodecount, len(pools))
	}

	// get the node IPs
	locs, err := k8sclient.GetNodeLocs()
	if err != nil {
		return nodeIp, poolDevice, nodeName, fmt.Errorf("MSV grpc check failed to get nodes, err: %s", err.Error())
	}
	if len(locs) < nodecount {
		return nodeIp, poolDevice, nodeName, fmt.Errorf("Expected %d ips, found %d", nodecount, len(locs))
	}
	pool := pools[index]
	nodeName = pool.Spec.Node

	for _, node := range locs {
		if node.NodeName == pool.Spec.Node {
			logf.Log.Info("offlined node", "name", node.NodeName, "IP", node.IPAddress)
			logf.Log.Info("offlined pool", "name", pool.Name, "device", pool.Spec.Disks[0])
			nodeIp = node.IPAddress
			if len(pool.Spec.Disks) != 1 {
				return nodeIp, poolDevice, nodeName, fmt.Errorf("Unexpected number of disks, expected 1 found %d", len(pool.Spec.Disks))
			}
			poolDevice = pool.Spec.Disks[0]
			if !strings.HasPrefix(poolDevice, "/dev/") {
				return nodeIp, poolDevice, nodeName, fmt.Errorf("Unexpected device path %s", poolDevice)
			}
			poolDevice = poolDevice[5:]
			break
		}
	}
	if nodeIp == "" {
		return nodeIp, poolDevice, nodeName, fmt.Errorf("Could not find node for pool %s", pool.Name)
	}
	// we need the IP address of a node and its pool
	// a pool has spec.node, so we can find the node name
	// GetNodeLocs includes the node name as well as the IPs

	res, err := e2eagent.ControlDevice(nodeIp, poolDevice, "offline")
	if err != nil {
		return nodeIp, poolDevice, nodeName, err
	}
	if err = checkDeviceState(nodeIp, poolDevice, "offline"); err != nil {
		return nodeIp, poolDevice, nodeName, err
	}
	logf.Log.Info("offline device succeeded", "device", poolDevice, "node", nodeIp, "response", res)

	return nodeIp, poolDevice, nodeName, err
}

func onlineDevice(nodeIp string, poolDevice string) error {
	var err error

	res, err := e2eagent.ControlDevice(nodeIp, poolDevice, "running")
	if err != nil {
		return err
	}
	if err = checkDeviceState(nodeIp, poolDevice, "running"); err != nil {
		return err
	}
	logf.Log.Info("online device succeeded", "device", poolDevice, "response", res)
	return err
}

func randomSleep() error {
	sleepMinutesSet := []int{2, 5, 10, 30}
	idx, err := EtfwRandom(uint32(len(sleepMinutesSet)))
	if err != nil {
		return err
	}
	sleepMins := sleepMinutesSet[idx]
	logf.Log.Info("sleeping", "minutes", sleepMins)
	time.Sleep(time.Duration(sleepMins) * time.Minute)
	return err
}

func perturbVolume(
	testConductor *tc.TestConductor,
	sc_name string,
	duration time.Duration) error {

	var endTime = time.Now().Add(duration)
	var err error
	var suffix string
	const podDeletionTimeoutSecs = 120

	msNodeIps, err := k8sclient.GetMayastorNodeIPs()
	if err != nil {
		return fmt.Errorf("Failed to get MS node IPs, error: %v", err)
	}
	if len(msNodeIps) == 0 {
		return fmt.Errorf("No MS nodes")
	}
	var cpNodeIp = msNodeIps[0]

	vol_type := k8sclient.VolRawBlock // golang has no ternary operator
	if testConductor.Config.ReplicaPerturbation.FsVolume != 0 {
		vol_type = k8sclient.VolFileSystem
		suffix = "fs"
	} else {
		suffix = "block"
	}

	pvc_name := fmt.Sprintf("%s-pvc-%s-%s", testName, suffix, sc_name)
	fio_name := fmt.Sprintf("%s-fio-%s-%s", testName, suffix, sc_name)

	// create PVC
	msv_uid, err := k8sclient.MkPVC(
		testConductor.Config.ReplicaPerturbation.VolumeSizeMb,
		pvc_name,
		sc_name,
		vol_type,
		k8sclient.NSDefault,
		false)
	if err != nil {
		return fmt.Errorf("failed to create pvc %s, error: %v", pvc_name, err)
	}
	logf.Log.Info("Created pvc", "name", pvc_name, "msv UID", msv_uid)

	// deploy fio
	if err = k8sclient.DeployFio(
		testConductor.Config.E2eFioImage,
		fio_name,
		pvc_name,
		vol_type,
		testConductor.Config.ReplicaPerturbation.VolumeSizeMb,
		1000000,
		testConductor.Config.ReplicaPerturbation.ThinkTime,
		testConductor.Config.ReplicaPerturbation.ThinkTimeBlocks,
	); err != nil {
		return fmt.Errorf("failed to deploy pod %s, error: %v", fio_name, err)
	}
	logf.Log.Info("Created pod", "pod", fio_name)

	if err = tc.AddWorkload(
		testConductor.WorkloadMonitorClient,
		fio_name,
		k8sclient.NSDefault,
		violations); err != nil {
		return fmt.Errorf("failed to inform workload monitor of %s, error: %v", fio_name, err)
	}

	var nexus_node string
	var offlinedNodeIp string
	var offlinedNodeName string
	var uuid string
	var status string
	var poolToAffect int
	var poolDevice string

	// While (test_not_failed) Loop:
	for {
		if time.Now().After(endTime) {
			break
		}

		//  Check the MSV is healthy state
		if uuid, status, nexus_node, err = getOnlyVolume(msNodeIps[0]); err != nil {
			err = fmt.Errorf("failed to get volumes via rest, error: %v", err)
			break
		}
		logf.Log.Info("got volume", "uuid", uuid, "status", status, "nexus node", nexus_node)
		if status != MCP_MSV_ONLINE {
			err = fmt.Errorf("Unexpected volume status, expected %s, got %s", MCP_MSV_ONLINE, status)
			break
		}

		if testConductor.Config.ReplicaPerturbation.OfflineDeviceTest != 0 {
			logf.Log.Info("1) ============ offline / online one backing device ============")
			//  Select one Disk Pool (at “random”) as a victim
			//  The selection algorithm used should feature a randomising element.
			if poolToAffect, err = EtfwRandom(uint32(testConductor.Config.Msnodes)); err != nil {
				break
			}

			//  Offline the disk backing the pool
			if offlinedNodeIp, poolDevice, offlinedNodeName, err = offlineDevice(poolToAffect, testConductor.Config.Msnodes); err != nil {
				break
			}
			logf.Log.Info("made device offline", "index", poolToAffect, "node", offlinedNodeName)
			if offlinedNodeName == nexus_node {
				logf.Log.Info("affected nexus node")
			}

			//  Wait for MSV state to become degraded
			if err = waitForVolumeStatus(cpNodeIp, uuid, MCP_MSV_DEGRADED); err != nil {
				break
			}

			//  Online disk backing the pool
			if err = onlineDevice(offlinedNodeIp, poolDevice); err != nil {
				err = fmt.Errorf("failed to make device online, error: %v", err)
				break
			}
			offlinedNodeIp = ""
			//  Wait for MSV state to become healthy
			if err = waitForVolumeStatus(cpNodeIp, uuid, MCP_MSV_ONLINE); err != nil {
				break
			}

			if err = waitForReplicas(cpNodeIp, uuid, 3); err != nil {
				break
			}

			if testConductor.Config.ReplicaPerturbation.DoRandomSleep != 0 {
				if err = randomSleep(); err != nil {
					break
				}
			}
		}

		if testConductor.Config.ReplicaPerturbation.ReplicasTest != 0 {

			logf.Log.Info("2) ============ decrement / increment replica cound 2<->3 ============")
			logf.Log.Info("about to reduce replica count to 2")

			//  Change MSV replica count (decrement, from 3 to 2)
			if err = setReplicas(cpNodeIp, uuid, 2); err != nil {
				break
			}

			if err = waitForReplicas(cpNodeIp, uuid, 2); err != nil {
				break
			}

			logf.Log.Info("about to increase replica count to 3")
			//  Change MSV replica count (increment, from 2 to 3)
			if err = setReplicas(cpNodeIp, uuid, 3); err != nil {
				break
			}

			//  Wait for MSV state = degraded
			if err = waitForVolumeStatus(cpNodeIp, uuid, MCP_MSV_DEGRADED); err != nil {
				break
			}

			logf.Log.Info("waiting for 3 replicas")
			if err = waitForReplicas(cpNodeIp, uuid, 3); err != nil {
				break
			}

			//  Wait for MSV state to be healthy
			if err = waitForVolumeStatus(cpNodeIp, uuid, MCP_MSV_ONLINE); err != nil {
				break
			}
			if testConductor.Config.ReplicaPerturbation.DoRandomSleep != 0 {
				if err = randomSleep(); err != nil {
					break
				}
			}
		}

		if testConductor.Config.ReplicaPerturbation.OfflineDevAndReplicasTest != 0 {
			logf.Log.Info("3) ============ off/online device and change replica count ============")

			if poolToAffect, err = EtfwRandom(uint32(testConductor.Config.Msnodes)); err != nil {
				break
			}

			// offline a device
			if offlinedNodeIp, poolDevice, offlinedNodeName, err = offlineDevice(poolToAffect, testConductor.Config.Msnodes); err != nil {
				break
			}

			logf.Log.Info("made device offline", "index", poolToAffect, "node", offlinedNodeName)
			if offlinedNodeName == nexus_node {
				logf.Log.Info("affected nexus node")
			}

			// wait for the volume to be degraded
			if err = waitForVolumeStatus(cpNodeIp, uuid, MCP_MSV_DEGRADED); err != nil {
				break
			}

			logf.Log.Info("about to reduce replica count to 2")
			//  Change MSV replica count (decrement, from 3 to 2)
			if err = setReplicas(cpNodeIp, uuid, 2); err != nil {
				break
			}

			//  Wait for MSV state to be healthy
			if err = waitForVolumeStatus(cpNodeIp, uuid, MCP_MSV_ONLINE); err != nil {
				break
			}

			logf.Log.Info("about to increase replica count to 3")
			//  Change MSV replica count (increment, from 2 to 3)
			//  This should fail.
			if err = setReplicas(cpNodeIp, uuid, 3); err == nil {
				err = fmt.Errorf("set replicas to 3 did not fail")
				break
			}

			logf.Log.Info("about to online the device")
			// online the pool
			if err = onlineDevice(offlinedNodeIp, poolDevice); err != nil {
				err = fmt.Errorf("failed to online device, error: %v", err)
				break
			}
			offlinedNodeIp = ""

			logf.Log.Info("about to increase replica count to 3")
			// Change MSV replica count (increment, from 2 to 3)
			// This will fail until the pool is healthy.
			done := false
			for ix := 0; ix < 60; ix++ {
				time.Sleep(time.Duration(time.Second))
				if err = setReplicas(cpNodeIp, uuid, 3); err == nil {
					done = true
					break
				}
				logf.Log.Info("waiting for increase of replica count to succeed")
			}
			if !done {
				err = fmt.Errorf("timed out trying to set replicas to 3")
				break
			}
			//  Wait for MSV state to be healthy
			if err = waitForVolumeStatus(cpNodeIp, uuid, MCP_MSV_ONLINE); err != nil {
				break
			}
			if testConductor.Config.ReplicaPerturbation.DoRandomSleep != 0 {
				if err = randomSleep(); err != nil {
					break
				}
			}
		}
	}
	if err != nil {
		logf.Log.Info("test failed", "error", err)
	}
	if offlinedNodeIp != "" {
		_ = onlineDevice(offlinedNodeIp, poolDevice) // avoid breaking the node
	}
	if locerr := tc.DeleteWorkload(testConductor.WorkloadMonitorClient, fio_name, k8sclient.NSDefault); locerr != nil {
		err = CombineErrors(err, fmt.Errorf("Failed to delete application workloads, err %v", locerr))
	}
	if locerr := k8sclient.DeletePod(fio_name, k8sclient.NSDefault, podDeletionTimeoutSecs); locerr != nil {
		err = CombineErrors(err, fmt.Errorf("Failed to delete application %s, error = %v", fio_name, locerr))
	}
	if locerr := k8sclient.DeletePVC(pvc_name, k8sclient.NSDefault); locerr != nil {
		err = CombineErrors(err, fmt.Errorf("Failed to delete pvc %s, error = %v", pvc_name, locerr))
	}
	return err
}

func testVolumePerturbation(
	testConductor *tc.TestConductor,
	sc_name string,
	duration time.Duration) error {

	var combinederr error
	var errchan = make(chan error, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := perturbVolume(testConductor, sc_name, duration)
		if err != nil {
			logf.Log.Info("Test thread exiting", "error", err.Error())
			errchan <- err
		}
	}()

	wg.Wait()
	close(errchan)
	for err := range errchan {
		combinederr = CombineErrors(combinederr, err)
	}
	return combinederr
}

func ReplicaPerturbationTest(testConductor *tc.TestConductor) error {
	var combinederr error
	var err error

	if err = SendTestRunToDo(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of test start, error: %v", err)
	}

	var protocol k8sclient.ShareProto = k8sclient.ShareProtoNvmf
	var sc_name string

	if err = SendTestRunStarted(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of test start, error: %v", err)
	}

	if err = tc.AddCommonWorkloads(
		testConductor.WorkloadMonitorClient,
		violations); err != nil {
		return fmt.Errorf("failed add common workloads, error: %v", err)
	}

	duration, err := GetDuration(testConductor.Config.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse duration %v", err)
	}

	// create storage classes
	if testConductor.Config.ReplicaPerturbation.LocalVolume != 0 {
		sc_name = "sc-local-wait"
		if err = k8sclient.NewScBuilder().
			WithName(sc_name).
			WithReplicas(testConductor.Config.ReplicaPerturbation.Replicas).
			WithProtocol(protocol).
			WithNamespace(k8sclient.NSDefault).
			WithVolumeBindingMode(storageV1.VolumeBindingWaitForFirstConsumer).
			WithLocal(true).
			BuildAndCreate(); err != nil {

			logf.Log.Info("Created storage class failed", "error", err.Error())

			return fmt.Errorf("failed to create sc %v", err)
		}
	} else {
		sc_name = "sc-immed"
		if err = k8sclient.NewScBuilder().
			WithName(sc_name).
			WithReplicas(testConductor.Config.ReplicaPerturbation.Replicas).
			WithProtocol(protocol).
			WithNamespace(k8sclient.NSDefault).
			WithVolumeBindingMode(storageV1.VolumeBindingImmediate).
			WithLocal(false).
			BuildAndCreate(); err != nil {

			logf.Log.Info("Created storage class failed", "error", err.Error())

			return fmt.Errorf("failed to create sc %v", err)
		}
	}
	logf.Log.Info("Created storage class", "sc", sc_name)

	combinederr = testVolumePerturbation(
		testConductor,
		sc_name,
		duration)

	if err := tc.DeleteWorkloads(testConductor.WorkloadMonitorClient); err != nil {
		combinederr = CombineErrors(combinederr, fmt.Errorf("Failed to delete all registered workloads, error = %v", err))
		logf.Log.Info("failed to delete all registered workloads", "error", err)
	}
	if err = k8sclient.DeleteSc(sc_name); err != nil {
		combinederr = CombineErrors(combinederr, fmt.Errorf("Failed to delete SC %s, error = %v", sc_name, err))
		logf.Log.Info("failed to delete SC", "error", err)
	}

	if combinederr != nil {
		logf.Log.Info("Errors found", "error", combinederr.Error())
	}
	return combinederr
}
