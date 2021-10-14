package list

import (
	"sync"

	"github.com/go-openapi/strfmt"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/models"
)

type WorkloadListItem struct {
	Restarts int32
	Wl       *models.Workload
	Rid      strfmt.UUID
}

type WorkloadList struct {
	mu          sync.Mutex
	WorkloadMap map[strfmt.UUID]map[strfmt.UUID]WorkloadListItem
}

var gWorkloadList WorkloadList

func init() {
	gWorkloadList.WorkloadMap = make(map[strfmt.UUID]map[strfmt.UUID]WorkloadListItem)
}

func Lock() {
	gWorkloadList.mu.Lock()
}

func Unlock() {
	gWorkloadList.mu.Unlock()
}

func AddToWorkloadList(pwl *models.Workload, rid strfmt.UUID, wid strfmt.UUID) {
	if _, found := gWorkloadList.WorkloadMap[rid]; !found {
		gWorkloadList.WorkloadMap[rid] = make(map[strfmt.UUID]WorkloadListItem)
	}
	wli := WorkloadListItem{}
	wli.Restarts = 0
	wli.Wl = pwl
	wli.Rid = rid
	gWorkloadList.WorkloadMap[rid][wid] = wli
}

func SetWorkloadListItemRestarts(rid strfmt.UUID, wid strfmt.UUID, restarts int32) bool {
	if _, found := gWorkloadList.WorkloadMap[rid]; found {
		if wli, found := gWorkloadList.WorkloadMap[rid][wid]; found {
			wli.Restarts = restarts
			gWorkloadList.WorkloadMap[rid][wid] = wli
			return true
		}
	}
	return false
}

func GetWorkload(rid strfmt.UUID, wid strfmt.UUID) *models.Workload {
	if wlmap, found := gWorkloadList.WorkloadMap[rid]; found {
		if wli, found := wlmap[wid]; found {
			return wli.Wl
		}
	}
	return nil
}

func DeleteWorkload(rid strfmt.UUID, wid strfmt.UUID) {
	if wlmap, found := gWorkloadList.WorkloadMap[rid]; found {
		delete(wlmap, wid)
	}
}

func DeleteWorkloads(rid strfmt.UUID) int64 {
	items := 0
	if wlmap, found := gWorkloadList.WorkloadMap[rid]; found {
		items = len(wlmap)
		delete(gWorkloadList.WorkloadMap, rid)
	}
	return int64(items)
}

func GetWorkloadListByRegistrant(rid strfmt.UUID) []*models.Workload {
	var list []*models.Workload
	if wlmap, found := gWorkloadList.WorkloadMap[rid]; found {
		for _, wli := range wlmap {
			list = append(list, wli.Wl)
		}
	}
	return list
}

func GetWorkloadList() []*models.Workload {
	var list []*models.Workload

	for _, wlmap := range gWorkloadList.WorkloadMap {
		for _, wli := range wlmap {
			list = append(list, wli.Wl)
		}
	}
	return list
}

func GetWorkloadItemList() []WorkloadListItem {
	var list []WorkloadListItem

	for _, wlmap := range gWorkloadList.WorkloadMap {
		for _, wli := range wlmap {
			list = append(list, wli)
		}
	}
	return list
}

func DeleteWorkloadById(ID strfmt.UUID) {
	for _, wlmap := range gWorkloadList.WorkloadMap {
		for wid, wli := range wlmap {
			if wli.Wl.ID == ID {
				delete(wlmap, wid)
			}
		}
	}
}
