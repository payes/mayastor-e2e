package v1

// Utility functions for Mayastor control plane volume
import (
	"encoding/json"
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"os/exec"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type MayastorCpNode struct {
	Spec  msnSpec  `json:"spec"`
	State msnState `json:"state"`
}

type msnSpec struct {
	GrpcEndpoint string `json:"grpcEndpoint"`
	ID           string `json:"id"`
}

type msnState struct {
	GrpcEndpoint string `json:"grpcEndpoint"`
	ID           string `json:"id"`
	Status       string `json:"status"`
}

func GetMayastorCpNode(nodeName string, address []string) (*MayastorCpNode, error) {
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
		cmd := exec.Command(pluginpath, "-r", url, "-ojson", "get", "node", nodeName)
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
	var response MayastorCpNode
	err = json.Unmarshal(jsonInput, &response)
	if err != nil {
		msg := string(jsonInput)
		if !HasNotFoundRestJsonError(msg) {
			logf.Log.Info("Failed to unmarshal (get node)", "string", msg)
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return &response, nil
}

func ListMayastorCpNodes(address []string) ([]MayastorCpNode, error) {
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
		cmd := exec.Command(pluginpath, "-r", url, "-ojson", "get", "nodes")
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
	var response []MayastorCpNode
	err = json.Unmarshal(jsonInput, &response)
	if err != nil {
		errMsg := string(jsonInput)
		logf.Log.Info("Failed to unmarshal (get nodes)", "string", string(jsonInput))
		return []MayastorCpNode{}, fmt.Errorf("%s", errMsg)
	}
	return response, nil
}

func GetMayastorNodeStatus(nodeName string, address []string) (string, error) {
	msn, err := GetMayastorCpNode(nodeName, address)
	if err == nil {
		return msn.State.Status, nil
	}
	return "", err
}

func cpNodeToMsn(cpMsn *MayastorCpNode, address []string) common.MayastorNode {
	return common.MayastorNode{
		Spec: common.MayastorNodeSpec{
			ID:           cpMsn.Spec.ID,
			GrpcEndpoint: cpMsn.Spec.GrpcEndpoint,
		},
		State: common.MayastorNodeState{
			ID:           cpMsn.State.ID,
			Status:       cpMsn.State.Status,
			GrpcEndpoint: cpMsn.State.GrpcEndpoint,
		},
	}
}

// GetMSN Get pointer to a mayastor control plane volume
// returns nil and no error if the msn is in pending state.
func (cp CPv1) GetMSN(nodeName string) (*common.MayastorNode, error) {
	cpMsn, err := GetMayastorCpNode(nodeName, *cp.nodeIPAddresses)
	if err != nil {
		return nil, fmt.Errorf("GetMSN: %v", err)
	}

	if cpMsn == nil {
		logf.Log.Info("Msn not found", "node", nodeName)
		return nil, nil
	}

	msn := cpNodeToMsn(cpMsn, *cp.nodeIPAddresses)
	return &msn, nil
}

func (cp CPv1) ListMsns() ([]common.MayastorNode, error) {
	var msns []common.MayastorNode
	list, err := ListMayastorCpNodes(*cp.nodeIPAddresses)
	if err == nil {
		for _, item := range list {
			msns = append(msns, cpNodeToMsn(&item, *cp.nodeIPAddresses))
		}
	}
	return msns, err
}

func (cp CPv1) GetMsNodeStatus(nodeName string) (string, error) {
	cpMsn, err := GetMayastorCpNode(nodeName, *cp.nodeIPAddresses)
	if err != nil {
		return "", fmt.Errorf("GetMsNodeStatus: %v", err)
	}
	return cpMsn.State.Status, nil
}
