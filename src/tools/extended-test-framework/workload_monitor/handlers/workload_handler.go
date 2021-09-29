package handlers

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/k8sclient"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/list"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/models"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/restapi/operations/workload_monitor"
)

//var WorkloadMap = map[strfmt.UUID]map[strfmt.UUID]*models.Workload{}

type putWorkloadByRegistrantImpl struct{}

func NewPutWorkloadByRegistrantHandler() workload_monitor.PutWorkloadByRegistrantHandler {
	return &putWorkloadByRegistrantImpl{}
}

func (impl *putWorkloadByRegistrantImpl) Handle(params workload_monitor.PutWorkloadByRegistrantParams) middleware.Responder {
	var wl models.Workload
	wl.ID = params.Wid
	wl.Name = ""
	wl.Namespace = ""
	wl.WorkloadSpec = *params.Body

	name, namespace, err := k8sclient.GetPodNameAndNamespaceFromUuid(string(params.Wid))
	if err == nil {
		wl.Name = models.RFC1123Label(name)
		wl.Namespace = models.RFC1123Label(namespace)
	} else {
		fmt.Printf("failed to get pod form uuid, error: %v\n", err)
	}
	list.Lock()
	list.AddToWorkloadList(&wl, params.Rid, params.Wid)
	list.Unlock()
	return workload_monitor.NewPutWorkloadByRegistrantOK().WithPayload(&wl)
}

type getWorkloadByRegistrantImpl struct{}

func NewGetWorkloadByRegistrantHandler() workload_monitor.GetWorkloadByRegistrantHandler {
	return &getWorkloadByRegistrantImpl{}
}

func (impl *getWorkloadByRegistrantImpl) Handle(params workload_monitor.GetWorkloadByRegistrantParams) middleware.Responder {
	list.Lock()
	wl := *list.GetWorkload(params.Rid, params.Wid)
	list.Unlock()
	return workload_monitor.NewGetWorkloadByRegistrantOK().WithPayload(&wl)
}

type deleteWorkloadByRegistrantImpl struct{}

func NewDeleteWorkloadByRegistrantHandler() workload_monitor.DeleteWorkloadByRegistrantHandler {
	return &deleteWorkloadByRegistrantImpl{}
}

func (impl *deleteWorkloadByRegistrantImpl) Handle(params workload_monitor.DeleteWorkloadByRegistrantParams) middleware.Responder {
	wl := *list.GetWorkload(params.Rid, params.Wid)
	list.Lock()
	list.DeleteWorkload(params.Rid, params.Wid)
	list.Unlock()
	return workload_monitor.NewDeleteWorkloadByRegistrantOK().WithPayload(&wl)
}

type deleteWorkloadsByRegistrantImpl struct{}

func NewDeleteWorkloadsByRegistrantHandler() workload_monitor.DeleteWorkloadsByRegistrantHandler {
	return &deleteWorkloadsByRegistrantImpl{}
}

func (impl *deleteWorkloadsByRegistrantImpl) Handle(params workload_monitor.DeleteWorkloadsByRegistrantParams) middleware.Responder {
	wl := models.RequestOutcome{}
	list.Lock()
	items := list.DeleteWorkloads(params.Rid)
	list.Unlock()
	wl.ItemsAffected = &items
	wl.Details = ""
	wl.Result = "OK"
	return workload_monitor.NewDeleteWorkloadsByRegistrantOK().WithPayload(&wl)
}
