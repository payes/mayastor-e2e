package v1

import (
	"encoding/json"
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"os/exec"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type MayastorCpPool struct {
	Id    string   `json:"id"`
	Spec  mspSpec  `json:"spec"`
	State mspState `json:"state"`
}

type mspSpec struct {
	Disks  []string          `json:"disks"`
	Id     string            `json:"id"`
	Labels map[string]string `json:"labels"`
	Node   string            `json:"node"`
	Status string            `json:"status"`
}

type mspState struct {
	Capacity int64    `json:"capacity"`
	Disks    []string `json:"disks"`
	ID       string   `json:"id"`
	Node     string   `json:"node"`
	Status   string   `json:"status"`
	Used     int64    `json:"used"`
}

func GetMayastorCpPool(name string, address []string) (*MayastorCpPool, error) {
	pluginpath := fmt.Sprintf("%s/%s",
		e2e_config.GetConfig().KubectlPluginDir,
		common.KubectlMayastorPlugin)

	if len(address) == 0 {
		return nil, fmt.Errorf("mayastor nodes not found")
	}
	var jsonInput []byte
	var err error
	for _, addr := range address {
		url := fmt.Sprintf("http://%s:%s", addr, common.PluginPort)
		cmd := exec.Command(pluginpath, "-r", url, "-ojson", "get", "pool", name)
		jsonInput, err = cmd.CombinedOutput()
		if err == nil {
			break
		} else {
			logf.Log.Info("Error while executing kubectl-plugin mayastor command", "node IP", addr, "error", err)
		}
	}
	if err != nil {
		return nil, err
	}

	var response MayastorCpPool
	err = json.Unmarshal(jsonInput, &response)
	if err != nil {
		msg := string(jsonInput)
		if !HasNotFoundRestJsonError(msg) {
			logf.Log.Info("Failed to unmarshal (get pool)", "string", msg)
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return &response, nil
}

func ListMayastorCpPools(address []string) ([]MayastorCpPool, error) {
	pluginpath := fmt.Sprintf("%s/%s",
		e2e_config.GetConfig().KubectlPluginDir,
		common.KubectlMayastorPlugin)

	if len(address) == 0 {
		return nil, fmt.Errorf("mayastor nodes not found")
	}
	var jsonInput []byte
	var err error
	for _, addr := range address {
		url := fmt.Sprintf("http://%s:%s", addr, common.PluginPort)
		cmd := exec.Command(pluginpath, "-r", url, "-ojson", "get", "pools")
		jsonInput, err = cmd.CombinedOutput()
		if err == nil {
			break
		} else {
			logf.Log.Info("Error while executing kubectl mayastor command", "node IP", addr, "error", err)
		}
	}
	if err != nil {
		return nil, err
	}
	var response []MayastorCpPool
	err = json.Unmarshal(jsonInput, &response)
	if err != nil {
		errMsg := string(jsonInput)
		logf.Log.Info("Failed to unmarshal (get pools)", "string", string(jsonInput))
		return []MayastorCpPool{}, fmt.Errorf("%s", errMsg)
	}
	return response, nil
}

func cpMspToMsp(cpMsp *MayastorCpPool, address []string) common.MayastorPool {
	return common.MayastorPool{
		Name: cpMsp.Id,
		Spec: common.MayastorPoolSpec{
			Node:  cpMsp.Spec.Node,
			Disks: cpMsp.Spec.Disks,
		},
		Status: common.MayastorPoolStatus{
			Capacity: cpMsp.State.Capacity,
			Used:     cpMsp.State.Used,
			Disks:    cpMsp.State.Disks,
			Spec: common.MayastorPoolSpec{
				Disks: cpMsp.Spec.Disks,
				Node:  cpMsp.Spec.Node,
			},
			State:  cpMsp.State.Status,
			Avail:  cpMsp.State.Capacity - cpMsp.State.Used,
			Reason: "",
		},
	}
}

// GetMsPool Get pointer to a mayastor control plane pool
func (cp CPv1) GetMsPool(poolName string) (*common.MayastorPool, error) {
	cpMsp, err := GetMayastorCpPool(poolName, *cp.nodeIPAddresses)
	if err != nil {
		return nil, fmt.Errorf("GetMsPool: %v", err)
	}

	if cpMsp == nil {
		logf.Log.Info("Msp not found", "pool", poolName)
		return nil, nil
	}

	msp := cpMspToMsp(cpMsp, *cp.nodeIPAddresses)
	return &msp, nil
}

func (cp CPv1) ListMsPools() ([]common.MayastorPool, error) {
	var msps []common.MayastorPool
	list, err := ListMayastorCpPools(*cp.nodeIPAddresses)
	if err == nil {
		for _, item := range list {
			msps = append(msps, cpMspToMsp(&item, *cp.nodeIPAddresses))
		}
	}
	return msps, err
}
