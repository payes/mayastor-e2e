package tests

import (
	"fmt"
	"mayastor-e2e/tools/extended-test-framework/common"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"

	"time"

	"github.com/google/uuid"

	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

func ReplicaPerturbationTest(testConductor *tc.TestConductor) error {
	const testName = "replica-perturbation"

	if testConductor.Config.Install {
		if err := tc.InstallMayastor(testConductor.Config.PoolDevice); err != nil {
			return fmt.Errorf("failed to install mayastor %v", err)
		}
	}
	var protocol k8sclient.ShareProto = k8sclient.ShareProtoNvmf
	var mode storageV1.VolumeBindingMode = storageV1.VolumeBindingImmediate
	var sc_name = testName + "-sc"
	var pvc_name = testName + "-pvc"
	var fio_name = testName + "-fio"
	var vol_type = k8sclient.VolRawBlock

	if err := tc.AddWorkload(
		testConductor.WorkloadMonitorClient,
		"test-conductor",
		common.EtfwNamespace,
		violations); err != nil {
		return fmt.Errorf("failed to inform workload monitor of test-conductor, error: %v", err)
	}

	if err := SendTestPreparing(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of completion event, error: %v", err)
	}

	duration, err := time.ParseDuration(testConductor.Config.ReplicaPerturbation.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse duration %v", err)
	}

	testRun := uuid.New()
	testRunId := testRun.String()

	// create storage class
	if err := k8sclient.NewScBuilder().
		WithName(sc_name).
		WithReplicas(testConductor.Config.ReplicaPerturbation.Replicas).
		WithProtocol(protocol).
		WithNamespace(k8sclient.NSDefault).
		WithVolumeBindingMode(mode).
		BuildAndCreate(); err != nil {
		return fmt.Errorf("failed to create sc %v", err)
	}
	logf.Log.Info("Created storage class", "sc", sc_name)

	// create PV
	msv_uid, err := k8sclient.MkPVC(
		testConductor.Config.ReplicaPerturbation.VolumeSizeMb,
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
	if err := k8sclient.DeployFio(
		fio_name,
		pvc_name,
		vol_type,
		testConductor.Config.ReplicaPerturbation.VolumeSizeMb,
		1000000); err != nil {
		return fmt.Errorf("failed to deploy pod %s, error: %v", fio_name, err)
	}
	logf.Log.Info("Created pod", "pod", fio_name)

	if err := tc.AddWorkload(
		testConductor.WorkloadMonitorClient,
		fio_name,
		k8sclient.NSDefault,
		violations); err != nil {
		return fmt.Errorf("failed to inform workload monitor of %s, error: %v", fio_name, err)
	}

	if err := tc.AddWorkloadsInNamespace(
		testConductor.WorkloadMonitorClient,
		"mayastor",
		violations); err != nil {
		return fmt.Errorf("failed to inform workload monitor of mayastor pods, error: %v", err)
	}

	if err := SendTestStarted(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of completion event, error: %v", err)
	}

	if err := tc.SendRunStarted(
		testConductor.TestDirectorClient,
		testRunId,
		"",
		testConductor.Config.Test); err != nil {
		return fmt.Errorf("failed to inform test director of test start, error: %v", err)
	}

	// allow the test to run
	// ========= TODO implement perturbations =============
	logf.Log.Info("Running test", "duration (s)", duration.Seconds())
	failmessage := MonitorCRs(testConductor, []string{msv_uid}, duration, false)

	if failmessage != "" {
		if err := SendTestCompletedFail(testConductor, failmessage); err != nil {
			return fmt.Errorf("failed to inform test director of completion, error: %v", err)
		}
	}

	if err := tc.DeleteWorkloads(testConductor.WorkloadMonitorClient); err != nil {
		return fmt.Errorf("failed to delete all registered workloads, error: %v", err)
	}

	if err := k8sclient.DeletePod(fio_name, k8sclient.NSDefault); err != nil {
		return fmt.Errorf("failed to delete pod %s, error: %v", fio_name, err)
	}

	if err := k8sclient.DeletePVC(pvc_name, k8sclient.NSDefault); err != nil {
		return fmt.Errorf("failed to delete PVC %s, error: %v", pvc_name, err)
	}

	if err := k8sclient.DeleteSc(sc_name); err != nil {
		return fmt.Errorf("failed to delete storage class %s, error: %v", sc_name, err)
	}

	if err := tc.SendRunCompletedOk(
		testConductor.TestDirectorClient,
		testRunId,
		"",
		testConductor.Config.Test); err != nil {
		return fmt.Errorf("failed to inform test director of completion, error: %v", err)
	}

	if err := ReportResult(testConductor, failmessage, testRunId); err != nil {
		return fmt.Errorf("failed to report test outcome, error: %v", err)
	}

	return err
}
