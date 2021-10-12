package xray

import (
	"fmt"
	"github.com/tidwall/gjson"
	"test-director/models"
)

func CreateTestExecution(testPlanId, testId, testJiraKey string) string {
	s := fmt.Sprintf(`mutation{createTestExecution(testIssueIds:["%s"] jira: {fields: {summary: "Test Execution for %s", project: {key: "ET"}}}) {testExecution {issueId jira(fields: ["key"])} warnings createdTestEnvironments}}`, testId, testJiraKey)
	json := sendQuery(s)
	testExecId := gjson.Get(json, "data.createTestExecution.testExecution.issueId").Str
	addTestExecutionToTestPlan(testPlanId, testExecId)
	return testExecId
}

func DeleteTestExecution(run models.TestRun) {
	s := fmt.Sprintf(`mutation{deleteTestExecution(issueId: "%s")}`, run.TestExecIssueID)
	sendQuery(s)
}

func addTestExecutionToTestPlan(testPlanId, testExecutionId string) {
	s := fmt.Sprintf(`mutation{addTestExecutionsToTestPlan(issueId: "%s", testExecIssueIds: ["%s"]) {addedTestExecutions warning}}`, testPlanId, testExecutionId)
	sendQuery(s)
}
