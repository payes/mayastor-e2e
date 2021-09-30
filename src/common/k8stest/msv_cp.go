package k8stest

// Utility functions for Mayastor control plane volume
import (
	"encoding/json"
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type CpMsv struct{}

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
		logf.Log.Info("Failed to unmarshal", "string", string(jsonInput))
		return &MayastorCpVolume{}, nil
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
	if msv.Spec.Uuid == "" {
		return true
	}
	if strings.ToLower(msv.State.Status) == "destroyed" {
		return false
	}
	if err != nil && errors.IsNotFound(err) {
		return true
	}
	return false
}

func CheckForMayastorVolumes() (bool, error) {
	logf.Log.Info("CheckForMayastorVolumes")
	foundResources := false

	msvs, err := ListMayastorCpVolumes()
	if err == nil && msvs != nil && len(msvs) != 0 {
		logf.Log.Info("CheckForVolumeResources: found MayastorVolumes",
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
				logf.Log.Info("CheckAllMayastorVolumesAreHealthy", "msv", msv)
				allHealthy = false
			}
		}
	}

	if !allHealthy {
		return fmt.Errorf("all MSVs were not healthy")
	}
	return err
}

func cpVolumeToMsv(cpMsv *MayastorCpVolume) MayastorVolume {
	var nexusChildren []NexusChild
	var childrenUri = make(map[string]bool)

	for _, children := range cpMsv.State.Child.Children {
		nexusChildren = append(nexusChildren, NexusChild{
			State: children.State,
			Uri:   children.Uri,
		})
		//storing uri as key in map[string]boll
		//will be used to determine replica corresponding to children uri
		childrenUri[children.Uri] = true
	}

	var replicas []Replica
	cpReplicas, err := ListMayastorCpReplicas()
	if err != nil {
		logf.Log.Info("Failed to list replicas", "error", err)
		return MayastorVolume{}
	}
	for _, cpReplica := range cpReplicas {
		if _, ok := childrenUri[cpReplica.Uri]; ok {
			replicas = append(replicas, Replica{
				Uri:     cpReplica.Uri,
				Offline: strings.ToLower(cpReplica.State) != "online",
				Node:    cpReplica.Node,
				Pool:    cpReplica.Pool,
			})
		}

	}

	return MayastorVolume{
		Name: cpMsv.Spec.Uuid,
		Spec: MayastorVolumeSpec{
			Protocol:      cpMsv.Spec.Protocol,
			ReplicaCount:  cpMsv.Spec.Num_replicas,
			RequiredBytes: int(cpMsv.Spec.Size),
		},
		Status: MayastorVolumeStatus{
			Nexus: Nexus{
				Children:  nexusChildren,
				DeviceUri: cpMsv.State.Child.DeviceUri,
				Node:      cpMsv.State.Child.Node,
				State:     cpMsv.State.Child.State,
			},
			Replicas: replicas,
			Size:     cpMsv.State.Size,
			State:    cpMsv.State.Status,
		},
	}
}

