package handlers

import (
	"encoding/json"
	"github.com/go-openapi/errors"
	"github.com/patrickmn/go-cache"
	"test-director/models"
)

var planInterface testPlanInterface

type testPlanInterface interface {
	Delete(key string) error
	DeleteAll() *models.RequestOutcome
	Get(key string) (*models.TestPlan, error)
	GetAll() []*models.TestPlan
	Set(key string, data models.TestPlan) error
}

type TestPlanCache struct {
	client *cache.Cache
}
func (r *TestPlanCache) Delete(key string) error {
	tp, _ := r.Get(key)
	if tp != nil{
		r.client.Delete(key)
		return nil
	}
	return errors.NotFound("Not found")
}

func (r *TestPlanCache) DeleteAll() *models.RequestOutcome {
	tps := r.GetAll()
	i64 := int64(len(tps))
	r.client.Flush()
	ro := models.RequestOutcome{
		Details:       "Deleted all test plans",
		ItemsAffected: &i64,
		Result:        models.RequestOutcomeResultOK,
	}
	return &ro
}

func (r *TestPlanCache) Get(key string) (*models.TestPlan, error) {
	tp, exist := r.client.Get(key)
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

func (r *TestPlanCache) Set(key string, data models.TestPlan) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r.client.Set(key, b, -1)
	return nil
}



func InitTestPlanCache() {
	planInterface = &TestPlanCache{
		client: cache.New(-1, 0),
	}
}
