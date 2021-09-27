// This file is safe to edit. Once it exists it will not be overwritten

package wm

import (
	"sync"

	"github.com/go-openapi/strfmt"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/models"
)

type WorkloadList struct {
	mu          sync.Mutex
	WorkloadMap map[strfmt.UUID]map[strfmt.UUID]*models.Workload
}

var gWorkloadList WorkloadList

func init() {
	gWorkloadList.WorkloadMap = make(map[strfmt.UUID]map[strfmt.UUID]*models.Workload)
}

func AddToWorkloadList(pwl *models.Workload, rid strfmt.UUID, wid strfmt.UUID) {
	gWorkloadList.mu.Lock()
	if _, found := gWorkloadList.WorkloadMap[rid]; !found {
		gWorkloadList.WorkloadMap[rid] = make(map[strfmt.UUID]*models.Workload)
	}
	gWorkloadList.WorkloadMap[rid][wid] = pwl
	gWorkloadList.mu.Unlock()
}

func GetWorkload(rid strfmt.UUID, wid strfmt.UUID) *models.Workload {
	gWorkloadList.mu.Lock()
	if wlmap, found := gWorkloadList.WorkloadMap[rid]; found {
		if pwl, found := wlmap[wid]; found {
			wl := *pwl
			gWorkloadList.mu.Unlock()
			return &wl
		}
	}
	gWorkloadList.mu.Unlock()
	return nil
}

func DeleteWorkload(rid strfmt.UUID, wid strfmt.UUID) {
	gWorkloadList.mu.Lock()
	if wlmap, found := gWorkloadList.WorkloadMap[rid]; found {
		delete(wlmap, wid)
	}
	gWorkloadList.mu.Unlock()
}

func DeleteWorkloads(rid strfmt.UUID) int64 {
	gWorkloadList.mu.Lock()
	items := 0
	if wlmap, found := gWorkloadList.WorkloadMap[rid]; found {
		items = len(wlmap)
		delete(gWorkloadList.WorkloadMap, rid)
	}
	gWorkloadList.mu.Unlock()
	return int64(items)
}

func GetWorkloadListByRegistrant(rid strfmt.UUID) []models.Workload {
	gWorkloadList.mu.Lock()
	var list []models.Workload
	if wlmap, found := gWorkloadList.WorkloadMap[rid]; found {
		for _, wl := range wlmap {
			list = append(list, *wl)
		}
	}
	gWorkloadList.mu.Unlock()
	return list
}

func GetWorkloadList() []models.Workload {
	gWorkloadList.mu.Lock()
	var list []models.Workload

	for _, wlmap := range gWorkloadList.WorkloadMap {
		for _, wl := range wlmap {
			list = append(list, *wl)
		}
	}
	gWorkloadList.mu.Unlock()
	return list
}

func DeleteWorkloadById(ID strfmt.UUID) {
	gWorkloadList.mu.Lock()

	for _, wlmap := range gWorkloadList.WorkloadMap {
		for wid, wl := range wlmap {
			if wl.ID == ID {
				delete(wlmap, wid)
			}
		}
	}
	gWorkloadList.mu.Unlock()
}
