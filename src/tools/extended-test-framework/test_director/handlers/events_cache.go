package handlers

import (
	"encoding/json"
	"fmt"
	"test-director/connectors"
	"test-director/models"
	"time"

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
	setupSlackNotification(&data.EventSpec)
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

func setupSlackNotification(data *models.EventSpec) {
	sc := connectors.SlackClient{
		WebHookUrl: "https://hooks.slack.com/services/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",
		UserName:   string(*data.SourceClass),
		Channel:    "#test_director",
		TimeOut:    10 * time.Second,
	}
	sn := connectors.SlackJobNotification{
		Details:   *data.Message,
		IconEmoji: ":ghost:",
		Text:      string(*data.Class) + " - " + *data.SourceInstance,
	}
	switch *data.Class {
	case models.EventClassEnumFAIL:
		sn.Color = "danger"
	case models.EventClassEnumWARN:
		sn.Color = "warning"
	default:
		sn.Color = "good"
	}
	err := sc.SendJobNotification(sn)
	if err != nil {
		fmt.Print(err.Error())
	}
}