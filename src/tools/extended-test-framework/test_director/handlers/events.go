package handlers

import (

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	"test-director/models"
	"test-director/restapi/operations/test_director"
	"time"
)

type getEventsImpl struct {}

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

type postEventImpl struct {}

func NewAddEventHandler() test_director.AddEventHandler {
	return &postEventImpl{}
}

func (impl *postEventImpl) Handle(params test_director.AddEventParams) middleware.Responder {
	eventSpec := params.Body
	event := models.Event{
		ID: strfmt.UUID(uuid.New().String()), //random uuid or root_request_id inside body
		LoggedDateTime: strfmt.DateTime(time.Now()),             //missing
		EventSpec:      *eventSpec,
	}
	err := eventInterface.Set(event.ID.String(), event)
	if err != nil {
		return test_director.NewPutTestPlanByIDBadRequest().WithPayload(&models.RequestOutcome{
			Details:       err.Error(),
			ItemsAffected: nil,
			Result:        models.RequestOutcomeResultREFUSED,
		})
	}
	return test_director.NewAddEventOK().WithPayload(&event)
}