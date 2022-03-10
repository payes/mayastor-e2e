package xray

import (
	"errors"
	"fmt"
	"test-director/models"

	"github.com/tidwall/gjson"
)

func GetTestPlan(testPlanId string) ([]*models.Test, error) {
	s := fmt.Sprintf(`{getTestPlan(issueId: "%s") {tests(limit: 50) {results {issueId}}}}`, testPlanId)
	json, err := sendXrayQuery(s)
	if err != nil {
		return nil, err
	}
	if json == "" {
		return nil, errors.New("unable to fetch xray data")
	}
	res := gjson.Get(json, "data.getTestPlan.tests.results.#.issueId")
	m := make([]*models.Test, 0, len(res.Array()))
	for _, id := range res.Array() {
		idStr := id.String()
		test := models.Test{
			IssueID: &idStr,
		}
		m = append(m, &test)
	}
	return m, nil
}
