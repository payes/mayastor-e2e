package tests

import (
	"fmt"
	"mayastor-e2e/tools/extended-test-framework/common/custom_resources"

	"mayastor-e2e/tools/extended-test-framework/common"

	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"

	td "mayastor-e2e/tools/extended-test-framework/common/td/models"
	"mayastor-e2e/tools/extended-test-framework/common/wm/models"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"time"

	"github.com/go-openapi/strfmt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const MCP_NEXUS_ONLINE = "NEXUS_ONLINE"
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

func CheckMSVwithGrpc(ms_ips []string, msv_uid string) error {
	state, err := k8sclient.GetVolumeState(ms_ips, msv_uid)
	if err != nil {
		return fmt.Errorf("MSV grpc check failed, err: %v", err)
	}
	if state != MCP_NEXUS_ONLINE {
		return fmt.Errorf("MSV is not healthy, state is %s", state)
	}
	//logf.Log.Info("check MSV", "uid", msv_uid, "state", state)
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

func MonitorCRs(testConductor *tc.TestConductor, msv_uids []string, duration time.Duration, moac bool) string {

	var endTime = time.Now().Add(duration)
	var waitSecs = 5
	for {
		if moac {
			for _, msv := range msv_uids {
				if err := CheckMSVmoac(msv); err != nil {
					return fmt.Sprintf("MSV %s check failed, err: %s\n", msv, err.Error())
				}
			}
		} else {
			ms_ips, err := k8sclient.GetMayastorNodeIPs()
			if err != nil {
				return fmt.Sprintf("MSV grpc check failed to get nodes, err: %s\n", err.Error())
			}
			for _, msv := range msv_uids {
				if err := CheckMSVwithGrpc(ms_ips, msv); err != nil {
					return fmt.Sprintf("MSV grpc %s check failed, err: %s\n", msv, err.Error())
				}
			}
		}
		if err := CheckPools(testConductor.Config.Msnodes); err != nil {
			return fmt.Sprintf("MSP check failed, err: %s\n", err.Error())
		}
		if moac {
			if err := CheckNodes(testConductor.Config.Msnodes); err != nil {
				return fmt.Sprintf("MSN check failed, err: %s\n", err.Error())
			}
		}
		if time.Now().After(endTime) {
			break
		}
		time.Sleep(time.Duration(waitSecs) * time.Second)
	}
	return ""
}

func ReportResult(testConductor *tc.TestConductor, failmessage string, testRunId strfmt.UUID, err error) error {
	if err != nil {
		logf.Log.Info("failed to run test", "error", err)
		if failmessage != "" {
			failmessage = "\n"
		}
		failmessage = failmessage + err.Error()
	}
	if failmessage == "" {
		if err := common.SendTestRunCompletedOk(
			testConductor.TestDirectorClient,
			testRunId,
			"",
			testConductor.Config.Test); err != nil {
			return fmt.Errorf("failed to inform test director of completion, error: %v", err)
		}
		if err := SendEventTestCompletedOk(testConductor, testRunId); err != nil {
			return fmt.Errorf("failed to inform test director of completion event, error: %v", err)
		}
	} else {
		if err := common.SendTestRunCompletedFail(
			testConductor.TestDirectorClient,
			testRunId,
			failmessage,
			testConductor.Config.Test); err != nil {
			return fmt.Errorf("failed to inform test director of completion, error: %v", err)
		}
	}
	return nil
}

func SendEventTestPreparing(testConductor *tc.TestConductor, testUid strfmt.UUID) error {
	message := "Test preparing, " + testConductor.Config.Test
	if err := common.SendEventInfo(testConductor.TestDirectorClient, testUid, message, td.EventSourceClassEnumTestDashConductor); err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return nil
}

func SendEventTestStarted(testConductor *tc.TestConductor, testUid strfmt.UUID) error {
	message := "Test started, " + testConductor.Config.Test
	if err := common.SendEventInfo(testConductor.TestDirectorClient, testUid, message, td.EventSourceClassEnumTestDashConductor); err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return nil
}

func SendEventTestCompletedOk(testConductor *tc.TestConductor, testUid strfmt.UUID) error {
	message := "Test completed, " + testConductor.Config.Test
	if err := common.SendEventInfo(testConductor.TestDirectorClient, testUid, message, td.EventSourceClassEnumTestDashConductor); err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return nil
}

func SendEventTestCompletedFail(testConductor *tc.TestConductor, text string, testUid strfmt.UUID) error {
	message := "Test failed:, " + text
	if err := common.SendEventFail(testConductor.TestDirectorClient, testUid, message, td.EventSourceClassEnumTestDashConductor); err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return nil
}
