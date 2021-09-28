package tests

import (
	"fmt"
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

func SendTestPreparing(testConductor *tc.TestConductor) error {
	message := "Test preparing, " + testConductor.Config.Test + " time: " + getTime()
	err := tc.SendEventInfo(testConductor.TestDirectorClient, message, tc.SourceInstance)
	if err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return err
}

func SendTestStarted(testConductor *tc.TestConductor) error {
	message := "Test started, " + testConductor.Config.Test + " time: " + getTime()
	err := tc.SendEventInfo(testConductor.TestDirectorClient, message, tc.SourceInstance)
	if err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return err
}

func SendTestCompletedOk(testConductor *tc.TestConductor) error {
	message := "Test completed, " + testConductor.Config.Test + " time: " + getTime()
	err := tc.SendEventInfo(testConductor.TestDirectorClient, message, tc.SourceInstance)
	if err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return err
}

func SendTestCompletedFail(testConductor *tc.TestConductor, text string) error {
	message := "Test failed:, " + text + " time: " + getTime()
	err := tc.SendEventFail(testConductor.TestDirectorClient, message, tc.SourceInstance)
	if err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return err
}
