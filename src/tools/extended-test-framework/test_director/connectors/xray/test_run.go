package xray

import (
	"fmt"
	"test-director/models"
	"time"

	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func UpdateTestRun(testRun models.TestRun) error {
	var err error
	testRunId, err := getTestRun(testRun.TestID, testRun.TestExecIssueID)
	if err != nil {
		return err
	}
	s := fmt.Sprintf(
		`mutation{updateTestRunStatus( id: "%s", status: "%s")}`,
		testRunId,
		testRun.TestRunSpec.Status,
	)
	if _, err = sendXrayQuery(s); err != nil {
		return err
	}
	s = fmt.Sprintf(
		`mutation{updateTestRun( id: "%s", comment: "%s", startedOn: "%s", finishedOn: "%s") {warnings}}`,
		testRunId,
		testRun.Data,
		getRFC3339Format(testRun.StartDateTime),
		getRFC3339Format(testRun.EndDateTime),
	)
	_, err = sendXrayQuery(s)
	return err
}

func getTestRun(testId, testExecutionId string) (string, error) {
	s := fmt.Sprintf(
		`{getTestRun( testIssueId: "%s", testExecIssueId: "%s") {id status {name color description}}}`,
		testId,
		testExecutionId,
	)
	json, err := sendXrayQuery(s)
	if err != nil {
		return "", err
	}
	return gjson.Get(json, "data.getTestRun.id").Str, nil
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
