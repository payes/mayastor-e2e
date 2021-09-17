package handlers

import (
	"encoding/json"
	"test-director/models"

	"github.com/patrickmn/go-cache"
)

var eventInterface EventInterface

type EventInterface interface {
	GetAll() []*models.Event
	Set(key string, data models.Event) error
}

type EventCache struct {
	client *cache.Cache
}

func (r *EventCache) GetAll() []*models.Event {
	items := r.client.Items()
	if len(items) == 0 {
		return nil
	}
	m := make([]*models.Event, 0, len(items))
	for _, val := range items {
		var result models.Event
		err := json.Unmarshal(val.Object.([]byte), &result)
		if err != nil {
			return nil
		}
		m = append(m, &result)
	}
	return m
}

func (r *EventCache) Set(key string, data models.Event) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r.client.Set(key, b, -1)
	return nil
}



func InitEventCache() {
	eventInterface = &EventCache{
		client: cache.New(-1, 0),
	}
}

