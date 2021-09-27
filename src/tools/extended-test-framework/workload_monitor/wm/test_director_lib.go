package wm

import (
	"fmt"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/client"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/client/test_director"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/models"
)

func SendEvent(client *client.Etfw, message string, pod string) error {

	var class = models.EventClassEnumFAIL
	var sourceClass = models.EventSourceClassEnumWorkloadDashMonitor
	var sourceInstance = pod
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
