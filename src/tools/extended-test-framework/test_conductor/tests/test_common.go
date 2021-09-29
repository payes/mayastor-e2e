package tests

import (
	"fmt"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/lib"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/wm/models"
	"time"
)

var violations = []models.WorkloadViolationEnum{
	models.WorkloadViolationEnumRESTARTED,
	models.WorkloadViolationEnumTERMINATED,
	models.WorkloadViolationEnumNOTPRESENT,
}

func getTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func CheckMSV(msv_uid string) error {
	msv, err := lib.GetMSV(msv_uid)
	if err != nil {
		return fmt.Errorf("failed to get the MSV, error: %v", err)
	}
	if msv == nil {
		return fmt.Errorf("failed to get the MSV, MSV is nil")
	}
	if msv.Status.State != "healthy" {
		return fmt.Errorf("MSV is not healthy, state is %s", msv.Status.State)
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

func SendTestPreparing(testConductor *tc.TestConductor) error {
	message := "Test preparing, " + testConductor.Config.Test + ", time: " + getTime()
	if err := tc.SendEventInfo(testConductor.TestDirectorClient, message, tc.SourceInstance); err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return nil
}

func SendTestStarted(testConductor *tc.TestConductor) error {
	message := "Test started, " + testConductor.Config.Test + ", time: " + getTime()
	if err := tc.SendEventInfo(testConductor.TestDirectorClient, message, tc.SourceInstance); err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return nil
}

func SendTestCompletedOk(testConductor *tc.TestConductor) error {
	message := "Test completed, " + testConductor.Config.Test + ", time: " + getTime()
	if err := tc.SendEventInfo(testConductor.TestDirectorClient, message, tc.SourceInstance); err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return nil
}

func SendTestCompletedFail(testConductor *tc.TestConductor, text string) error {
	message := "Test failed:, " + text + ", time: " + getTime()
	if err := tc.SendEventFail(testConductor.TestDirectorClient, message, tc.SourceInstance); err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return nil
}
