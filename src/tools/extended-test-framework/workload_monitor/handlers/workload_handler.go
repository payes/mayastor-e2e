package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/list"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/models"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/restapi/operations/workload_monitor"
)

type putWorkloadByRegistrantImpl struct{}

func NewPutWorkloadByRegistrantHandler() workload_monitor.PutWorkloadByRegistrantHandler {
	return &putWorkloadByRegistrantImpl{}
}

func (impl *putWorkloadByRegistrantImpl) Handle(params workload_monitor.PutWorkloadByRegistrantParams) middleware.Responder {
	logf.Log.Info("received PutWorkloadByRegistrant")
	if params.Body == nil {
		i := int64(1)
		return workload_monitor.NewPutWorkloadByRegistrantBadRequest().WithPayload(&models.RequestOutcome{
			Details:       "Body not provided",
			ItemsAffected: &i,
			Result:        models.RequestOutcomeResultREFUSED,
		})
	}

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
		logf.Log.Info("failed to get pod from uuid", "error", err)
		i := int64(1)
		return workload_monitor.NewPutWorkloadByRegistrantBadRequest().WithPayload(&models.RequestOutcome{
			Details:       err.Error(),
			ItemsAffected: &i,
			Result:        models.RequestOutcomeResultREFUSED,
		})
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
	logf.Log.Info("received GetWorkloadByRegistrant")
	var wl models.Workload
	list.Lock()
	pwl := list.GetWorkload(params.Rid, params.Wid)
	if pwl != nil {
		wl = *pwl
	}
	list.Unlock()
	if pwl != nil {
		return workload_monitor.NewGetWorkloadByRegistrantOK().WithPayload(&wl)
	} else {
		return workload_monitor.NewGetWorkloadByRegistrantNotFound()
	}
}

type getWorkloadsByRegistrantImpl struct{}

func NewGetWorkloadsByRegistrantHandler() workload_monitor.GetWorkloadsByRegistrantHandler {
	return &getWorkloadsByRegistrantImpl{}
}

func (impl *getWorkloadsByRegistrantImpl) Handle(params workload_monitor.GetWorkloadsByRegistrantParams) middleware.Responder {
	logf.Log.Info("received GetWorkloadsByRegistrant")
	list.Lock()
	wll := list.GetWorkloadListByRegistrant(params.Rid)
	list.Unlock()
	return workload_monitor.NewGetWorkloadsByRegistrantOK().WithPayload(wll)
}

type getWorkloadsImpl struct{}

func NewGetWorkloadsHandler() workload_monitor.GetWorkloadsHandler {
	return &getWorkloadsImpl{}
}

func (impl *getWorkloadsImpl) Handle(params workload_monitor.GetWorkloadsParams) middleware.Responder {
	logf.Log.Info("received GetWorkloads")
	list.Lock()
	wll := list.GetWorkloadList()
	list.Unlock()
	return workload_monitor.NewGetWorkloadsOK().WithPayload(wll)
}

type deleteWorkloadByRegistrantImpl struct{}

func NewDeleteWorkloadByRegistrantHandler() workload_monitor.DeleteWorkloadByRegistrantHandler {
	return &deleteWorkloadByRegistrantImpl{}
}

func (impl *deleteWorkloadByRegistrantImpl) Handle(params workload_monitor.DeleteWorkloadByRegistrantParams) middleware.Responder {
	logf.Log.Info("received DeleteWorkloadByRegistrant")
	var wl models.Workload
	list.Lock()
	pwl := list.GetWorkload(params.Rid, params.Wid)
	if pwl != nil {
		list.DeleteWorkload(params.Rid, params.Wid)
		wl = *pwl
	}
	list.Unlock()
	if pwl != nil {
		return workload_monitor.NewDeleteWorkloadByRegistrantOK().WithPayload(&wl)
	} else {
		logf.Log.Info("DeleteWorkloadByRegistrant could not find workload", "Rid", params.Rid, "Wid", params.Wid)
		return workload_monitor.NewDeleteWorkloadByRegistrantNotFound()
	}
}

type deleteWorkloadsByRegistrantImpl struct{}

func NewDeleteWorkloadsByRegistrantHandler() workload_monitor.DeleteWorkloadsByRegistrantHandler {
	return &deleteWorkloadsByRegistrantImpl{}
}

func (impl *deleteWorkloadsByRegistrantImpl) Handle(params workload_monitor.DeleteWorkloadsByRegistrantParams) middleware.Responder {
	logf.Log.Info("received DeleteWorkloadsByRegistrant")
	wl := models.RequestOutcome{}
	list.Lock()
	items := list.DeleteWorkloads(params.Rid)
	list.Unlock()
	wl.ItemsAffected = &items
	wl.Details = ""
	wl.Result = "OK"
	return workload_monitor.NewDeleteWorkloadsByRegistrantOK().WithPayload(&wl)
}
