package tests

import (
	"fmt"
	"mayastor-e2e/tools/extended-test-framework/common"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"sync"
	"time"

	//"github.com/go-openapi/strfmt"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

func testVolume(
	testConductor *tc.TestConductor,
	id int,
	testName string,
	vol_spec VolSpec,
	duration time.Duration,
	timeout time.Duration) error {
	var i int
	var endTime = time.Now().Add(duration)
	var finalerr error
	var suffix string

	noOfSpecs := len(vol_spec.vol_types)
	noOfSCs := len(vol_spec.sc_names)
	if noOfSpecs == 0 || noOfSCs == 0 {
		return fmt.Errorf("Invalid volume spec")
	}

	for {
		i = i + 1
		if time.Now().After(endTime) {
			break
		}
		vol_type := vol_spec.vol_types[(i+id)%noOfSpecs]
		sc_name := vol_spec.sc_names[((i+id)/2)%noOfSCs]

		if vol_type == k8sclient.VolFileSystem {
			suffix = "fs"
		} else {
			suffix = "block"
		}
		pvc_name := fmt.Sprintf("%s-pvc-%d-%d-%s-%s", testName, id, i, suffix, sc_name)
		fio_name := fmt.Sprintf("%s-fio-%d-%d-%s-%s", testName, id, i, suffix, sc_name)
		// create PV
		msv_uid, err := k8sclient.MkPVC(
			vol_spec.vol_size_mb,
			pvc_name,
			sc_name,
			vol_type,
			k8sclient.NSDefault,
			false)
		if err != nil {
			finalerr = fmt.Errorf("failed to create pvc %s, error: %v", pvc_name, err)
			break
		}
		logf.Log.Info("Created pvc", "name", pvc_name, "msv UID", msv_uid)

		// deploy fio
		if err = k8sclient.DeployFio(
			fio_name,
			pvc_name,
			vol_type,
			vol_spec.vol_size_mb,
			1,
			testConductor.Config.NonSteadyState.ThinkTime,
			testConductor.Config.NonSteadyState.ThinkTimeBlocks,
		); err != nil {
			finalerr = fmt.Errorf("failed to deploy pod %s, error: %v", fio_name, err)
			break
		}
		logf.Log.Info("Created pod", "pod", fio_name)

		if err = tc.AddWorkload(
			testConductor.WorkloadMonitorClient,
			fio_name,
			k8sclient.NSDefault,
			violations); err != nil {
			finalerr = fmt.Errorf("failed to inform workload monitor of %s, error: %v", fio_name, err)
			break
		}

		// allow the test to run
		// wait for fio to be not running
		if err = WaitPodNotRunning(fio_name, timeout); err != nil {
			finalerr = err
			break
		}

		if err = tc.DeleteWorkload(testConductor.WorkloadMonitorClient, fio_name, k8sclient.NSDefault); err != nil {
			finalerr = fmt.Errorf("failed to delete application workload %s, error = %v", fio_name, err)
			break
		}
		if err = k8sclient.DeletePodIfCompleted(fio_name, k8sclient.NSDefault); err != nil {
			finalerr = fmt.Errorf("failed to delete application %s, error = %v", fio_name, err)
			break
		}
		if err = k8sclient.DeletePVC(pvc_name, k8sclient.NSDefault); err != nil {
			finalerr = fmt.Errorf("failed to delete pvc %s, error = %v", pvc_name, err)
			break
		}
	}
	if finalerr != nil {
		logf.Log.Info("test failed", "error", finalerr)
		if err := SendEventTestCompletedFail(testConductor, finalerr.Error()); err != nil {
			logf.Log.Info("failed to send fail event", "error", err)
		}
	}
	return finalerr
}

func testVolumes(
	concurrentVolumes int,
	testConductor *tc.TestConductor,
	testName string,
	vol_spec VolSpec,
	duration time.Duration,
	timeout time.Duration) error {

	var allerrs error
	var errchan = make(chan error, concurrentVolumes+1)
	var wg sync.WaitGroup

	for i := 0; i < concurrentVolumes; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := testVolume(testConductor, i, testName, vol_spec, duration, timeout)
			if err != nil {
				logf.Log.Info("Test thread exiting", "thread", i, "error", err.Error())
				errchan <- err
			}
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := MonitorCRs(
			testConductor,
			[]string{},
			duration,
			false,
			"",
		)
		if err != nil {
			if senderr := SendEventTestCompletedFail(testConductor, err.Error()); senderr != nil {
				logf.Log.Info("failed to send fail event", "error", senderr)
			}
			logf.Log.Info("Monitor thread exiting", "error", err.Error())
			errchan <- err
		}
	}()

	wg.Wait()
	close(errchan)
	for err := range errchan {
		if allerrs == nil {
			allerrs = err
		} else {
			allerrs = fmt.Errorf("%v: %v", allerrs, err)
		}
	}
	return allerrs
}

