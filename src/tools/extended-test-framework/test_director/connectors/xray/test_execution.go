package xray

import (
	"fmt"
	"github.com/tidwall/gjson"
)

func CreateTestExecution(testPlanId, testId, testJiraKey string) string {
	s := fmt.Sprintf(`mutation{createTestExecution(testIssueIds:["%s"] jira: {fields: {summary: "Test Execution for %s", project: {key: "ET"}}}) {testExecution {issueId jira(fields: ["key"])} warnings createdTestEnvironments}}`, testId, testJiraKey)
	json := sendQuery(s)
	testExecId := gjson.Get(json, "data.createTestExecution.testExecution.issueId").Str
	addTestExecutionToTestPlan(testPlanId, testExecId)
	return testExecId
}

func addTestExecutionToTestPlan(testPlanId, testExecutionId string) {
	s := fmt.Sprintf(`mutation{addTestExecutionsToTestPlan(issueId: "%s", testExecIssueIds: ["%s"]) {addedTestExecutions warning}}`, testPlanId, testExecutionId)
	sendQuery(s)
}
