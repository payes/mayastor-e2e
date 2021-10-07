package k8stest

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

func GetMayastorCpPool(name string) (*MayastorCpPool, error) {
	pluginpath := fmt.Sprintf("%s/%s",
		e2e_config.GetConfig().KubectlPluginDir,
		common.KubectlMayastorPlugin)
	address := GetMayastorNodeIPAddresses()

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
		return nil, err
	}
	return &response, nil
}
