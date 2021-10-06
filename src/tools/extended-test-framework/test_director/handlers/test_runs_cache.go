package handlers

import (
	"encoding/json"
	"github.com/go-openapi/errors"
	"github.com/patrickmn/go-cache"
	"test-director/models"
)

var runInterface TestRunInterface

type TestRunInterface interface {
	Delete(key string) error
	Get(key string) (*models.TestRun, error)
	GetAll() []*models.TestRun
	Set(key string, data models.TestRun) error
}

type TestRunCache struct {
	client *cache.Cache
}

func (r *TestRunCache) Delete(key string) error {
	tp, _ := r.Get(key)
	if tp != nil {
		r.client.Delete(key)
		return nil
	}
	return errors.NotFound("Not found")
}

func (r *TestRunCache) Get(key string) (*models.TestRun, error) {
	tp, exist := r.client.Get(key)
	if !exist {
		return nil, nil
	}

	var result models.TestRun
	err := json.Unmarshal(tp.([]byte), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *TestRunCache) GetAll() []*models.TestRun {
	items := r.client.Items()
	if len(items) == 0 {
		return nil
	}
	m := make([]*models.TestRun, 0, len(items))
	for _, val := range items {
		var result models.TestRun
		err := json.Unmarshal(val.Object.([]byte), &result)
		if err != nil {
			return nil
		}
		m = append(m, &result)
	}
	return m
}

func (r *TestRunCache) Set(key string, data models.TestRun) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r.client.Set(key, b, -1)
	return nil
}

func InitTestRunCache() {
	runInterface = &TestRunCache{
		client: cache.New(-1, 0),
	}
}
