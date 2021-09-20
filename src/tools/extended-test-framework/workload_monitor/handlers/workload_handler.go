package handlers

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/models"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/util"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/restapi/operations/workload_monitor"
)

//var WorkloadMap = map[strfmt.UUID]map[strfmt.UUID]*models.Workload{}

type putWorkloadByRegistrantImpl struct{}

func NewPutWorkloadByRegistrantHandler() workload_monitor.PutWorkloadByRegistrantHandler {
	return &putWorkloadByRegistrantImpl{}
}

func (impl *putWorkloadByRegistrantImpl) Handle(params workload_monitor.PutWorkloadByRegistrantParams) middleware.Responder {

	fmt.Println("put workload request received")
	var wl models.Workload
	wl.ID = params.Wid
	wl.Name = ""
	wl.Namespace = ""
	wl.WorkloadSpec = *params.Body

	name, namespace, err := util.GetPodNameAndNamespaceFromUuid(string(params.Wid))
	if err == nil {
		wl.Name = models.RFC1123Label(name)
		wl.Namespace = models.RFC1123Label(namespace)
	} else {
		fmt.Printf("failed to get pod form uuid, error: %v\n", err)
	}

	util.AddToWorkloadList(&wl, params.Rid, params.Wid)
	return workload_monitor.NewPutWorkloadByRegistrantOK().WithPayload(&wl)
}

type getWorkloadByRegistrantImpl struct{}

func NewGetWorkloadByRegistrantHandler() workload_monitor.GetWorkloadByRegistrantHandler {
	return &getWorkloadByRegistrantImpl{}
}

func (impl *getWorkloadByRegistrantImpl) Handle(params workload_monitor.GetWorkloadByRegistrantParams) middleware.Responder {
	pwl := util.GetWorkload(params.Rid, params.Wid)

	return workload_monitor.NewGetWorkloadByRegistrantOK().WithPayload(pwl)
}

type deleteWorkloadByRegistrantImpl struct{}

func NewDeleteWorkloadByRegistrantHandler() workload_monitor.DeleteWorkloadByRegistrantHandler {
	return &deleteWorkloadByRegistrantImpl{}
}

func (impl *deleteWorkloadByRegistrantImpl) Handle(params workload_monitor.DeleteWorkloadByRegistrantParams) middleware.Responder {
	pwl := util.GetWorkload(params.Rid, params.Wid)
	util.DeleteWorkload(params.Rid, params.Wid)
	return workload_monitor.NewDeleteWorkloadByRegistrantOK().WithPayload(pwl)
}

type deleteWorkloadsByRegistrantImpl struct{}

func NewDeleteWorkloadsByRegistrantHandler() workload_monitor.DeleteWorkloadsByRegistrantHandler {
	return &deleteWorkloadsByRegistrantImpl{}
}

func (impl *deleteWorkloadsByRegistrantImpl) Handle(params workload_monitor.DeleteWorkloadsByRegistrantParams) middleware.Responder {
	wl := models.RequestOutcome{}
	items := util.DeleteWorkloads(params.Rid)
	wl.ItemsAffected = &items
	wl.Details = ""
	wl.Result = "OK"
	return workload_monitor.NewDeleteWorkloadsByRegistrantOK().WithPayload(&wl)
}
