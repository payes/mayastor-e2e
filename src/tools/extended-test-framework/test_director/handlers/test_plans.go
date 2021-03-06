package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"
	"test-director/connectors"
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

type getTestPlanImpl struct{}

func NewGetTestPlanByIdHandler() test_director.GetTestPlanByIDHandler {
	return &getTestPlanImpl{}
}

func (impl *getTestPlanImpl) Handle(params test_director.GetTestPlanByIDParams) middleware.Responder {
	id := params.ID
	plan := planInterface.Get(models.JiraKey(id))
	if plan == nil {
		return test_director.NewGetTestPlanByIDNotFound()
	}
	return test_director.NewGetTestPlanByIDOK().WithPayload(plan)
}

type deleteTestPlanImpl struct{}

func NewDeleteTestPlanByIdHandler() test_director.DeleteTestPlanByIDHandler {
	return &deleteTestPlanImpl{}
}

func (impl *deleteTestPlanImpl) Handle(params test_director.DeleteTestPlanByIDParams) middleware.Responder {
	id := params.ID
	err := planInterface.Delete(models.JiraKey(id))
	if err != nil {
		return test_director.NewDeleteTestPlanByIDNotFound()
	}
	return test_director.NewDeleteTestPlanByIDOK()
}

type deleteTestPlansImpl struct{}

func NewDeleteTestPlansHandler() test_director.DeleteTestPlansHandler {
	return &deleteTestPlansImpl{}
}

func (impl *deleteTestPlansImpl) Handle(params test_director.DeleteTestPlansParams) middleware.Responder {
	plan := planInterface.DeleteAll()
	return test_director.NewDeleteTestPlansOK().WithPayload(plan)
}

type putTestPlanImpl struct{}

func NewPutTestPlanHandler() test_director.PutTestPlanByIDHandler {
	return &putTestPlanImpl{}
}

func (impl *putTestPlanImpl) Handle(params test_director.PutTestPlanByIDParams) middleware.Responder {
	id := params.ID
	tps := params.Body
	plan := planInterface.Get(models.JiraKey(id))
	b := true
	if plan != nil {
		if *plan.Status == models.TestPlanStatusEnumNOTSTARTED && *tps.Status == models.TestPlanStatusEnumRUNNING {
			plan.Status = tps.Status
			plan.IsActive = &b
		} else if *plan.Status == models.TestPlanStatusEnumRUNNING && (*tps.Status == models.TestPlanStatusEnumCOMPLETEFAIL || *tps.Status == models.TestPlanStatusEnumCOMPLETEPASS) {
			plan.Status = tps.Status
			plan.IsActive = &b
		}
	} else {
		jt, err := connectors.GetJiraTaskDetails(id)
		if err != nil {
			log.Errorf("can't fetch jira task detail id: %s err: %s", id, err)
			return testPlanBadRequest(err)
		}
		if jt.Fields.IssueType.Name != "Test Plan" {
			log.Errorf("Jira key: %s issue type: %s is not a Test Plan", jt.Key, jt.Fields.IssueType.Name)
			return testPlanBadRequest(err)
		}

		plan = &models.TestPlan{
			IsActive: &b,
			Key:      models.JiraKey(id),
			TestPlanSpec: models.TestPlanSpec{
				JiraID:   *jt.Id,
				Assignee: jt.Fields.Assignee.Name,
				Name:     *jt.Fields.Name,
				Status:   tps.Status,
			},
		}
	}
	if *plan.IsActive {
		items := planInterface.GetAll()
		for _, item := range items {
			bool := false
			item.IsActive = &bool
			planInterface.Set(item.Key, *item)
		}
	}
	err := planInterface.Set(models.JiraKey(id), *plan)
	if err != nil {
		return testPlanBadRequest(err)
	}
	return test_director.NewPutTestPlanByIDOK().WithPayload(plan)
}

func testPlanBadRequest(err error) middleware.Responder {
	i := int64(1)
	return test_director.NewPutTestPlanByIDBadRequest().WithPayload(&models.RequestOutcome{
		Details:       err.Error(),
		ItemsAffected: &i,
		Result:        models.RequestOutcomeResultREFUSED,
	})
}
