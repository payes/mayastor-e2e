package handlers

import (
	"errors"
	"test-director/connectors"
	"test-director/connectors/xray"
	"test-director/models"
	"test-director/restapi/operations/test_director"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
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
	run := runInterface.Get(id)
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
	testRun := runInterface.Get(id)
	if testRun != nil {
		if testRun.Status == models.TestRunStatusEnumTODO && testRunSpec.Status == models.TestRunStatusEnumEXECUTING {
			testRun.StartDateTime = strfmt.DateTime(time.Now())
			testRun.TestRunSpec.Data = testRunSpec.Data
			testRun.Status = testRunSpec.Status
			if err := xray.UpdateTestRun(*testRun); err != nil {
				return badRequestResponse(err)
			}
			tp := planInterface.GetActive()
			if tp == nil || !contains(tp.Tests, &testRun.TestID) {
				return badRequestResponse(errors.New("there is no active test plan belongs to test run"))
			}
			*tp.Status = models.TestPlanStatusEnumRUNNING
			planInterface.Set(tp.Key, *tp)
		} else if testRun.Status != models.TestRunStatusEnumTODO && testRunSpec.Status == models.TestRunStatusEnumFAILED {
			testRun.EndDateTime = strfmt.DateTime(time.Now())
			if testRun.Data != "" {
				testRun.Data += ": "
			}
			testRun.Data += testRunSpec.Data
			testRun.Status = testRunSpec.Status
			if err := UpdateTestRun(testRun); err != nil {
				return badRequestResponse(err)
			}
		} else if testRun.Status != models.TestRunStatusEnumTODO && testRunSpec.Status == models.TestRunStatusEnumPASSED {
			testRun.EndDateTime = strfmt.DateTime(time.Now())
			testRun.TestRunSpec.Data = testRunSpec.Data
			testRun.Status = testRunSpec.Status
			if err := UpdateTestRun(testRun); err != nil {
				return badRequestResponse(err)
			}
		}
	} else {
		jt, err := connectors.GetJiraTaskDetails(string(*testRunSpec.TestKey))
		if err != nil {
			return badRequestResponse(err)
		}

		if jt.Fields.IssueType.Name != "Test" {
			return badRequestResponse(errors.New("test doesn't have issueType 'Test'"))
		}

		tp := planInterface.GetActive()
		if tp == nil {
			return badRequestResponse(errors.New("there is no active test plan"))
		}

		if !contains(tp.Tests, jt.Id) {
			return badRequestResponse(errors.New("test doesn't belong to active test plan"))
		}
		id, err := xray.CreateTestExecution(tp.JiraID, *jt.Id, jt.Key)
		if err != nil {
			return badRequestResponse(err)
		}
		testRun = &models.TestRun{
			ID: id,
			TestRunSpec: models.TestRunSpec{
				Data:            testRunSpec.Data,
				Status:          models.TestRunStatusEnumTODO,
				TestExecIssueID: id,
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

func UpdateTestRun(testRun *models.TestRun) error {
	if err := xray.UpdateTestRun(*testRun); err != nil {
		return err
	}
	tp := planInterface.Get(*testRun.TestKey)
	if tp != nil {
		if testRun.Status == models.TestRunStatusEnumPASSED && *tp.Status != models.TestPlanStatusEnumCOMPLETEFAIL {
			*tp.Status = models.TestPlanStatusEnumCOMPLETEPASS
		} else {
			*tp.Status = models.TestPlanStatusEnumCOMPLETEFAIL
		}
		planInterface.Set(tp.Key, *tp)
	}
	return nil
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
