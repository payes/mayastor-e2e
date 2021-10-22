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
	Get(key models.JiraKey) *models.TestPlan
	GetActive() *models.TestPlan
	GetAll() []*models.TestPlan
	Set(key models.JiraKey, plan models.TestPlan) error
}

type TestPlanCache struct {
	client *cache.Cache
}

func (r *TestPlanCache) Delete(key models.JiraKey) error {
	tp := r.Get(key)
	if tp != nil {
		r.client.Delete(string(key))
		return nil
	}
	return errors.NotFound("Not found")
}

func (r *TestPlanCache) DeleteAll() *models.RequestOutcome {
	var counter int64
	counter = 0
	for _, item := range r.GetAll() {
		if !*item.IsActive {
			r.Delete(item.Key)
			counter++
		}
	}
	ro := models.RequestOutcome{
		Details:       "Deleted all test plans instead of active one",
		ItemsAffected: &counter,
		Result:        models.RequestOutcomeResultOK,
	}
	return &ro
}

func (r *TestPlanCache) Get(key models.JiraKey) *models.TestPlan {
	tp, exist := r.client.Get(string(key))
	if !exist {
		return nil
	}

	var result models.TestPlan
	err := json.Unmarshal(tp.([]byte), &result)
	if err != nil {
		log.Error("Failed to unmarshall test plan record.", err)
		return nil
	}

	return &result
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
			log.Error("Failed to unmarshall test plan records.", err)
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
		log.Error("default test plan is invalid.")
		return
	}
	b := true
	tests, err := xray.GetTestPlan(*jt.Id)
	if err != nil {
		log.Errorf("the default test plan cannot be initialized: %s", err)
		return
	}
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
