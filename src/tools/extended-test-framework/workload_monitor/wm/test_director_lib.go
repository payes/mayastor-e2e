package wm

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/client"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/client/test_director"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/models"
)

func SendEvent(client *client.Etfw, message string, pod string, sourceInstanceUid *strfmt.UUID) error {

	var class = models.EventClassEnumFAIL
	var sourceClass = models.EventSourceClassEnumWorkloadDashMonitor
	var sourceInstance = ""
	if sourceInstanceUid != nil {
		sourceInstance = string(*sourceInstanceUid)
	}

	eventSpec := models.EventSpec{}
	eventSpec.Class = &class
	eventSpec.Data = []string{""}
	eventSpec.Message = &message
	eventSpec.Resource = pod
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
