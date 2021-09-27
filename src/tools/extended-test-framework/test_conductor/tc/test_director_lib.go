package tc

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/test_conductor/td/client"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/td/client/test_director"

	"mayastor-e2e/tools/extended-test-framework/test_conductor/td/models"
)

func GetTestPlans(client *client.Etfw) error {
	testPlanParams := test_director.NewGetTestPlansParams()
	pTestPlansOk, err := client.TestDirector.GetTestPlans(testPlanParams)

	if err != nil {
		return fmt.Errorf("failed to get plans %v %v\n", err, pTestPlansOk)
	} else {
		logf.Log.Info("got test plans", "count", len(pTestPlansOk.Payload))
		for _, tp := range pTestPlansOk.Payload {
			logf.Log.Info("plan", "name", tp.Name, "key", tp.Key)
		}
	}
	return nil
}

func SendTestPlan(client *client.Etfw, name string, jira_key string, status models.TestPlanStatusEnum) error {
	var tpname = "test"

	testPlanSpec := models.TestPlanSpec{}
	testPlanSpec.Name = tpname
	testPlanSpec.Status = &status

	testPlanParams := test_director.NewPutTestPlanByIDParams()
	testPlanParams.ID = jira_key
	testPlanParams.Body = &testPlanSpec

	pPutTestPlansOk, err := client.TestDirector.PutTestPlanByID(testPlanParams)

	if err != nil {
		return fmt.Errorf("failed to put plans, error: %v %v", err, pPutTestPlansOk)
	} else {
		logf.Log.Info("put plan",
			"name", pPutTestPlansOk.Payload.Name,
			"payload", pPutTestPlansOk.Payload,
			"plan ID", pPutTestPlansOk.Payload.Key)
	}
	return nil
}

func SendTestPlanRunning(client *client.Etfw, name string, jira_key string) error {
	return SendTestPlan(client, name, jira_key, models.TestPlanStatusEnumRUNNING)
}

func SendRunStatus(client *client.Etfw, uuid string, message string, jira_key_str string, status models.TestRunStatusEnum) error {

	var jira_key = models.JiraKey(jira_key_str)
	testRunSpec := models.TestRunSpec{}
	testRunSpec.Data = message
	testRunSpec.Status = status
	testRunSpec.TestKey = &jira_key

	params := test_director.NewPutTestRunByIDParams()
	params.Body = &testRunSpec
	params.ID = strfmt.UUID(uuid)

	pRunStatusOk, err := client.TestDirector.PutTestRunByID(params)

	if err != nil {
		return fmt.Errorf("failed to put event, error: %v %v\n", err, pRunStatusOk)
	} else {
		logf.Log.Info("put test run",
			"message", pRunStatusOk.Payload.Data,
			"status", pRunStatusOk.Payload.Status,
			"key", pRunStatusOk.Payload.TestKey)
	}
	return nil
}

func SendRunCompletedOk(client *client.Etfw, uuid string, message string, jira_key string) error {
	return SendRunStatus(client, uuid, message, jira_key, models.TestRunStatusEnumCOMPLETEPASS)
}

func SendRunCompletedFail(client *client.Etfw, uuid string, message string, jira_key string) error {
	return SendRunStatus(client, uuid, message, jira_key, models.TestRunStatusEnumCOMPLETEFAIL)
}

func SendRunStarted(client *client.Etfw, uuid string, message string, jira_key string) error {
	return SendRunStatus(client, uuid, message, jira_key, models.TestRunStatusEnumRUNNING)
}

func SendEvent(client *client.Etfw, message string, source string, eventClass models.EventClassEnum) error {

	var class = models.EventClassEnumFAIL
	var sourceClass = models.EventSourceClassEnumWorkloadDashMonitor
	var sourceInstance = source

	eventSpec := models.EventSpec{}
	eventSpec.Class = &class
	eventSpec.Data = []string{""}
	eventSpec.Message = &message
	eventSpec.Resource = ""
	eventSpec.SourceClass = &sourceClass
	eventSpec.SourceInstance = &sourceInstance

	params := test_director.NewAddEventParams()
	params.Body = &eventSpec

	pAddEventOk, err := client.TestDirector.AddEvent(params)

	if err != nil {
		return fmt.Errorf("failed to put event, error: %v %v\n", err, pAddEventOk)
	} else {
		logf.Log.Info("put event",
			"data", pAddEventOk.Payload.Data[0],
			"message", *pAddEventOk.Payload.Message,
			"resource", pAddEventOk.Payload.Resource)
	}
	return nil
}

func SendEventFail(client *client.Etfw, message string, source string) error {
	return SendEvent(client, message, source, models.EventClassEnumFAIL)
}

func SendEventInfo(client *client.Etfw, message string, source string) error {
	return SendEvent(client, message, source, models.EventClassEnumINFO)
}

func SendEventWarn(client *client.Etfw, message string, source string) error {
	return SendEvent(client, message, source, models.EventClassEnumWARN)
}
