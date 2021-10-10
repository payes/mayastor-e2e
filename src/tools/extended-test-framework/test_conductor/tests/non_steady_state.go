package tests

import (
	"fmt"
	"mayastor-e2e/tools/extended-test-framework/common"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
	tc "mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"time"

	"github.com/go-openapi/strfmt"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

func NonSteadyStateTest(testConductor *tc.TestConductor) (testRunId strfmt.UUID, failmessage string, err error) {
	const testName = "non-steady-state"

	// the test run ID is the same as the uuid of the test conductor pod
	tcpod, err := k8sclient.GetPod("test-conductor", common.EtfwNamespace)
	if err != nil {
		err = fmt.Errorf("failed to get tc pod uid, error: %v\n", err)
		return
	}
	testRunId = strfmt.UUID(tcpod.ObjectMeta.UID)

	if err = common.SendTestRunToDo(
		testConductor.TestDirectorClient,
		testRunId,
		"",
		testConductor.Config.Test); err != nil {

		err = fmt.Errorf("failed to inform test director of test start, error: %v", err)
		return
	}

	if err = SendEventTestPreparing(testConductor, testRunId); err != nil {
		err = fmt.Errorf("failed to inform test director of preparation event, error: %v", err)
		return
	}

	if testConductor.Config.Install {
		if err = tc.InstallMayastor(testConductor.Config.PoolDevice); err != nil {
			err = fmt.Errorf("failed to install mayastor %v", err)
			return
		}
	}
	var protocol k8sclient.ShareProto = k8sclient.ShareProtoNvmf
	var mode storageV1.VolumeBindingMode = storageV1.VolumeBindingImmediate
	var pvc_name = ""
	var fio_name = ""
	var sc_name = testName + "-sc"
	var vol_type = k8sclient.VolRawBlock

	duration, err := time.ParseDuration(testConductor.Config.NonSteadyState.Duration)
	if err != nil {
		err = fmt.Errorf("failed to parse duration %v", err)
		return
	}

	vollifetime, err := time.ParseDuration(testConductor.Config.NonSteadyState.VolLifetime)
	if err != nil {
		err = fmt.Errorf("failed to parse duration %v", err)
		return
	}

	if err = common.SendTestRunStarted(
		testConductor.TestDirectorClient,
		testRunId,
		"",
		testConductor.Config.Test); err != nil {
		err = fmt.Errorf("failed to inform test director of test start, error: %v", err)
		return
	}

	if err = SendEventTestStarted(testConductor, testRunId); err != nil {
		err = fmt.Errorf("failed to inform test director of start event, error: %v", err)
		return
	}

	// create storage class
	if err = k8sclient.NewScBuilder().
		WithName(sc_name).
		WithReplicas(testConductor.Config.NonSteadyState.Replicas).
		WithProtocol(protocol).
		WithNamespace(k8sclient.NSDefault).
		WithVolumeBindingMode(mode).
		BuildAndCreate(); err != nil {

		err = fmt.Errorf("failed to create sc %v", err)
		logf.Log.Info("Created storage class failed", "error", err.Error())
		return
	}
	logf.Log.Info("Created storage class", "sc", sc_name)

	if err = tc.AddWorkload(
		testConductor.WorkloadMonitorClient,
		"test-conductor",
		common.EtfwNamespace,
		violations); err != nil {
		err = fmt.Errorf("failed to inform workload monitor of test-conductor, error: %v", err)
		return
	}

	if err = tc.AddWorkloadsInNamespace(
		testConductor.WorkloadMonitorClient,
		"mayastor",
		violations); err != nil {
		err = fmt.Errorf("failed to inform workload monitor of mayastor pods, error: %v", err)
		return
	}

	var endTime = time.Now().Add(duration)
	var i int
	for {
		i = i + 1
		if time.Now().After(endTime) {
			break
		}
		pvc_name = fmt.Sprintf("%s-pvc-%d", testName, i)
		fio_name = fmt.Sprintf("%s-fio-%d", testName, i)
		// create PV
		msv_uid, pvcerr := k8sclient.MkPVC(
			testConductor.Config.NonSteadyState.VolumeSizeMb,
			pvc_name,
			sc_name,
			vol_type,
			k8sclient.NSDefault,
			false)
		if pvcerr != nil {
			failmessage = fmt.Sprintf("failed to create pvc %v\n", err)
			break
		}
		logf.Log.Info("Created pvc", "msv UID", msv_uid)

		// deploy fio
		if err = k8sclient.DeployFio(
			fio_name,
			pvc_name,
			vol_type,
			testConductor.Config.NonSteadyState.VolumeSizeMb,
			1000000); err != nil {
			failmessage = fmt.Sprintf("failed to deploy pod %s, error: %v\n", fio_name, err)
			break
		}
		logf.Log.Info("Created pod", "pod", fio_name)

		if err = tc.AddWorkload(
			testConductor.WorkloadMonitorClient,
			fio_name,
			k8sclient.NSDefault,
			violations); err != nil {
			failmessage = fmt.Sprintf("failed to inform workload monitor of %s, error: %v\n", fio_name, err)
			break
		}

		// ======== TODO implement changes for test, add/remove pods and PVs ========
		// allow the test to run
		logf.Log.Info("Running test", "duration (s)", vollifetime.Seconds())

		failmessage = MonitorCRs(testConductor, []string{msv_uid}, vollifetime, false)

		if err = tc.DeleteWorkload(testConductor.WorkloadMonitorClient, fio_name, k8sclient.NSDefault); err != nil {
			failmessage = failmessage + fmt.Sprintf("failed to delete application workload %s, error = %v\n", fio_name, err)
			logf.Log.Info("failed to delete all application workload", "error", failmessage)
		}
		if err = k8sclient.DeletePod(fio_name, k8sclient.NSDefault); err != nil {
			failmessage = failmessage + fmt.Sprintf("failed to delete application workload %s, error = %v\n", fio_name, err)
			logf.Log.Info("failed to delete pod", "error", failmessage)
		}
		fio_name = ""
		if err = k8sclient.DeletePVC(pvc_name, k8sclient.NSDefault); err != nil {
			failmessage = failmessage + fmt.Sprintf("failed to delete pvc %s, error = %v\n", pvc_name, err)
			logf.Log.Info("failed to delete PVC", "error", failmessage)
		}
		pvc_name = ""
		if failmessage != "" {
			break
		}
	}
	if err = tc.DeleteWorkloads(testConductor.WorkloadMonitorClient); err != nil {
		logf.Log.Info("failed to delete all registered workloads", "error", err)
	}

	if fio_name != "" {
		if err = k8sclient.DeletePod(fio_name, k8sclient.NSDefault); err != nil {
			failmessage = failmessage + fmt.Sprintf("failed to delete application workload %s, error = %v\n", fio_name, err)
			logf.Log.Info("failed to delete pod", "error", err)
		}
	}
	if pvc_name != "" {
		if err = k8sclient.DeletePVC(pvc_name, k8sclient.NSDefault); err != nil {
			failmessage = failmessage + fmt.Sprintf("failed to delete pvc %s, error = %v\n", pvc_name, err)
			logf.Log.Info("failed to delete PVC", "error", err)
		}
	}
	if err = k8sclient.DeleteSc(sc_name); err != nil {
		failmessage = failmessage + fmt.Sprintf("failed to delete SC %s, error = %v\n", sc_name, err)
		logf.Log.Info("failed to delete SC", "error", err)
	}
	return
}