func NonSteadyStateTest(testConductor *tc.TestConductor) error {
	var testName = testConductor.Config.TestName
	var combinederr error
	var err error

	common.WaitTestDirector(testConductor.TestDirectorClient)

	if err = SendTestRunToDo(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of test start, error: %v", err)
	}

	var protocol k8sclient.ShareProto = k8sclient.ShareProtoNvmf
	var sc_name = "sc-immed"
	var sc_name_local = "sc-local-wait"

	duration, err := time.ParseDuration(testConductor.Config.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse duration %v", err)
	}

	timeout, err := time.ParseDuration(testConductor.Config.NonSteadyState.Timeout)
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

	// create storage classes
	if err = k8sclient.NewScBuilder().
		WithName(sc_name).
		WithReplicas(testConductor.Config.NonSteadyState.Replicas).
		WithProtocol(protocol).
		WithNamespace(k8sclient.NSDefault).
		WithVolumeBindingMode(storageV1.VolumeBindingImmediate).
		WithLocal(false).
		BuildAndCreate(); err != nil {

		logf.Log.Info("Created storage class failed", "error", err.Error())

		return fmt.Errorf("failed to create sc %v", err)
	}
	logf.Log.Info("Created storage class", "sc", sc_name)

	if err = k8sclient.NewScBuilder().
		WithName(sc_name_local).
		WithReplicas(testConductor.Config.NonSteadyState.Replicas).
		WithProtocol(protocol).
		WithNamespace(k8sclient.NSDefault).
		WithVolumeBindingMode(storageV1.VolumeBindingWaitForFirstConsumer).
		WithLocal(true).
		BuildAndCreate(); err != nil {

		logf.Log.Info("Created storage class failed", "error", err.Error())

		return fmt.Errorf("failed to create sc %v", err)
	}
	logf.Log.Info("Created storage class", "sc", sc_name_local)

	var vol_spec VolSpec
	vol_spec.sc_names = []string{sc_name_local, sc_name}
	vol_spec.vol_types = []k8sclient.VolumeType{k8sclient.VolFileSystem, k8sclient.VolRawBlock}
	vol_spec.vol_size_mb = testConductor.Config.NonSteadyState.VolumeSizeMb

	if err = testVolumes(
		testConductor.Config.NonSteadyState.ConcurrentVols,
		testConductor,
		testName,
		vol_spec,
		duration,
		timeout); err != nil {
		combinederr = err
	}

	if err = tc.DeleteWorkloads(testConductor.WorkloadMonitorClient); err != nil {
		combinederr = fmt.Errorf("%v: failed to delete all registered workloads, err %v", combinederr, err)
		logf.Log.Info("failed to delete all registered workloads", "error", err)
	}

	if err = k8sclient.DeleteSc(sc_name); err != nil {
		combinederr = fmt.Errorf("%v: failed to delete SC %s, error = %v", combinederr, sc_name, err)
		logf.Log.Info("failed to delete SC", "error", err)
	}

	if err = k8sclient.DeleteSc(sc_name_local); err != nil {
		combinederr = fmt.Errorf("%v: failed to delete SC %s, error = %v", combinederr, sc_name, err)
		logf.Log.Info("failed to delete SC", "error", err)
	}

	return combinederr
}
