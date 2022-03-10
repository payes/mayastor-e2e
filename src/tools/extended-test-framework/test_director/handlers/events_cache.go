package handlers

import (
	"encoding/json"
	"os"
	"strings"
	"test-director/config"
	"test-director/connectors"
	"test-director/models"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/patrickmn/go-cache"
)

const SlackChannel = "#test_director"

var slackWebHook = os.Getenv("SLACK_WEB_HOOK")
var eventInterface EventInterface

type EventInterface interface {
	GetAll() []*models.Event
	Set(key string, data models.Event) error
}

type EventCache struct {
	client *cache.Cache
	cfg    *config.ServerConfig
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
			log.Error("Failed to unmarshall event records.", err)
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

	var tr *models.TestRun
	if *data.Class != models.EventClassEnumINFO {
		tr = runInterface.Get(*data.SourceInstance)
		if tr.Data != "" {
			tr.Data += ": "
		}
		tr.Data += *data.Message + " " + strings.Join(data.Data, ", ")
		err = runInterface.Set(*data.SourceInstance, *tr)
		if err != nil {
			return err
		}
	}
	if *data.Class == models.EventClassEnumFAIL && *data.SourceClass != models.EventSourceClassEnumLogDashMonitor {
		tr.Status = models.TestRunStatusEnumFAILED
		if err = UpdateTestRun(tr); err != nil {
			return err
		}
	}
	// -1 means that the item never expires
	r.client.Set(key, b, -1)
	if r.cfg.SlackNotification {
		sendSlackNotification(&data.EventSpec)
	}
	return nil
}

func InitEventCache(cfg *config.ServerConfig) {
	eventInterface = &EventCache{
		client: cache.New(-1, 0),
		cfg:    cfg,
	}
}

func sendSlackNotification(data *models.EventSpec) {
	sc := connectors.SlackClient{
		WebHookUrl: slackWebHook,
		UserName:   string(*data.SourceClass),
		Channel:    SlackChannel,
		TimeOut:    10 * time.Second,
	}
	sn := connectors.SlackJobNotification{
		Details:   *data.Message + " " + strings.Join(data.Data, ", "),
		IconEmoji: ":ghost:",
		Text:      string(*data.Class) + " - TestRun ID: " + *data.SourceInstance,
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
		log.Error(err.Error())
	}
}
