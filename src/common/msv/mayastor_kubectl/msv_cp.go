package mayastor_kubectl

// Utility functions for Mayastor control plane volume
import (
	"encoding/json"
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"os/exec"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type CpMsv struct {
	nodeIPAddresses []string
}

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
	Target   target `json:"target"`
	Protocol string `json:"protocol"`
	Size     int64  `json:"size"`
	Status   string `json:"status"`
	Uuid     string `json:"uuid"`
}

type target struct {
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

func GetMayastorCpVolume(uuid string, address []string) (*MayastorCpVolume, error) {
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
		//FIXME: got response which is not valid JSON
		return &MayastorCpVolume{}, nil
	}
	return &response, nil
}

func ListMayastorCpVolumes(address []string) ([]MayastorCpVolume, error) {
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
		logf.Log.Info("Failed to unmarshal", "string", string(jsonInput))
		//FIXME: got response which is not valid JSON
		return []MayastorCpVolume{}, nil
	}
	return response, nil
}

func scaleMayastorVolume(uuid string, replicaCount int, address []string) error {
	pluginpath := fmt.Sprintf("%s/%s",
		e2e_config.GetConfig().KubectlPluginDir,
		common.KubectlMayastorPlugin)

	if len(address) == 0 {
		return fmt.Errorf("mayastor nodes not found")
	}
	var err error
	for _, addr := range address {
		url := fmt.Sprintf("http://%s:%s", addr, common.PluginPort)
		cmd := exec.Command(pluginpath, "-r", url, "scale", "volume", uuid, strconv.Itoa(replicaCount))
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

func GetMayastorVolumeState(volName string, address []string) (string, error) {
	msv, err := GetMayastorCpVolume(volName, address)
	if err == nil {
		return msv.State.Status, nil
	}
	return "", err
}

func GetMayastorVolumeChildren(volName string, address []string) ([]children, error) {
	msv, err := GetMayastorCpVolume(volName, address)
	if err != nil {
		return nil, err
	}
	return msv.State.Target.Children, nil
}

func GetMayastorVolumeChildState(uuid string, address []string) (string, error) {
	msv, err := GetMayastorCpVolume(uuid, address)
	if err != nil {
		return "", err
	}
	return msv.State.Target.State, nil
}

func IsMmayastorVolumePublished(uuid string, address []string) bool {
	msv, err := GetMayastorCpVolume(uuid, address)
	if err == nil {
		return msv.Spec.Target_node != ""
	}
	return false
}

func IsMayastorVolumeDeleted(uuid string, address []string) bool {
	msv, err := GetMayastorCpVolume(uuid, address)
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

func CheckForMayastorVolumes(address []string) (bool, error) {
	logf.Log.Info("CheckForMayastorVolumes")
	foundResources := false

	msvs, err := ListMayastorCpVolumes(address)
	if err == nil && msvs != nil && len(msvs) != 0 {
		logf.Log.Info("CheckForVolumeResources: found MayastorVolumes",
			"MayastorVolumes", msvs)
		foundResources = true
	}
	return foundResources, err
}

func CheckAllMayastorVolumesAreHealthy(address []string) error {
	allHealthy := true
	msvs, err := ListMayastorCpVolumes(address)
	if err == nil && msvs != nil && len(msvs) != 0 {
		for _, msv := range msvs {
			if msv.State.Status != "Online" {
				logf.Log.Info("CheckAllMayastorVolumesAreHealthy", "msv.State.Status", msv.State.Status)
				allHealthy = false
			}
		}
	}

	if !allHealthy {
		return fmt.Errorf("all MSVs were not healthy")
	}
	return err
}

func cpVolumeToMsv(cpMsv *MayastorCpVolume, address []string) common.MayastorVolume {
	var nexusChildren []common.NexusChild
	var childrenUri = make(map[string]bool)

	for _, children := range cpMsv.State.Target.Children {
		nexusChildren = append(nexusChildren, common.NexusChild{
			State: children.State,
			Uri:   children.Uri,
		})
		//storing uri as key in map[string]boll
		//will be used to determine replica corresponding to children uri
		childrenUri[children.Uri] = true
	}

	var replicas []common.Replica
	cpReplicas, err := listMayastorCpReplicas(address)
	if err != nil {
		logf.Log.Info("Failed to list replicas", "error", err)
		return common.MayastorVolume{}
	}
	for _, cpReplica := range cpReplicas {
		if _, ok := childrenUri[cpReplica.Uri]; ok {
			replicas = append(replicas, common.Replica{
				Uri:     cpReplica.Uri,
				Offline: strings.ToLower(cpReplica.State) != "online",
				Node:    cpReplica.Node,
				Pool:    cpReplica.Pool,
			})
		}

	}

	return common.MayastorVolume{
		Name: cpMsv.Spec.Uuid,
		Spec: common.MayastorVolumeSpec{
			Protocol:      cpMsv.Spec.Protocol,
			ReplicaCount:  cpMsv.Spec.Num_replicas,
			RequiredBytes: int(cpMsv.Spec.Size),
		},
		Status: common.MayastorVolumeStatus{
			Nexus: common.Nexus{
				Children:  nexusChildren,
				DeviceUri: cpMsv.State.Target.DeviceUri,
				Node:      cpMsv.State.Target.Node,
				State:     cpMsv.State.Target.State,
			},
			Replicas: replicas,
			Size:     cpMsv.State.Size,
			State:    cpMsv.State.Status,
		},
	}
}

// GetMSV Get pointer to a mayastor control plane volume
// returns nil and no error if the msv is in pending state.
func (mc CpMsv) GetMSV(uuid string) (*common.MayastorVolume, error) {
	cpMsv, err := GetMayastorCpVolume(uuid, mc.nodeIPAddresses)
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

	msv := cpVolumeToMsv(cpMsv, mc.nodeIPAddresses)
	return &msv, nil
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
func (mc CpMsv) GetMsvNodes(uuid string) (string, []string) {
	msv, err := GetMayastorCpVolume(uuid, mc.nodeIPAddresses)
	if err != nil {
		logf.Log.Info("failed to get mayastor volume", "uuid", uuid)
		return "", nil
	}
	node := msv.State.Target.Node
	replicas := make([]string, msv.Spec.Num_replicas)

	msvReplicas, err := mc.GetMsvReplicas(uuid)
	if err != nil {
		logf.Log.Info("failed to get mayastor volume replica", "uuid", uuid)
		return node, nil
	}
	for ix, r := range msvReplicas {
		replicas[ix] = r.Node
	}
	return node, replicas
}

func (mc CpMsv) ListMsvs() ([]common.MayastorVolume, error) {
	var msvs []common.MayastorVolume
	list, err := ListMayastorCpVolumes(mc.nodeIPAddresses)
	if err == nil {
		for _, item := range list {
			msvs = append(msvs, cpVolumeToMsv(&item, mc.nodeIPAddresses))
		}
	}
	return msvs, err
}

func (mc CpMsv) SetMsvReplicaCount(uuid string, replicaCount int) error {
	err := scaleMayastorVolume(uuid, replicaCount, mc.nodeIPAddresses)
	logf.Log.Info("ScaleMayastorVolume", "Num_replicas", replicaCount)
	return err
}

func (mc CpMsv) GetMsvState(uuid string) (string, error) {
	return GetMayastorVolumeState(uuid, mc.nodeIPAddresses)
}

func (mc CpMsv) GetMsvReplicas(volName string) ([]common.Replica, error) {
	var replicas []common.Replica
	var childrenUri = make(map[string]bool)
	cpVolumeChildren, err := GetMayastorVolumeChildren(volName, mc.nodeIPAddresses)
	if err == nil {
		for _, child := range cpVolumeChildren {
			//storing uri as key in map[string]boll
			//will be used to determine replica corresponding to children uri
			childrenUri[child.Uri] = true
		}
		cpReplicas, err := listMayastorCpReplicas(mc.nodeIPAddresses)
		if err != nil {
			logf.Log.Info("Failed to list replicas", "error", err)
			return nil, err
		}
		for _, cpReplica := range cpReplicas {
			if _, ok := childrenUri[cpReplica.Uri]; ok {
				replicas = append(replicas, common.Replica{
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

func (mc CpMsv) GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	var children []common.NexusChild
	cpVolumeChildren, err := GetMayastorVolumeChildren(volName, mc.nodeIPAddresses)
	if err == nil {
		for _, child := range cpVolumeChildren {
			children = append(children, common.NexusChild{
				State: child.State,
				Uri:   child.Uri,
			})
		}
	}
	return children, err
}

func (mc CpMsv) GetMsvNexusState(uuid string) (string, error) {
	return GetMayastorVolumeChildState(uuid, mc.nodeIPAddresses)
}

func (mc CpMsv) IsMsvPublished(uuid string) bool {
	return IsMmayastorVolumePublished(uuid, mc.nodeIPAddresses)
}

func (mc CpMsv) IsMsvDeleted(uuid string) bool {
	return IsMayastorVolumeDeleted(uuid, mc.nodeIPAddresses)
}

func (mc CpMsv) CheckForMsvs() (bool, error) {
	return CheckForMayastorVolumes(mc.nodeIPAddresses)
}

func (mc CpMsv) CheckAllMsvsAreHealthy() error {
	return CheckAllMayastorVolumesAreHealthy(mc.nodeIPAddresses)
}

func (mc CpMsv) DeleteMsv(volName string) error {
	return fmt.Errorf("delete of mayastor volume not supported %v", volName)
}

func MakeCpMsv(nodeIPAddresses []string) CpMsv {
	return CpMsv{
		nodeIPAddresses: nodeIPAddresses,
	}
}
