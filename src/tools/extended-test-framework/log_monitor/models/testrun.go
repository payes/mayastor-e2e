package models

type TestRun struct {
	EndDateTime   string `json:"endDateTime,omitempty"`
	ID            string `json:"id,omitempty"`
	StartDateTime string `json:"startDateTime,omitempty"`
	Spec          TestRunSpec
}

type TestRunSpec struct {
	Data            string `json:"data,omitempty"`
	Status          string `json:"status,omitempty"`
	TestExecIssueID string `json:"testExecIssueId,omitempty"`
	TestID          string `json:"testId,omitempty"`
	TestKey         string `json:"testKey"`
}
