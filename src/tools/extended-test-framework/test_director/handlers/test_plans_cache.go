package handlers

import (
	"encoding/json"
	"github.com/go-openapi/errors"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"test-director/config"
	"test-director/connectors"
	"test-director/connectors/xray"
	"test-director/models"
)

var planInterface testPlanInterface

type testPlanInterface interface {
	Delete(key models.JiraKey) error
	DeleteAll() *models.RequestOutcome
	Get(key models.JiraKey) (*models.TestPlan, error)
	GetActive() *models.TestPlan
	GetAll() []*models.TestPlan
	Set(key models.JiraKey, plan models.TestPlan) error
}

type TestPlanCache struct {
	client *cache.Cache
}

func (r *TestPlanCache) Delete(key models.JiraKey) error {
	tp, _ := r.Get(key)
	if tp != nil {
		r.client.Delete(string(key))
		return nil
	}
	return errors.NotFound("Not found")
}

func (r *TestPlanCache) DeleteAll() *models.RequestOutcome {
	i64 := int64(r.client.ItemCount() - 1)
	for _, item := range r.GetAll() {
		if !*item.IsActive {
			r.Delete(item.Key)
		}
	}
	ro := models.RequestOutcome{
		Details:       "Deleted all test plans instead of active one",
		ItemsAffected: &i64,
		Result:        models.RequestOutcomeResultOK,
	}
	return &ro
}

func (r *TestPlanCache) Get(key models.JiraKey) (*models.TestPlan, error) {
	tp, exist := r.client.Get(string(key))
	if !exist {
		return nil, nil
	}

	var result models.TestPlan
	err := json.Unmarshal(tp.([]byte), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *TestPlanCache) GetActive() *models.TestPlan {
	items := r.client.Items()
	if len(items) == 0 {
		return nil
	}
	for _, val := range items {
		var result models.TestPlan
		err := json.Unmarshal(val.Object.([]byte), &result)
		if err != nil {
			return nil
		}
		if *result.IsActive {
			return &result
		}
	}
	return nil
}

func (r *TestPlanCache) GetAll() []*models.TestPlan {
	items := r.client.Items()
	if len(items) == 0 {
		return nil
	}
	m := make([]*models.TestPlan, 0, len(items))
	for _, val := range items {
		var result models.TestPlan
		err := json.Unmarshal(val.Object.([]byte), &result)
		if err != nil {
			return nil
		}
		m = append(m, &result)
	}
	return m
}

func (r *TestPlanCache) Set(key models.JiraKey, plan models.TestPlan) error {
	b, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	r.client.Set(string(key), b, -1)
	return nil
}

func InitTestPlanCache(dtp *config.ServerConfig) {
	planInterface = &TestPlanCache{
		client: cache.New(-1, 0),
	}
	jt, err := connectors.GetJiraTaskDetails(dtp.DefaultTestPlan)
	if err != nil {
		log.Error("default test plan is invalid")
		return
	}
	b := true
	tests := xray.GetTestPlan(*jt.Id)
	plan := &models.TestPlan{
		IsActive: &b,
		Key:      models.JiraKey(dtp.DefaultTestPlan),
		TestPlanSpec: models.TestPlanSpec{
			JiraID:   *jt.Id,
			Assignee: jt.Fields.Assignee.Name,
			Name:     *jt.Fields.Name,
			Status:   models.NewTestPlanStatusEnum("NOT_STARTED"),
			Tests:    tests,
		},
	}
	planInterface.Set(models.JiraKey(dtp.DefaultTestPlan), *plan)
}
