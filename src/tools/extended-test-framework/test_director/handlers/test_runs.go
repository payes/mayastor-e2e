package handlers

import (

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"test-director/models"
	"test-director/restapi/operations/test_director"
)

type getTestRunsImpl struct{}

func NewGetTestRunsHandler() test_director.GetTestRunsHandler {
	return &getTestRunsImpl{}
}

func (impl *getTestRunsImpl) Handle(test_director.GetTestRunsParams) middleware.Responder {
	runs := runInterface.GetAll()
	if runs == nil {
		return test_director.NewGetTestRunsNotFound()
	}
	return test_director.NewGetTestRunsOK().WithPayload(runs)
}

type getTestRunImpl struct {}

func NewGetTestRunByIdHandler() test_director.GetTestRunByIDHandler {
	return &getTestRunImpl{}
}

func (impl *getTestRunImpl) Handle(params test_director.GetTestRunByIDParams) middleware.Responder {
	id := params.ID
	run, _  := runInterface.Get(id)
	if run == nil {
		return test_director.NewGetTestRunByIDNotFound()
	}
	return test_director.NewGetTestRunByIDOK().WithPayload(run)
}

type deleteTestRunImpl struct {}

func NewDeleteTestRunByIdHandler() test_director.DeleteTestRunByIDHandler {
	return &deleteTestRunImpl{}
}

func (impl *deleteTestRunImpl) Handle(params test_director.DeleteTestRunByIDParams) middleware.Responder {
	id := params.ID
	err  := runInterface.Delete(id)
	if err != nil {
		return test_director.NewDeleteTestRunByIDNotFound()
	}
	return test_director.NewDeleteTestRunByIDOK()
}

type putTestRunImpl struct {}

func NewPutTestRunHandler() test_director.PutTestRunByIDHandler {
	return &putTestRunImpl{}
}

func (impl *putTestRunImpl) Handle(params test_director.PutTestRunByIDParams) middleware.Responder {
	id := params.ID
	testRunSpec := params.Body
	testRun := models.TestRun{
		EndDateTime:   strfmt.DateTime{}, //missing
		ID:            id,
		StartDateTime: strfmt.DateTime{}, //missing
		TestPlanKey:   "", //missing
		TestRunSpec:   *testRunSpec,
	}
	err := runInterface.Set(id.String(), testRun)
	if err != nil {
		return test_director.NewPutTestRunByIDBadRequest().WithPayload(&models.RequestOutcome{
			Details:       err.Error(),
			ItemsAffected: nil, //missing
			Result:        models.RequestOutcomeResultREFUSED,
		})
	}
	return test_director.NewPutTestRunByIDOK().WithPayload(&testRun)
}