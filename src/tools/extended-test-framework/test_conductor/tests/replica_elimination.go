package tests

import (
	"fmt"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
	"sync"
	"time"

	mayastorGrpc "mayastor-e2e/common/mayastorclient/protobuf"
	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

const storage_tester_image = "ci-registry.mayastor-ci.mayadata.io/mayadata/e2e-storage-tester:latest"

// checkUsableVolumeReplicas poll the mayastor pods for the replicas
// in the given volume for the given number of seconds.
// Return with error if the count is zero within that time or if
// all of the replicas are faulted.
func checkUsableVolumeReplicas(ms_node_ips []string, uuid string, seconds int) error {

	for i := 0; i < seconds; i++ {
		// for debugging
		status, volerr := getVolumeStatus(ms_node_ips[0], uuid)
		if volerr != nil {
			logf.Log.Info("Could not get volume status", "error", volerr.Error())
		} else {
			logf.Log.Info("Volume status", "status", status)
		}

		nexuses, err := k8sclient.GetNexuses(ms_node_ips)
		if err != nil {
			return fmt.Errorf("failed to get nexuses, uuid %s, err: %v", uuid, err)
		}
		if len(nexuses) == 0 {
			return fmt.Errorf("no nexuses found, uuid %s", uuid)
		} else if len(nexuses) != 1 {
			// there could be an orphaned nexus
			logf.Log.Info("unexpected number of nexuses", "count", len(nexuses))
		}
		var foundnexus bool = false
		for _, nexus := range nexuses {
			// the uuid forms the last 36 chars of the uri
			extractedUuid := nexus.DeviceUri
			extractedUuid = extractedUuid[len(extractedUuid)-36:]
			if uuid == extractedUuid {
				foundnexus = true
				logf.Log.Info("found nexus", "uuid", nexus.Uuid, "uri", nexus.DeviceUri)
				if len(nexus.Children) == 0 {
					return fmt.Errorf("no nexus children found, uuid %s", uuid)
				}
				var allfaulted bool = true
				for _, replica := range nexus.Children {
					logf.Log.Info("replica check", "replica state", replica.State)
					if replica.State != mayastorGrpc.ChildState_CHILD_FAULTED {
						allfaulted = false
						break
					}
				}
				if allfaulted {
					return fmt.Errorf("all %d children are faulted, uuid %s", len(nexuses[0].Children), uuid)
				}
				logf.Log.Info("replica check", "replicas", len(nexus.Children), "iteration", i, "uuid", uuid)
				break
			}
		}
		if !foundnexus {
			return fmt.Errorf("could not find nexus with uuid %s", uuid)
		}
		time.Sleep(time.Second)
	}
	logf.Log.Info("replicas still valid after timeout", "timeout seconds", seconds, "uuid", uuid)
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
	var blocksParam = fmt.Sprintf("%d", config.BlocksToWrite)
	const deviceParam = "/dev/sdm"

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
	var storage_tester_name string

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
	var zeroerr error

	for {
		// While (test_not_failed) Loop:
		if time.Now().After(endTime) {
			break
		}

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

		storage_tester_name = "e2e-storage-tester-write"
		// deploy storage tester before any perturbation, write and verify a pattern.
		var args = []string{"./e2e-storage-tester", "-w", "-v", "-n", blocksParam, deviceParam}
		if err = k8sclient.DeployStorageTester(
			storage_tester_image,
			storage_tester_name,
			pvc_name,
			vol_type,
			args,
		); err != nil {
			err = fmt.Errorf("failed to deploy e2e-storage-tester pod %s, error: %v", storage_tester_name, err)
			break
		}
		logf.Log.Info("Created pod.", "pod", storage_tester_name)
		logf.Log.Info("Waiting for e2e-storage-tester to complete.", "pod", storage_tester_name)

		// wait until the pod completes
		if err = WaitPodNotRunning(storage_tester_name, timeout); err != nil {
			break
		}
		// this will fail if the pod has finished with an error
		if err = k8sclient.CheckPodAndDelete(storage_tester_name, k8sclient.NSDefault, podDeletionTimeoutSecs); err != nil {
			break
		}

		//  Offline disk 0
		if devs, err = offlineDevice(0, testConductor.Config.Msnodes, devs); err != nil {
			break
		}

		storage_tester_name = "e2e-storage-tester-read"
		// deploy the storage tester to send IO (reads) to trigger the fault
		args = []string{"./e2e-storage-tester", "-t", "150", "-r", "-n", blocksParam, deviceParam}
		if err = k8sclient.DeployStorageTester(
			storage_tester_image,
			storage_tester_name,
			pvc_name,
			vol_type,
			args,
		); err != nil {
			err = fmt.Errorf("failed to deploy e2e-storage-tester pod %s, error: %v", storage_tester_name, err)
			break
		}
		logf.Log.Info("Created load pod", "pod", storage_tester_name)

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

		// check that the number of replicas does not go to zero and are not all faulted
		if zeroerr = checkUsableVolumeReplicas(msNodeIps, uuid, 100); zeroerr != nil {
			// don't abort just yet, keep the error and see if we get a verification error as well
			logf.Log.Info("vol zero error", "error", zeroerr.Error())
		}

		if !k8sclient.IsPodRunning(storage_tester_name, k8sclient.NSDefault) {
			err = fmt.Errorf("expected pod %s to still be running", storage_tester_name)
			break
		}
		logf.Log.Info("e2e-storage-tester is still running.", "pod", storage_tester_name)

		// wait until the pod finishes
		logf.Log.Info("Waiting for e2e-storage-tester to complete.", "pod", storage_tester_name)
		if err = WaitPodNotRunning(storage_tester_name, timeout); err != nil {
			break
		}
		logf.Log.Info("e2e-storage-tester has completed.", "pod", storage_tester_name)
		// it should have failed due to I/O errors
		if !k8sclient.IsPodFailed(storage_tester_name, k8sclient.NSDefault) {
			err = fmt.Errorf("expected pod %s to have failed", storage_tester_name)
			break
		}
		logf.Log.Info("deleting e2e-storage-tester-read")
		if err = k8sclient.DeletePod(storage_tester_name, k8sclient.NSDefault, podDeletionTimeoutSecs); err != nil {
			logf.Log.Info("failed to delete e2e-storage-tester", "error", err.Error())
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

		// verify the data is valid
		storage_tester_name = "e2e-storage-tester-verify"
		args = []string{"./e2e-storage-tester", "-v", "-n", blocksParam, deviceParam}
		// deploy storage tester
		if err = k8sclient.DeployStorageTester(
			storage_tester_image,
			storage_tester_name,
			pvc_name,
			vol_type,
			args,
		); err != nil {
			err = fmt.Errorf("failed to deploy e2e-storage-tester pod %s, error: %v", storage_tester_name, err)
			break
		}
		logf.Log.Info("Created pod", "pod", storage_tester_name)

		// wait until the pod completes with a success state
		logf.Log.Info("Waiting for e2e-storage-tester to complete.", "pod", storage_tester_name)
		if err = WaitPodNotRunning(storage_tester_name, timeout); err != nil {
			break
		}
		// this will fail if the pod has finished with an error
		if err = k8sclient.CheckPodAndDelete(storage_tester_name, k8sclient.NSDefault, podDeletionTimeoutSecs); err != nil {
			break
		}
		if zeroerr != nil { // we found zero or all-faulted replicas earlier
			break
		}
		if err = randomSleep(); err != nil {
			break
		}
	}
	if zeroerr != nil {
		err = CombineErrors(err, zeroerr)
	}

	_, _ = restoreDevices(devs) // avoid breaking the cluster on error

	if err == nil { // otherwise the pod may exist and we can't delete the PVC
		err = k8sclient.DeletePVC(pvc_name, k8sclient.NSDefault)
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

			logf.Log.Info("failed to create storage class", "storage class", sc_name, "error", err.Error())

			return fmt.Errorf("failed to create sc %s %v", sc_name, err)
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

			logf.Log.Info("failed to create storage class", "storage class", sc_name, "error", err.Error())

			return fmt.Errorf("failed to create sc %s %v", sc_name, err)
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
