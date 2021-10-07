package common

import (
	"fmt"

	"mayastor-e2e/tools/extended-test-framework/common/td/client"
	"mayastor-e2e/tools/extended-test-framework/common/td/client/test_director"

	"github.com/go-openapi/strfmt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/common/td/models"
)

func sendTestRunStatus(client *client.Etfw, uuid strfmt.UUID, message string, jira_key_str string, status models.TestRunStatusEnum) error {

	var jira_key = models.JiraKey(jira_key_str)
	testRunSpec := models.TestRunSpec{}
	testRunSpec.Data = message
	testRunSpec.Status = status
	testRunSpec.TestKey = &jira_key

	params := test_director.NewPutTestRunByIDParams()
	params.Body = &testRunSpec
	params.ID = string(uuid)

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

func SendTestRunCompletedOk(client *client.Etfw, uuid strfmt.UUID, message string, jira_key string) error {
	logf.Log.Info("SendTestRunCompletedOk")
	return sendTestRunStatus(client, uuid, message, jira_key, models.TestRunStatusEnumPASSED)
}

func SendTestRunCompletedFail(client *client.Etfw, uuid strfmt.UUID, message string, jira_key string) error {
	logf.Log.Info("SendTestRunCompletedFail")
	return sendTestRunStatus(client, uuid, message, jira_key, models.TestRunStatusEnumFAILED)
}

func SendTestRunStarted(client *client.Etfw, uuid strfmt.UUID, message string, jira_key string) error {
	logf.Log.Info("SendTestRunStarted")
	return sendTestRunStatus(client, uuid, message, jira_key, models.TestRunStatusEnumEXECUTING)
}

func SendTestRunToDo(client *client.Etfw, uuid strfmt.UUID, message string, jira_key string) error {
	logf.Log.Info("SendTestRunToDo")
	return sendTestRunStatus(client, uuid, message, jira_key, models.TestRunStatusEnumTODO)
}

func sendEvent(
	client *client.Etfw,
	sourceInstance strfmt.UUID,
	message string,
	eventClass models.EventClassEnum,
	sourceClass models.EventSourceClassEnum) error {

	var sourceInstanceString = string(sourceInstance)

	eventSpec := models.EventSpec{}
	eventSpec.Class = &eventClass
	eventSpec.Data = []string{""}
	eventSpec.Message = &message
	eventSpec.Resource = ""
	eventSpec.SourceClass = &sourceClass
	eventSpec.SourceInstance = &sourceInstanceString

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

func SendEventFail(client *client.Etfw, source strfmt.UUID, message string, sourceClass models.EventSourceClassEnum) error {
	logf.Log.Info("SendEventFail")
	return sendEvent(client, source, message, models.EventClassEnumFAIL, sourceClass)
}

func SendEventInfo(client *client.Etfw, source strfmt.UUID, message string, sourceClass models.EventSourceClassEnum) error {
	logf.Log.Info("SendEventInfo")
	return sendEvent(client, source, message, models.EventClassEnumINFO, sourceClass)
}

func SendEventWarn(client *client.Etfw, source strfmt.UUID, message string, sourceClass models.EventSourceClassEnum) error {
	logf.Log.Info("SendEventWarn")
	return sendEvent(client, source, message, models.EventClassEnumWARN, sourceClass)
}
