package xray

import (
	"fmt"
	"github.com/tidwall/gjson"
	"strings"
	"test-director/models"
)

func CreateTestExecution(testPlanId, testId, testJiraKey string) string {
	s := fmt.Sprintf(
		`mutation{createTestExecution(testIssueIds:["%s"] jira: {fields: {summary: "Test Execution for %s", project: {key: "%s"}}}) {testExecution {issueId jira(fields: ["key"])} warnings createdTestEnvironments}}`,
		testId,
		testJiraKey,
		strings.Split(testJiraKey, "-")[0],
	)
	json := sendXrayQuery(s)
	testExecId := gjson.Get(json, "data.createTestExecution.testExecution.issueId").Str
	addTestExecutionToTestPlan(testPlanId, testExecId)
	return testExecId
}

func DeleteTestExecution(run models.TestRun) {
	s := fmt.Sprintf(`mutation{deleteTestExecution(issueId: "%s")}`, run.TestExecIssueID)
	sendXrayQuery(s)
}

func addTestExecutionToTestPlan(testPlanId, testExecutionId string) {
	s := fmt.Sprintf(`mutation{addTestExecutionsToTestPlan(issueId: "%s", testExecIssueIds: ["%s"]) {addedTestExecutions warning}}`, testPlanId, testExecutionId)
	sendXrayQuery(s)
}
