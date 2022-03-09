package xray

import (
	"fmt"
	"strings"
	"test-director/models"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func CreateTestExecution(testPlanId, testId, testJiraKey string) (string, error) {
	s := fmt.Sprintf(
		`mutation{createTestExecution(testIssueIds:["%s"] jira: {fields: {summary: "Test Execution for %s", project: {key: "%s"}}}) {testExecution {issueId jira(fields: ["key"])} warnings createdTestEnvironments}}`,
		testId,
		testJiraKey,
		strings.Split(testJiraKey, "-")[0],
	)
	json, err := sendXrayQuery(s)
	if err != nil {
		log.Errorf("Failed to send Xray query err:%s\n", err)
		return "", err
	}

	testExecId := gjson.Get(json, "data.createTestExecution.testExecution.issueId").Str
	err = addTestExecutionToTestPlan(testPlanId, testExecId)
	return testExecId, err
}

func DeleteTestExecution(run models.TestRun) error {
	s := fmt.Sprintf(`mutation{deleteTestExecution(issueId: "%s")}`, run.TestExecIssueID)
	_, err := sendXrayQuery(s)
	return err
}

func addTestExecutionToTestPlan(testPlanId, testExecutionId string) error {
	s := fmt.Sprintf(`mutation{addTestExecutionsToTestPlan(issueId: "%s", testExecIssueIds: ["%s"]) {addedTestExecutions warning}}`, testPlanId, testExecutionId)
	_, err := sendXrayQuery(s)
	return err
}
