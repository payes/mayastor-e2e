package k8stest

import (
	"encoding/json"
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"os/exec"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type MayastorVolume struct {
	Spec  msvSpec  `json:"spec"`
	State msvState `json:"state"`
}

type msvSpec struct {
	Labels       []string `json:"labels"`
	Num_replicas int      `json:"num_replicas"`
	Operation    string   `json:"operation"`
	Protocol     string   `json:"protocol"`
	Size         int64    `json:"size"`
	State        string   `json:"state"`
	Target_node  string   `json:"target_node"`
	Uuid         string   `json:"uuid"`
}

type msvState struct {
	Child    child  `json:"child"`
	Protocol string `json:"protocol"`
	Size     int64  `json:"size"`
	Status   string `json:"status"`
	Uuid     string `json:"uuid"`
}

type child struct {
	Children  []children `json:"children"`
	DeviceUri string     `json:"deviceUri"`
	Node      string     `json:"node"`
	Rebuilds  int        `json:"rebuilds"`
	Share     string     `json:"share"`
	Size      int64      `json:"size"`
	State     string     `json:"state"`
	Uuid      string     `json:"uuid"`
}

type children struct {
	RebuildProgress string `json:"rebuildProgress"`
	State           string `json:"state"`
	Uri             string `json:"uri"`
}

func GetMayastorVolume(uuid string) (*MayastorVolume, error) {
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
		cmd := exec.Command(pluginpath, "-r", url, "-ojson", "get", "volume", uuid)
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
	// FIXME use MayastorVolume when bug in kubectl mayastor plugin is fixed
	var response []MayastorVolume
	err = json.Unmarshal(jsonInput, &response)
	if err != nil {
		return nil, err
	}
	return &response[0], nil
}
