package tests

import (
	"fmt"

	"mayastor-e2e/tools/extended-test-framework/common/custom_resources"

	"mayastor-e2e/common/k8s_lib"
	"mayastor-e2e/tools/extended-test-framework/common"

	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"

	td "mayastor-e2e/tools/extended-test-framework/common/td/models"
	"mayastor-e2e/tools/extended-test-framework/common/wm/models"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"time"

	v1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type VolSpec struct {
	sc_names    []string
	vol_types   []k8sclient.VolumeType
	vol_size_mb int
}

const MCP_NEXUS_ONLINE = "NEXUS_ONLINE"
const MCP_NEXUS_FAULTED = "NEXUS_FAULTED"
const MSV_ONLINE = "healthy"

var violations = []models.WorkloadViolationEnum{
	models.WorkloadViolationEnumRESTARTED,
	models.WorkloadViolationEnumTERMINATED,
	models.WorkloadViolationEnumNOTPRESENT,
}

func CheckMSVmoac(msv_uid string) error {
	msv, err := k8sclient.GetMSV(msv_uid)
	if err != nil {
		return fmt.Errorf("failed to get the MSV, error: %v", err)
	}
	if msv == nil {
		return fmt.Errorf("failed to get the MSV, MSV is nil")
	}
	if msv.Status.State != MSV_ONLINE {
		return fmt.Errorf("MSV is not healthy, state is %s", msv.Status.State)
	}
	return nil
}

func CheckMSVwithGrpc(ms_ips []string) error {
	states, err := k8sclient.CheckVolumeStates(ms_ips)
	if err != nil {
		return fmt.Errorf("MSV grpc check failed, err: %v", err)
	}
	for _, state := range states {
		if state == MCP_NEXUS_FAULTED {
			return fmt.Errorf("MSV is not healthy, state is %s", state)
		}
	}
	return nil
}

func CheckPools(poolcount int) error {
	if err := custom_resources.CheckAllMsPoolsAreOnline(); err != nil {
		return fmt.Errorf("Not all pools are healthy, error: %v", err)
	}
	pools, err := custom_resources.ListMsPools()
	if err != nil {
		return fmt.Errorf("failed to list pools, err: %v", err)
	}
	if len(pools) != poolcount {
		return fmt.Errorf("Wrong number of pools, expected %d, have %d", poolcount, len(pools))
	}
	return nil
}

func CheckNodes(nodecount int) error {
	nodeList, err := custom_resources.ListMsNodes()
	if err != nil {
		return fmt.Errorf("Failed to list MS Nodes, error: %v", err)
	}
	if len(nodeList) != nodecount {
		return fmt.Errorf("Wrong number of nodes, expected %d, have %d", nodecount, len(nodeList))
	}
	for _, node := range nodeList {
		if node.Status != "online" {
			return fmt.Errorf("Found offline node %s", node.Name)
		}
	}
	return nil
}

// if pod_to_check is given, the test wait for the pod to complete and duration is a timeout
// otherwise run the check for the full length of duration.
func MonitorCRs(
	testConductor *tc.TestConductor,
	msv_uids []string,
	duration time.Duration,
	moac bool,
	pod_to_check string) error {

	var endTime = time.Now().Add(duration)
	var waitSecs = 5
	for i := 0; ; i++ {
		if i%100 == 0 {
			logf.Log.Info("Monitoring CRs", "seconds elapsed", i*waitSecs)
		}
		if moac {
			for _, msv := range msv_uids {
				if err := CheckMSVmoac(msv); err != nil {
					return fmt.Errorf("MSV %s check failed, err: %s", msv, err.Error())
				}
			}
		} else {
			ms_ips, err := k8sclient.GetMayastorNodeIPs()
			if err != nil {
				return fmt.Errorf("MSV grpc check failed to get nodes, err: %s", err.Error())
			}
			if err := CheckMSVwithGrpc(ms_ips); err != nil {
				return fmt.Errorf("MSV grpc check failed, err: %s", err.Error())
			}
		}
		if err := CheckPools(testConductor.Config.Msnodes); err != nil {
			return fmt.Errorf("MSP check failed, err: %s", err.Error())
		}
		if moac {
			if err := CheckNodes(testConductor.Config.Msnodes); err != nil {
				return fmt.Errorf("MSN check failed, err: %s", err.Error())
			}
		}
		if pod_to_check != "" {
			pod, err := k8s_lib.GetPod(pod_to_check, k8sclient.NSDefault)
			if err != nil {
				return fmt.Errorf("Failed to get application pod %s, err: %s", pod_to_check, err.Error())
			}
			if pod.Status.Phase != v1.PodRunning {
				break
			}
		}
		if time.Now().After(endTime) {
			if pod_to_check == "" {
				break // the run is time-limited, force stop
			} else {
				// we expected the pod to complete by now
				return fmt.Errorf("Pod is still running, timeout triggered")
			}
		}
		time.Sleep(time.Duration(waitSecs) * time.Second)
	}
	return nil
}

func WaitPodNotRunning(pod_to_check string, timeout time.Duration) error {

	var endTime = time.Now().Add(timeout)
	var waitSecs = 5

	for {
		pod, err := k8s_lib.GetPod(pod_to_check, k8sclient.NSDefault)
		if err != nil {
			return fmt.Errorf("Failed to get application pod %s, err: %s", pod_to_check, err.Error())
		}
		if pod.Status.Phase != v1.PodRunning {
			break
		}
		if time.Now().After(endTime) {
			return fmt.Errorf("Pod %s is still running, timeout triggered", pod_to_check)
		}
		time.Sleep(time.Duration(waitSecs) * time.Second)
	}
	return nil
}

func SendTestRunToDo(testConductor *tc.TestConductor) error {
	if testConductor.Config.SendXrayTest == 1 {
		if err := common.SendTestRunToDo(
			testConductor.TestDirectorClient,
			testConductor.TestRunId,
			"",
			testConductor.Config.Test); err != nil {

			return fmt.Errorf("failed to create test run, error: %v", err)
		}
	}
	if err := SendEventTestPreparing(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of preparation event, error: %v", err)
	}
	return nil
}

func SendTestRunStarted(testConductor *tc.TestConductor) error {
	if testConductor.Config.SendXrayTest == 1 {
		if err := common.SendTestRunStarted(
			testConductor.TestDirectorClient,
			testConductor.TestRunId,
			"",
			testConductor.Config.Test); err != nil {

			return fmt.Errorf("failed to set test run to executing, error: %v", err)
		}
	}
	if err := SendEventTestStarted(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of start event, error: %v", err)
	}
	return nil
}

func SendTestRunFinished(testConductor *tc.TestConductor, err error) error {
	if testConductor.Config.SendXrayTest == 1 {
		if err == nil {
			if err := common.SendTestRunCompletedOk(
				testConductor.TestDirectorClient,
				testConductor.TestRunId,
				"",
				testConductor.Config.Test); err != nil {
				return fmt.Errorf("failed to set test run to completed, error: %v", err)
			}
		} else {
			if err := common.SendTestRunCompletedFail(
				testConductor.TestDirectorClient,
				testConductor.TestRunId,
				err.Error(),
				testConductor.Config.Test); err != nil {
				return fmt.Errorf("failed to set test run to failed, error: %v", err)
			}
		}
	}
	if err == nil {
		if err := SendEventTestCompletedOk(testConductor); err != nil {
			return fmt.Errorf("failed to inform test director of completion event, error: %v", err)
		}
	} else {
		if err := SendEventTestCompletedFail(testConductor, err.Error()); err != nil {
			return fmt.Errorf("failed to inform test director of completion event, error: %v", err)
		}
	}
	return nil
}

func SendEventTestPreparing(testConductor *tc.TestConductor) error {
	return SendEvent(testConductor, td.EventClassEnumINFO, "Test preparing, "+testConductor.Config.Test)
}

func SendEventTestStarted(testConductor *tc.TestConductor) error {
	return SendEvent(testConductor, td.EventClassEnumINFO, "Test started, "+testConductor.Config.Test)
}

func SendEventTestCompletedOk(testConductor *tc.TestConductor) error {
	return SendEvent(testConductor, td.EventClassEnumINFO, "Test completed, "+testConductor.Config.Test)
}

func SendEventTestCompletedFail(testConductor *tc.TestConductor, text string) error {
	return SendEvent(testConductor, td.EventClassEnumFAIL, "Test failed:, "+text)
}

func SendEvent(testConductor *tc.TestConductor, eventClass td.EventClassEnum, message string) error {
	if testConductor.Config.SendEvent == 1 {
		if err := common.SendEvent(
			testConductor.TestDirectorClient,
			testConductor.TestRunId,
			message,
			eventClass,
			td.EventSourceClassEnumTestDashConductor); err != nil {
			return fmt.Errorf("failed to inform test director of event, error: %v", err)
		}
	}
	return nil
}