// GetMSV Get pointer to a mayastor control plane volume
// returns nil and no error if the msv is in pending state.
func (mc CpMsv) getMSV(uuid string) (*MayastorVolume, error) {
	cpMsv, err := GetMayastorCpVolume(uuid)
	if err != nil {
		return nil, fmt.Errorf("GetMSV: %v", err)
	}
	if cpMsv.Spec.Uuid == "" {
		logf.Log.Info("Msv not found", "uuid", uuid)
		return nil, nil
	}

	// pending means still being created
	if cpMsv.State.Status == "pending" {
		return nil, nil
	}

	//logf.Log.Info("GetMSV", "msv", msv)
	// Note: msVol.Node can be unassigned here if the volume is not mounted
	if cpMsv.State.Status == "" {
		return nil, fmt.Errorf("GetMSV: state not defined, got msv.Status=\"%v\"", cpMsv.State)
	}

	if cpMsv.Spec.Num_replicas < 1 {
		return nil, fmt.Errorf("GetMSV: msv.Spec.Num_replicas=\"%v\"", cpMsv.Spec.Num_replicas)
	}

	msv := cpVolumeToMsv(cpMsv)
	return &msv, nil
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
func (mc CpMsv) getMsvNodes(uuid string) (string, []string) {
	msv, err := GetMayastorCpVolume(uuid)
	if err != nil {
		logf.Log.Info("failed to get mayastor volume", "uuid", uuid)
		return "", nil
	}
	node := msv.State.Child.Node
	replicas := make([]string, msv.Spec.Num_replicas)

	msvReplicas, err := GetMsvIfc().getMsvReplicas(uuid)
	if err != nil {
		logf.Log.Info("failed to get mayastor volume replica", "uuid", uuid)
		return node, nil
	}
	for ix, r := range msvReplicas {
		replicas[ix] = r.Node
	}
	return node, replicas
}

func (mc CpMsv) listMsvs() ([]MayastorVolume, error) {
	var msvs []MayastorVolume
	list, err := ListMayastorCpVolumes()
	if err == nil {
		for _, item := range list {
			msvs = append(msvs, cpVolumeToMsv(&item))
		}
	}
	return msvs, err
}

func (mc CpMsv) setMsvReplicaCount(uuid string, replicaCount int) error {
	err := scaleMayastorVolume(uuid, replicaCount)
	logf.Log.Info("ScaleMayastorVolume", "Num_replicas", replicaCount)
	return err
}

func (mc CpMsv) getMsvState(uuid string) (string, error) {
	return GetMayastorVolumeState(uuid)
}

func (mc CpMsv) getMsvReplicas(volName string) ([]Replica, error) {
	var replicas []Replica
	var childrenUri = make(map[string]bool)
	cpVolumeChildren, err := GetMayastorVolumeChildren(volName)
	if err == nil {
		for _, child := range cpVolumeChildren {
			//storing uri as key in map[string]boll
			//will be used to determine replica corresponding to children uri
			childrenUri[child.Uri] = true
		}
		cpReplicas, err := ListMayastorCpReplicas()
		if err != nil {
			logf.Log.Info("Failed to list replicas", "error", err)
			return nil, err
		}
		for _, cpReplica := range cpReplicas {
			if _, ok := childrenUri[cpReplica.Uri]; ok {
				replicas = append(replicas, Replica{
					Uri:     cpReplica.Uri,
					Offline: strings.ToLower(cpReplica.State) != "online",
					Node:    cpReplica.Node,
					Pool:    cpReplica.Pool,
				})
			}

		}

	}
	return replicas, nil
}

func (mc CpMsv) getMsvNexusChildren(volName string) ([]NexusChild, error) {
	var children []NexusChild
	cpVolumeChildren, err := GetMayastorVolumeChildren(volName)
	if err == nil {
		for _, child := range cpVolumeChildren {
			children = append(children, NexusChild{
				State: child.State,
				Uri:   child.Uri,
			})
		}
	}
	return children, err
}

func (mc CpMsv) getMsvNexusState(uuid string) (string, error) {
	return GetMayastorVolumeChildState(uuid)
}

func (mc CpMsv) isMsvPublished(uuid string) bool {
	return IsMmayastorVolumePublished(uuid)
}

func (mc CpMsv) isMsvDeleted(uuid string) bool {
	return IsMayastorVolumeDeleted(uuid)
}

func (mc CpMsv) checkForMsvs() (bool, error) {
	return CheckForMayastorVolumes()
}

func (mc CpMsv) checkAllMsvsAreHealthy() error {
	return CheckAllMayastorVolumesAreHealthy()
}

func (mc CpMsv) deleteMsv(volName string) error {
	return fmt.Errorf("delete of mayastor volume not supported")
}

func MakeCpMsv() CpMsv {
	return CpMsv{}
}
