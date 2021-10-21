package tests

import (
	"fmt"
	"mayastor-e2e/tools/extended-test-framework/common"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

func SteadyStateTest(testConductor *tc.TestConductor) error {
	var testName = testConductor.Config.TestName
	var err error
	var combinederr error

	// the test run ID is the same as the uuid of the test conductor pod
	common.WaitTestDirector(testConductor.TestDirectorClient)

	if err = SendTestRunToDo(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of test creation, error: %v", err)
	}

	var protocol k8sclient.ShareProto = k8sclient.ShareProtoNvmf
	var mode storageV1.VolumeBindingMode = storageV1.VolumeBindingImmediate
	var sc_name = testName + "-sc"
	var pvc_name = testName + "-pvc"
	var fio_name = testName + "-fio"
	var vol_type = k8sclient.VolRawBlock

	duration, err := time.ParseDuration(testConductor.Config.Duration)
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

	// create storage class
	if err = k8sclient.NewScBuilder().
		WithName(sc_name).
		WithReplicas(testConductor.Config.SteadyState.Replicas).
		WithProtocol(protocol).
		WithNamespace(k8sclient.NSDefault).
		WithVolumeBindingMode(mode).
		BuildAndCreate(); err != nil {

		logf.Log.Info("Created storage class failed", "error", err.Error())
		return fmt.Errorf("failed to create sc %v", err)

	}
	logf.Log.Info("Created storage class", "sc", sc_name)

	// create PV
	msv_uid, err := k8sclient.MkPVC(
		testConductor.Config.SteadyState.VolumeSizeMb,
		pvc_name,
		sc_name,
		vol_type,
		k8sclient.NSDefault,
		false)
	if err != nil {
		return fmt.Errorf("failed to create pvc %v", err)
	}
	logf.Log.Info("Created pvc", "msv UID", msv_uid)

	// deploy fio
	if err = k8sclient.DeployFio(
		fio_name,
		pvc_name,
		vol_type,
		testConductor.Config.SteadyState.VolumeSizeMb,
		1000000,
		testConductor.Config.SteadyState.ThinkTime,
		testConductor.Config.SteadyState.ThinkTimeBlocks,
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

	// allow the test to run
	logf.Log.Info("Running test", "duration (s)", duration.Seconds())
	combinederr = MonitorCRs(testConductor, []string{msv_uid}, duration, false, "")

	if err = tc.DeleteWorkloads(testConductor.WorkloadMonitorClient); err != nil {
		logf.Log.Info("failed to delete all registered workloads", "error", err)
		combinederr = fmt.Errorf("%v: failed to delete all registered workloads, error: %v", combinederr, err)
	}

	if err = k8sclient.DeletePod(fio_name, k8sclient.NSDefault); err != nil {
		logf.Log.Info("failed to delete pod", "pod", fio_name, "error", err)
		combinederr = fmt.Errorf("%v: failed to delete pod %s, error: %v", combinederr, fio_name, err)
	}

	if err = k8sclient.DeletePVC(pvc_name, k8sclient.NSDefault); err != nil {
		logf.Log.Info("failed to delete PVC", "pvc", pvc_name, "error", err)
		combinederr = fmt.Errorf("%v: failed to delete PVC %s, error: %v", combinederr, pvc_name, err)
	}

	if err = k8sclient.DeleteSc(sc_name); err != nil {
		logf.Log.Info("failed to delete storage class", "sc", sc_name, "error", err)
		combinederr = fmt.Errorf("%v: failed to delete storage class %s, error: %v", combinederr, sc_name, err)
	}
	return combinederr
}
