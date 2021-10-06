package handlers

import (
	"errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"test-director/connectors"
	"test-director/connectors/xray"
	"test-director/models"
	"test-director/restapi/operations/test_director"
	"time"
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

type getTestRunImpl struct{}

func NewGetTestRunByIdHandler() test_director.GetTestRunByIDHandler {
	return &getTestRunImpl{}
}

func (impl *getTestRunImpl) Handle(params test_director.GetTestRunByIDParams) middleware.Responder {
	id := params.ID
	run, _ := runInterface.Get(id)
	if run == nil {
		return test_director.NewGetTestRunByIDNotFound()
	}
	return test_director.NewGetTestRunByIDOK().WithPayload(run)
}

type deleteTestRunImpl struct{}

func NewDeleteTestRunByIdHandler() test_director.DeleteTestRunByIDHandler {
	return &deleteTestRunImpl{}
}

func (impl *deleteTestRunImpl) Handle(params test_director.DeleteTestRunByIDParams) middleware.Responder {
	id := params.ID
	err := runInterface.Delete(id)
	if err != nil {
		return test_director.NewDeleteTestRunByIDNotFound()
	}
	return test_director.NewDeleteTestRunByIDOK()
}

type putTestRunImpl struct{}

func NewPutTestRunHandler() test_director.PutTestRunByIDHandler {
	return &putTestRunImpl{}
}

func (impl *putTestRunImpl) Handle(params test_director.PutTestRunByIDParams) middleware.Responder {
	id := params.ID
	testRunSpec := params.Body
	testRun, _ := runInterface.Get(id)
	if testRun != nil {
		if testRun.Status == models.TestRunStatusEnumTODO && testRunSpec.Status == models.TestRunStatusEnumEXECUTING {
			testRun.StartDateTime = strfmt.DateTime(time.Now())
			testRun.TestRunSpec.Data = testRunSpec.Data
			testRun.Status = testRunSpec.Status
			xray.UpdateTestRun(*testRun)
			tp, _ := planInterface.Get(*testRun.TestKey)
			if tp != nil {
				*tp.Status = models.TestPlanStatusEnumRUNNING
				planInterface.Set(tp.Key, *tp)
			}
		}
		if testRun.Status == models.TestRunStatusEnumEXECUTING && (testRunSpec.Status == models.TestRunStatusEnumFAILED || testRunSpec.Status == models.TestRunStatusEnumPASSED) {
			testRun.EndDateTime = strfmt.DateTime(time.Now())
			testRun.TestRunSpec.Data = testRunSpec.Data
			testRun.Status = testRunSpec.Status
			xray.UpdateTestRun(*testRun)
			tp, _ := planInterface.Get(*testRun.TestKey)
			if tp != nil {
				if testRun.Status == models.TestRunStatusEnumPASSED && *tp.Status != models.TestPlanStatusEnumCOMPLETEFAIL {
					*tp.Status = models.TestPlanStatusEnumCOMPLETEPASS
				} else {
					*tp.Status = models.TestPlanStatusEnumCOMPLETEFAIL
				}
				planInterface.Set(tp.Key, *tp)
			}
		}
	} else {
		jt, err := connectors.GetJiraTaskDetails(string(*testRunSpec.TestKey))
		if err != nil {
			return badRequestResponse(err)
		}

		if jt.Fields.IssueType.Name != "Test" {
			return badRequestResponse(err)
		}

		tp := planInterface.GetActive()
		if !contains(tp.Tests, jt.Id) {
			return badRequestResponse(errors.New("test doesn't belong to active test plan"))
		}

		testRun = &models.TestRun{
			ID: id,
			TestRunSpec: models.TestRunSpec{
				Data:            testRunSpec.Data,
				Status:          models.TestRunStatusEnumTODO,
				TestExecIssueID: xray.CreateTestExecution(tp.JiraID, *jt.Id, jt.Key),
				TestID:          *jt.Id,
				TestKey:         testRunSpec.TestKey,
			},
		}
	}
	err := runInterface.Set(id, *testRun)
	if err != nil {
		return badRequestResponse(err)
	}

	return test_director.NewPutTestRunByIDOK().WithPayload(testRun)
}

func badRequestResponse(err error) middleware.Responder {
	i := int64(1)
	return test_director.NewPutTestRunByIDBadRequest().WithPayload(&models.RequestOutcome{
		Details:       err.Error(),
		ItemsAffected: &i,
		Result:        models.RequestOutcomeResultREFUSED,
	})
}

func contains(tests []*models.Test, testId *string) bool {
	for _, test := range tests {
		if *test.IssueID == *testId {
			return true
		}
	}
	return false
}
