package tests

import (
	"fmt"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
	"mayastor-e2e/tools/extended-test-framework/common/mini_mcp_client"
	"sync"
	"time"

	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

// checkVolumeReplicasNotZero poll the control plane for the number of replicas
// in the given volume for the given number of seconds.
// Return with error if the count is zero within that time.
func checkVolumeReplicasNotZero(ms_ip string, uuid string, seconds int) error {
	for i := 0; i < seconds; i++ {
		reps, err := mini_mcp_client.GetVolumeReplicas(ms_ip, uuid)
		if err != nil {
			return fmt.Errorf("failed to get replicas, err: %v", err)
		}
		if reps == 0 {
			return fmt.Errorf("found zero replicas")
		}
		time.Sleep(time.Second)
	}
	return nil
}

func eliminateVolume(
	testConductor *tc.TestConductor,
	config tc.ReplicaElimination,
	sc_name string,
	duration time.Duration,
	timeout time.Duration) error {

	var endTime = time.Now().Add(duration)
	var err error
	var suffix string
	const podDeletionTimeoutSecs = 120
	const testName = "replica-elimination"

	msNodeIps, err := k8sclient.GetMayastorNodeIPs()
	if err != nil {
		return fmt.Errorf("Failed to get MS node IPs, error: %v", err)
	}
	if len(msNodeIps) == 0 {
		return fmt.Errorf("No MS nodes")
	}
	var cpNodeIp = msNodeIps[0]

	vol_type := k8sclient.VolRawBlock // golang has no ternary operator
	if config.FsVolume != 0 {
		vol_type = k8sclient.VolFileSystem
		suffix = "fs"
	} else {
		suffix = "block"
	}

	pvc_name := fmt.Sprintf("%s-pvc-%s-%s", testName, suffix, sc_name)
	fio_name := fmt.Sprintf("%s-fio-%s-%s", testName, suffix, sc_name)

	// create PVC
	msv_uid, err := k8sclient.MkPVC(
		config.VolumeSizeMb,
		pvc_name,
		sc_name,
		vol_type,
		k8sclient.NSDefault,
		false)
	if err != nil {
		return fmt.Errorf("failed to create pvc %s, error: %v", pvc_name, err)
	}
	logf.Log.Info("Created pvc", "name", pvc_name, "msv UID", msv_uid)

	var uuid string
	var status string
	var devs []deviceDescriptor

	for {
		// While (test_not_failed) Loop:
		if time.Now().After(endTime) {
			break
		}

		// deploy fio
		if err = k8sclient.DeployFioRestarting(
			testConductor.Config.E2eFioImage,
			fio_name,
			pvc_name,
			vol_type,
			config.VolumeSizeMb,
			1000000,
			config.ThinkTime,
			config.ThinkTimeBlocks,
		); err != nil {
			return fmt.Errorf("failed to deploy pod %s, error: %v", fio_name, err)
		}
		logf.Log.Info("Created pod", "pod", fio_name)

		//  Check MSV is in the healthy state
		if uuid, status, err = getOnlyVolume(msNodeIps[0]); err != nil {
			err = fmt.Errorf("failed to get volumes via rest, error: %v", err)
			break
		}
		logf.Log.Info("got volume", "uuid", uuid, "status", status)
		if status != MCP_MSV_ONLINE {
			err = fmt.Errorf("Unexpected volume status, expected %s, got %s", MCP_MSV_ONLINE, status)
			break
		}

		//  Select one Disk Pool (at “random”) as a victim
		//  The selection algorithm used should feature a randomising element.

		logf.Log.Info("======== offline all devices test ========")

		//  Offline disk 0
		if devs, err = offlineDevice(0, testConductor.Config.Msnodes, devs); err != nil {
			break
		}
		//  echo offline | sudo tee /sys/block/sdx/device/state
		//  Wait for MSV state to become degraded
		if err = waitForVolumeStatus(cpNodeIp, uuid, MCP_MSV_DEGRADED); err != nil {
			break
		}

		//  Offline disk 1
		if devs, err = offlineDevice(1, testConductor.Config.Msnodes, devs); err != nil {
			break
		}

		//  Offline disk 2
		if devs, err = offlineDevice(2, testConductor.Config.Msnodes, devs); err != nil {
			break
		}

		// check that the number of replicas does not go to zero
		if err = checkVolumeReplicasNotZero(cpNodeIp, uuid, 100); err != nil {
			logf.Log.Info("vol zero error", "error", err.Error())
			break
		}
		if err := k8sclient.DeletePod(fio_name, k8sclient.NSDefault, podDeletionTimeoutSecs); err != nil {
			break
		}

		//  Online the disks backing the pools
		if devs, err = restoreDevices(devs); err != nil {
			break
		}

		//  Wait for MSV state to become healthy
		if err = waitForVolumeStatus(cpNodeIp, uuid, MCP_MSV_ONLINE); err != nil {
			break
		}
		// deploy fio to verify the data
		// TODO: use a different application to verify data after I/O has errored.
		if err = k8sclient.DeployFioToVerify(
			testConductor.Config.E2eFioImage,
			fio_name,
			pvc_name,
			vol_type,
			config.VolumeSizeMb,
		); err != nil {
			break
		}
		logf.Log.Info("Recreated pod", "pod", fio_name)

		// wait until the pod completes with a success state
		if err = WaitPodNotRunning(fio_name, timeout); err != nil {
			break
		}

		if err = k8sclient.CheckPodAndDelete(fio_name, k8sclient.NSDefault, podDeletionTimeoutSecs); err != nil {
			break
		}

		if err = randomSleep(); err != nil {
			break
		}
	}
	_, _ = restoreDevices(devs) // avoid breaking the node on error
	_ = k8sclient.DeletePod(fio_name, k8sclient.NSDefault, podDeletionTimeoutSecs)

	if locerr := k8sclient.DeletePVC(pvc_name, k8sclient.NSDefault); locerr != nil {
		err = CombineErrors(err, fmt.Errorf("Failed to delete pvc %s, error = %v", pvc_name, locerr))
	}

	if err != nil {
		logf.Log.Info("test failed", "error", err)
	}
	return err
}

func testVolumeElimination(
	testConductor *tc.TestConductor,
	config tc.ReplicaElimination,
	sc_name string,
	duration time.Duration,
	timeout time.Duration) error {

	var combinederr error
	var errchan = make(chan error, 1)
	var wg sync.WaitGroup

	//
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := eliminateVolume(testConductor, config, sc_name, duration, timeout)
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

func ReplicaEliminationTest(testConductor *tc.TestConductor) error {
	var combinederr error
	var err error
	var config = testConductor.Config.ReplicaElimination

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

	timeout, err := GetDuration(config.Timeout)
	if err != nil {
		return fmt.Errorf("failed to parse duration %v", err)
	}

	// create storage classes
	if config.LocalVolume != 0 {
		sc_name = "sc-local-wait"
		if err = k8sclient.NewScBuilder().
			WithName(sc_name).
			WithReplicas(config.Replicas).
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
			WithReplicas(config.Replicas).
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

	combinederr = testVolumeElimination(
		testConductor,
		config,
		sc_name,
		duration,
		timeout,
	)

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
