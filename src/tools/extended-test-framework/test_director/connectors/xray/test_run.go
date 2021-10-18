package xray

import (
	"fmt"
	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"test-director/models"
	"time"
)

func UpdateTestRun(testRun models.TestRun) {
	testRunId := getTestRun(testRun.TestID, testRun.TestExecIssueID)
	s := fmt.Sprintf(
		`mutation{updateTestRunStatus( id: "%s", status: "%s")}`,
		testRunId,
		testRun.TestRunSpec.Status,
	)
	sendQuery(s)
	s = fmt.Sprintf(
		`mutation{updateTestRun( id: "%s", comment: "%s", startedOn: "%s", finishedOn: "%s") {warnings}}`,
		testRunId,
		testRun.Data,
		getRFC3339Format(testRun.StartDateTime),
		getRFC3339Format(testRun.EndDateTime),
	)
	sendQuery(s)
}

func getTestRun(testId, testExecutionId string) string {
	s := fmt.Sprintf(
		`{getTestRun( testIssueId: "%s", testExecIssueId: "%s") {id status {name color description}}}`,
		testId,
		testExecutionId,
	)
	json := sendQuery(s)
	return gjson.Get(json, "data.getTestRun.id").Str
}

func getRFC3339Format(t strfmt.DateTime) string {
	s, err := time.Parse(time.RFC3339, t.String())
	if err != nil {
		log.Error(err)
	}
	if s.IsZero() {
		return ""
	}
	return s.Format(time.RFC3339)
}
