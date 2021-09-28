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

func sendTestPreparing(testConductor *tc.TestConductor) error {
	message := "Test preparing, " + testConductor.Config.Test + " time: " + time.Now().String()
	err := tc.SendEventInfo(testConductor.TestDirectorClient, message, tc.SourceInstance)
	if err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return err
}

func sendTestStarted(testConductor *tc.TestConductor) error {
	message := "Test started, " + testConductor.Config.Test + " time: " + time.Now().String()
	err := tc.SendEventInfo(testConductor.TestDirectorClient, message, tc.SourceInstance)
	if err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return err
}

func sendTestCompletedOk(testConductor *tc.TestConductor) error {
	message := "Test completed, " + testConductor.Config.Test + " time: " + time.Now().String()
	err := tc.SendEventInfo(testConductor.TestDirectorClient, message, tc.SourceInstance)
	if err != nil {
		return fmt.Errorf("failed to inform test director of event, error: %v", err)
	}
	return err
}
