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
	Spec  Spec  `json:"spec"`
	State State `json:"state"`
}

type Spec struct {
	Labels       string `json:"labels"`
	Num_replicas int    `json:"num_replicas"`
	Operation    string `json:"operation"`
	Protocol     string `json:"protocol"`
	Size         int64  `json:"size"`
	State        string `json:"state"`
	Target_node  string `json:"target_node"`
	Uuid         string `json:"uuid"`
}

type State struct {
	Child    Child  `json:"child"`
	Protocol string `json:"protocol"`
	Size     int64  `json:"size"`
	Status   string `json:"status"`
	Uuid     string `json:"uuid"`
}

type Child struct {
	Children  []Children `json:"children"`
	DeviceUri string     `json:"deviceUri"`
	Node      string     `json:"node"`
	Rebuilds  int        `json:"rebuilds"`
	Share     string     `json:"share"`
	Size      int64      `json:"size"`
	State     string     `json:"state"`
	Uuid      string     `json:"uuid"`
}

type Children struct {
	RebuildProgress string `json:"rebuildProgress"`
	State           string `json:"state"`
	Uri             string `json:"uri"`
}

func GetMayastorVolume(uuid string) (*MayastorVolume, error) {
	pluginpath := fmt.Sprintf("%s/%s",
		e2e_config.GetConfig().KubectlPluginDir,
		common.KubectlMayastorPlugin)
	url := ""
	mayastorNodes, err := GetNodeLocs()
	if err != nil {
		return nil, err
	}
	if len(mayastorNodes) == 0 {
		return nil, fmt.Errorf("mayastor nodes not found")
	}
	for _, node := range mayastorNodes {
		if node.MayastorNode {
			url = fmt.Sprintf("http://%s:%s", node.IPAddress, common.PluginPort)
			break
		}
	}
	cmd := exec.Command(pluginpath, "-r", url, "-ojson", "get", "volume", uuid)
	jsonInput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	logf.Log.Info("cmd", "output", string(jsonInput))

	var response MayastorVolume
	err = json.Unmarshal(jsonInput, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
