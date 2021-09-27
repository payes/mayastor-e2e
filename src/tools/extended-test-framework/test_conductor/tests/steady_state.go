package tests

import (
	"fmt"
	"mayastor-e2e/lib"
	"time"

	"github.com/google/uuid"

	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/wm/models"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

var violations = []models.WorkloadViolationEnum{
	models.WorkloadViolationEnumRESTARTED,
	models.WorkloadViolationEnumTERMINATED,
	models.WorkloadViolationEnumNOTPRESENT,
}

func SteadyStateTest(testConductor *tc.TestConductor) error {
	if testConductor.Config.Install {
		if err := tc.InstallMayastor(testConductor.Clientset, testConductor.Config.PoolDevice); err != nil {
			return fmt.Errorf("failed to install mayastor %v", err)
		}
	}
	var protocol lib.ShareProto = lib.ShareProtoNvmf
	var mode storageV1.VolumeBindingMode = storageV1.VolumeBindingImmediate
	var sc_name = "steady-state-sc"
	var pvc_name = "steady-state-pvc"
	var fio_name = "steady-state-fio"
	var vol_type = lib.VolRawBlock

	duration, err := time.ParseDuration(testConductor.Config.SteadyStateTest.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse duration %v", err)
	}

	testRun := uuid.New()
	testRunId := testRun.String()

	// create storage class
	err = lib.NewScBuilder().
		WithName(sc_name).
		WithReplicas(testConductor.Config.ReplicaCount).
		WithProtocol(protocol).
		WithNamespace(lib.NSDefault).
		WithVolumeBindingMode(mode).
		BuildAndCreate(testConductor.Clientset)
	if err != nil {
		return fmt.Errorf("failed to create sc %v", err)
	}
	logf.Log.Info("Created storage class", "sc", sc_name)

	// create PV
	pvcname, err := lib.MkPVC(testConductor.Clientset, 64, pvc_name, sc_name, vol_type, lib.NSDefault)
	if err != nil {
		return fmt.Errorf("failed to create pvc %v", err)
	}
	logf.Log.Info("Created pvc", "pvc", pvcname)

	// deploy fio
	err = lib.DeployFio(testConductor.Clientset, fio_name, pvc_name, vol_type, 64, 1000000)
	if err != nil {
		return fmt.Errorf("failed to deploy pod %s, error: %v", fio_name, err)
	}
	logf.Log.Info("Created pod", "pod", fio_name)

	err = tc.AddWorkload(testConductor.Clientset, testConductor.WorkloadMonitorClient, fio_name, lib.NSDefault, violations)
	if err != nil {
		return fmt.Errorf("failed to inform workload monitor of %s, error: %v", fio_name, err)
	}

	err = tc.AddWorkloadsInNamespace(testConductor.Clientset, testConductor.WorkloadMonitorClient, "mayastor", violations)
	if err != nil {
		return fmt.Errorf("failed to inform workload monitor of mayastor pods, error: %v", err)
	}

	err = tc.SendEventInfo(testConductor.TestDirectorClient, "started", tc.SourceInstance)
	if err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	err = tc.SendRunStarted(testConductor.TestDirectorClient, testRunId, "started", "ET-389")
	if err != nil {
		return fmt.Errorf("failed to inform test director of test start, error: %v", err)
	}

	// allow the test to run
	logf.Log.Info("Running test", "duration (s)", duration.Seconds())
	time.Sleep(duration)

	err = tc.DeleteWorkload(testConductor.Clientset, testConductor.WorkloadMonitorClient, fio_name, lib.NSDefault)
	if err != nil {
		return fmt.Errorf("failed to delete workload %s, error: %v", fio_name, err)
	}

	err = tc.DeleteWorkloads(testConductor.Clientset, testConductor.WorkloadMonitorClient)
	if err != nil {
		return fmt.Errorf("failed to delete all registered workloads, error: %v", err)
	}

	err = tc.SendRunCompletedOk(testConductor.TestDirectorClient, testRunId, "finished", "ET-389")
	if err != nil {
		return fmt.Errorf("failed to inform test director of completion, error: %v", err)
	}

	err = tc.SendRunCompletedOk(testConductor.TestDirectorClient, testRunId, "finished", "ET-389")
	if err != nil {
		return fmt.Errorf("failed to inform test director of completion, error: %v", err)
	}
	err = tc.SendEventInfo(testConductor.TestDirectorClient, "finished", tc.SourceInstance)
	if err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}

	return err
}
