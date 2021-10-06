package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"test-director/models"
	"test-director/restapi/operations/test_director"
	"time"
)

type getEventsImpl struct{}

func NewGetEventsHandler() test_director.GetEventsHandler {
	return &getEventsImpl{}
}

func (impl *getEventsImpl) Handle(test_director.GetEventsParams) middleware.Responder {
	events := eventInterface.GetAll()
	if events == nil {
		return test_director.NewGetEventsNotFound()
	}
	return test_director.NewGetEventsOK().WithPayload(events)
}

type postEventImpl struct{}

func NewAddEventHandler() test_director.AddEventHandler {
	return &postEventImpl{}
}

func (impl *postEventImpl) Handle(params test_director.AddEventParams) middleware.Responder {
	eventSpec := params.Body
	event := models.Event{
		ID:             strfmt.UUID(uuid.New().String()),
		LoggedDateTime: strfmt.DateTime(time.Now()),
		EventSpec:      *eventSpec,
	}
	err := eventInterface.Set(event.ID.String(), event)
	if err != nil {
		i := int64(1)
		return test_director.NewPutTestPlanByIDBadRequest().WithPayload(&models.RequestOutcome{
			Details:       err.Error(),
			ItemsAffected: &i,
			Result:        models.RequestOutcomeResultREFUSED,
		})
	}
	return test_director.NewAddEventOK().WithPayload(&event)
}
