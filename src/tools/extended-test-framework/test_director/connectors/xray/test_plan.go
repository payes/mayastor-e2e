package xray

import (
	"fmt"
	"github.com/tidwall/gjson"
	"test-director/models"
)

func GetTestPlan(testPlanId string) []*models.Test {
	s := fmt.Sprintf(`{getTestPlan(issueId: "%s") {tests(limit: 20) {results {issueId}}}}`, testPlanId)
	json := sendQuery(s)
	res := gjson.Get(json, "data.getTestPlan.tests.results.#.issueId")
	m := make([]*models.Test, 0, len(res.Array()))
	for _, id := range res.Array() {
		test := models.Test{
			Description: "",
			IssueID:     &id.Str,
		}

		m = append(m, &test)
	}
	return m
}
