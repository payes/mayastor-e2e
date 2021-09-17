package handlers

import (

	"github.com/go-openapi/runtime/middleware"
	"test-director/models"
	"test-director/restapi/operations/test_director"
)

type getTestPlansImpl struct{}

func NewGetTestPlansHandler() test_director.GetTestPlansHandler {
	return &getTestPlansImpl{}
}

func (impl *getTestPlansImpl) Handle(test_director.GetTestPlansParams) middleware.Responder {
	plans := planInterface.GetAll()
	if plans == nil {
		return test_director.NewGetTestPlansNotFound()
	}
	return test_director.NewGetTestPlansOK().WithPayload(plans)
}

type getTestPlanImpl struct {}

func NewGetTestPlanByIdHandler() test_director.GetTestPlanByIDHandler {
	return &getTestPlanImpl{}
}

func (impl *getTestPlanImpl) Handle(params test_director.GetTestPlanByIDParams) middleware.Responder {
	id := params.ID
	plan, _  := planInterface.Get(id)
	if plan == nil {
		return test_director.NewGetTestPlanByIDNotFound()
	}
	return test_director.NewGetTestPlanByIDOK().WithPayload(plan)
}

type deleteTestPlanImpl struct {}

func NewDeleteTestPlanByIdHandler() test_director.DeleteTestPlanByIDHandler {
	return &deleteTestPlanImpl{}
}

func (impl *deleteTestPlanImpl) Handle(params test_director.DeleteTestPlanByIDParams) middleware.Responder {
	id := params.ID
	err  := planInterface.Delete(id)
	if err != nil {
		return test_director.NewDeleteTestPlanByIDNotFound()
	}
	return test_director.NewDeleteTestPlanByIDOK()
}

type deleteTestPlansImpl struct {}

func NewDeleteTestPlansHandler() test_director.DeleteTestPlansHandler {
	return &deleteTestPlansImpl{}
}

func (impl *deleteTestPlansImpl) Handle(params test_director.DeleteTestPlansParams) middleware.Responder {
	plan  := planInterface.DeleteAll()
	return test_director.NewDeleteTestPlansOK().WithPayload(plan)
}

type putTestPlanImpl struct {}

func NewPutTestPlanHandler() test_director.PutTestPlanByIDHandler {
	return &putTestPlanImpl{}
}

func (impl *putTestPlanImpl) Handle(params test_director.PutTestPlanByIDParams) middleware.Responder {
	id := params.ID
	tps := params.Body
	b := true
	plan := models.TestPlan{
		TestPlanSpec: *tps,
		IsActive: &b,
		Key: models.JiraKey(id),
		Status:   models.TestPlanStatusEnumNOTSTARTED, //missing
	}
	items := planInterface.GetAll()
	for _, item := range items {
		bool := false
		item.IsActive = &bool
		planInterface.Set(string(item.Key), *item)
	}
	err := planInterface.Set(id, plan)
	if err != nil {
		return test_director.NewPutTestPlanByIDBadRequest().WithPayload(&models.RequestOutcome{
			Details:       err.Error(),
			ItemsAffected: nil, //missing
			Result:        models.RequestOutcomeResultREFUSED,
		})
	}
	return test_director.NewPutTestPlanByIDOK().WithPayload(&plan)
}