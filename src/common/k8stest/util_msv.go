package k8stest

import (
	"encoding/json"
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type MayastorCpVolume struct {
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

func GetMayastorCpVolume(uuid string) (*MayastorCpVolume, error) {
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
	var response MayastorCpVolume
	err = json.Unmarshal(jsonInput, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func ListMayastorCpVolumes() ([]MayastorCpVolume, error) {
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
		cmd := exec.Command(pluginpath, "-r", url, "-ojson", "get", "volumes")
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
	var response []MayastorCpVolume
	err = json.Unmarshal(jsonInput, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func scaleMayastorVolume(uuid string, replicaCount int) error {
	pluginpath := fmt.Sprintf("%s/%s",
		e2e_config.GetConfig().KubectlPluginDir,
		common.KubectlMayastorPlugin)

	address := GetMayastorNodeIPAddresses()
	if len(address) == 0 {
		return fmt.Errorf("mayastor nodes not found")
	}
	var err error
	for _, addr := range address {
		url := fmt.Sprintf("http://%s:%s", addr, common.PluginPort)
		cmd := exec.Command(pluginpath, "-r", url, "scale", "volume", uuid, string(replicaCount))
		_, err = cmd.CombinedOutput()
		if err == nil {
			break
		} else {
			logf.Log.Info("Error while executing kubectl mayastor command", "node IP", addr, "error", err)
		}
	}
	if err != nil {
		return err
	}
	return nil
}

func GetMayastorVolumeState(volName string) (string, error) {
	msv, err := GetMayastorCpVolume(volName)
	if err == nil {
		return msv.State.Status, nil
	}
	return "", err
}

func GetMayastorVolumeChildren(volName string) ([]children, error) {
	msv, err := GetMayastorCpVolume(volName)
	if err != nil {
		return nil, err
	}
	return msv.State.Child.Children, nil
}

func GetMayastorVolumeChildState(uuid string) (string, error) {
	msv, err := GetMayastorCpVolume(uuid)
	if err != nil {
		return "", err
	}
	return msv.State.Child.State, nil
}

func IsMmayastorVolumePublished(uuid string) bool {
	msv, err := GetMayastorCpVolume(uuid)
	if err == nil {
		return msv.Spec.Target_node != ""
	}
	return false
}

func IsMayastorVolumeDeleted(uuid string) bool {
	msv, err := GetMayastorCpVolume(uuid)
	if strings.ToLower(msv.State.Status) == "destroyed" {
		return false
	}
	if err != nil && errors.IsNotFound(err) {
		return true
	}
	return false
}

func CheckForMayastorVolumes() (bool, error) {
	log.Log.Info("CheckForMayastorVolumes")
	foundResources := false

	msvs, err := ListMayastorCpVolumes()
	if err == nil && msvs != nil && len(msvs) != 0 {
		log.Log.Info("CheckForVolumeResources: found MayastorVolumes",
			"MayastorVolumes", msvs)
		foundResources = true
	}
	return foundResources, err
}

func CheckAllMayastorVolumesAreHealthy() error {
	allHealthy := true
	msvs, err := ListMayastorCpVolumes()
	if err == nil && msvs != nil && len(msvs) != 0 {
		for _, msv := range msvs {
			if strings.ToLower(msv.State.Status) != "healthy" {
				log.Log.Info("CheckAllMayastorVolumesAreHealthy", "msv", msv)
				allHealthy = false
			}
		}
	}

	if !allHealthy {
		return fmt.Errorf("all MSVs were not healthy")
	}
	return err
}
