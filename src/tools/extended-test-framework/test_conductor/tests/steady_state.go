package tests

import (
	"fmt"
	"mayastor-e2e/lib"
	"time"

	"github.com/google/uuid"

	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

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
	var failmessage = ""

	if err := SendTestPreparing(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of preparation event, error: %v", err)
	}

	duration, err := time.ParseDuration(testConductor.Config.SteadyState.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse duration %v", err)
	}

	testRun := uuid.New()
	testRunId := testRun.String()

	// create storage class
	if err := lib.NewScBuilder().
		WithName(sc_name).
		WithReplicas(testConductor.Config.SteadyState.Replicas).
		WithProtocol(protocol).
		WithNamespace(lib.NSDefault).
		WithVolumeBindingMode(mode).
		BuildAndCreate(testConductor.Clientset); err != nil {
		return fmt.Errorf("failed to create sc %v", err)
	}
	logf.Log.Info("Created storage class", "sc", sc_name)

	// create PV
	msv_uid, err := lib.MkPVC(
		testConductor.Clientset,
		testConductor.Config.SteadyState.VolumeSizeMb,
		pvc_name,
		sc_name,
		vol_type,
		lib.NSDefault)
	if err != nil {
		return fmt.Errorf("failed to create pvc %v", err)
	}
	logf.Log.Info("Created pvc", "msv UID", msv_uid)

	// deploy fio
	if err := lib.DeployFio(
		testConductor.Clientset,
		fio_name,
		pvc_name,
		vol_type,
		testConductor.Config.SteadyState.VolumeSizeMb,
		1000000); err != nil {
		return fmt.Errorf("failed to deploy pod %s, error: %v", fio_name, err)
	}
	logf.Log.Info("Created pod", "pod", fio_name)

	if err := tc.AddWorkload(
		testConductor.Clientset,
		testConductor.WorkloadMonitorClient,
		fio_name,
		lib.NSDefault,
		violations); err != nil {
		return fmt.Errorf("failed to inform workload monitor of %s, error: %v", fio_name, err)
	}
	if err := tc.AddWorkload(
		testConductor.Clientset,
		testConductor.WorkloadMonitorClient,
		"test-conductor",
		"mayastor-e2e",
		violations); err != nil {
		return fmt.Errorf("failed to inform workload monitor of test-conductor, error: %v", err)
	}
	if err := tc.AddWorkload(
		testConductor.Clientset,
		testConductor.WorkloadMonitorClient,
		"test-director",
		"mayastor-e2e",
		violations); err != nil {
		return fmt.Errorf("failed to inform workload monitor of test-director, error: %v", err)
	}

	if err := tc.AddWorkloadsInNamespace(
		testConductor.Clientset,
		testConductor.WorkloadMonitorClient,
		"mayastor",
		violations); err != nil {
		return fmt.Errorf("failed to inform workload monitor of mayastor pods, error: %v", err)
	}

	if err := SendTestStarted(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of start event, error: %v", err)
	}

	if err := tc.SendRunStarted(
		testConductor.TestDirectorClient,
		testRunId,
		"",
		testConductor.Config.Test); err != nil {
		return fmt.Errorf("failed to inform test director of test start, error: %v", err)
	}

	// allow the test to run
	logf.Log.Info("Running test", "duration (s)", duration.Seconds())
	var waitSecs = 5
	for ix := 0; ; ix = ix + waitSecs {
		if err := CheckMSV(msv_uid); err != nil {
			failmessage = fmt.Sprintf("MSV check failed, err: %s", err.Error())
			break
		}
		if err := CheckPools(3); err != nil {
			failmessage = fmt.Sprintf("MSP check failed, err: %s", err.Error())
			break
		}
		if err := CheckNodes(testConductor.Config.Msnodes); err != nil {
			failmessage = fmt.Sprintf("MSN check failed, err: %s", err.Error())
			break
		}
		if ix > int(duration.Seconds()) {
			break
		}
		time.Sleep(time.Duration(waitSecs) * time.Second)
	}
	if failmessage != "" {
		if err := SendTestCompletedFail(testConductor, failmessage); err != nil {
			return fmt.Errorf("failed to inform test director of completion, error: %v", err)
		}
	}

	if err := tc.DeleteWorkloads(testConductor.Clientset, testConductor.WorkloadMonitorClient); err != nil {
		return fmt.Errorf("failed to delete all registered workloads, error: %v", err)
	}

	if err := lib.DeletePod(testConductor.Clientset, fio_name, lib.NSDefault); err != nil {
		return fmt.Errorf("failed to delete pod %s, error: %v", fio_name, err)
	}

	if err := lib.DeletePVC(testConductor.Clientset, pvc_name, lib.NSDefault); err != nil {
		return fmt.Errorf("failed to delete PVC %s, error: %v", pvc_name, err)
	}

	if err := lib.DeleteSc(testConductor.Clientset, sc_name); err != nil {
		return fmt.Errorf("failed to delete storage class %s, error: %v", sc_name, err)
	}

	if failmessage == "" {
		if err := tc.SendRunCompletedOk(
			testConductor.TestDirectorClient,
			testRunId,
			"",
			testConductor.Config.Test); err != nil {
			return fmt.Errorf("failed to inform test director of completion, error: %v", err)
		}
		if err := SendTestCompletedOk(testConductor); err != nil {
			return fmt.Errorf("failed to inform test director of completion event, error: %v", err)
		}
	} else {
		if err := tc.SendRunCompletedFail(
			testConductor.TestDirectorClient,
			testRunId,
			failmessage,
			testConductor.Config.Test); err != nil {
			return fmt.Errorf("failed to inform test director of completion, error: %v", err)
		}
	}
	return nil
}
