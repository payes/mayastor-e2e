package v1

// Utility functions for Mayastor control plane volume
import (
	"encoding/json"
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func HasNotFoundRestJsonError(str string) bool {
	re := regexp.MustCompile(`Error error in response.*RestJsonError.*kind:\s*(\w+)`)
	frags := re.FindSubmatch([]byte(str))
	return len(frags) == 2 && string(frags[1]) == "NotFound"
}

type MayastorCpVolume struct {
	Spec  msvSpec  `json:"spec"`
	State msvState `json:"state"`
}

type msvSpec struct {
	Num_replicas int        `json:"num_replicas"`
	Size         int64      `json:"size"`
	Status       string     `json:"status"`
	Target       specTarget `json:"target"`
	Uuid         string     `json:"uuid"`
	Topology     topology   `json:"topology"`
	Policy       policy     `json:"policy"`
}

type policy struct {
	Self_heal bool `json:"self_heal"`
}
type specTarget struct {
	Protocol string `json:"protocol"`
	Node     string `json:"node"`
}

type topology struct {
	NodeTopology node_topology `json:"node_topology"`
	PoolTopology pool_topology `json:"pool_topology"`
}
type node_topology struct {
	Explicit explicit `json:"explicit"`
}
type pool_topology struct {
	Labelled labelled `json:"labelled"`
}
type labelled struct {
	Inclusion map[string]interface{} `json:"inclusion"`
	Exclusion map[string]interface{} `json:"exclusion"`
}

type explicit struct {
	AllowedNodes   []string `json:"allowed_nodes"`
	PreferredNodes []string `json:"preferred_nodes"`
}

type msvState struct {
	Target           stateTarget                 `json:"target"`
	Size             int64                       `json:"size"`
	Status           string                      `json:"status"`
	Uuid             string                      `json:"uuid"`
	Replica_topology map[string]replica_topology `json:"replica_topology"`
}

type replica_topology struct {
	Node  string `json:"node"`
	Pool  string `json:"pool"`
	State string `json:"state"`
}

type stateTarget struct {
	Children  []children `json:"children"`
	DeviceUri string     `json:"deviceUri"`
	Node      string     `json:"node"`
	Rebuilds  int        `json:"rebuilds"`
	Protocol  string     `json:"protocol"`
	Size      int64      `json:"size"`
	State     string     `json:"state"`
	Uuid      string     `json:"uuid"`
}

type children struct {
	State string `json:"state"`
	Uri   string `json:"uri"`
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
		msg := string(jsonInput)
		if !HasNotFoundRestJsonError(msg) {
			logf.Log.Info("Failed to unmarshal (get volume)", "string", msg)
		}
		return nil, fmt.Errorf("%s", msg)
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
		errMsg := string(jsonInput)
		logf.Log.Info("Failed to unmarshal (get volumes)", "string", string(jsonInput))
		return []MayastorCpVolume{}, fmt.Errorf("%s", errMsg)
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
		return msv.Spec.Target.Node != ""
	}
	return false
}

func IsMayastorVolumeDeleted(uuid string, address []string) bool {
	msv, err := GetMayastorCpVolume(uuid, address)
	if err != nil {
		if HasNotFoundRestJsonError(fmt.Sprintf("%v", err)) {
			return true
		}
		logf.Log.Error(err, "IsMayastorVolumeDeleted msv is nil")
		return false
	}
	if msv.Spec.Uuid == "" {
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

	for _, children := range cpMsv.State.Target.Children {
		nexusChildren = append(nexusChildren, common.NexusChild{
			State: children.State,
			Uri:   children.Uri,
		})
	}
	var replicas []common.Replica
	for uuid := range cpMsv.State.Replica_topology {
		replica, err := getMayastorCpReplica(uuid, address)
		if err != nil {
			logf.Log.Info("Failed to get replicas", "uuid", uuid, "error", err)
			return common.MayastorVolume{}
		}
		replicas = append(replicas, common.Replica{
			Uri:     replica.Uri,
			Offline: strings.ToLower(replica.State) != "online",
			Node:    replica.Node,
			Pool:    replica.Pool,
		})
	}

	return common.MayastorVolume{
		Name: cpMsv.Spec.Uuid,
		Spec: common.MayastorVolumeSpec{
			Protocol:      cpMsv.Spec.Target.Protocol,
			ReplicaCount:  cpMsv.Spec.Num_replicas,
			RequiredBytes: int(cpMsv.Spec.Size),
		},
		Status: common.MayastorVolumeStatus{
			Nexus: common.Nexus{
				Children:  nexusChildren,
				DeviceUri: cpMsv.State.Target.DeviceUri,
				Node:      cpMsv.State.Target.Node,
				State:     cpMsv.State.Target.State,
				Uuid:      cpMsv.State.Target.Uuid,
			},
			Replicas: replicas,
			Size:     cpMsv.State.Size,
			State:    cpMsv.State.Status,
		},
	}
}

// GetMSV Get pointer to a mayastor control plane volume
// returns nil and no error if the msv is in pending state.
func (cp CPv1) GetMSV(uuid string) (*common.MayastorVolume, error) {
	cpMsv, err := GetMayastorCpVolume(uuid, *cp.nodeIPAddresses)
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

	msv := cpVolumeToMsv(cpMsv, *cp.nodeIPAddresses)
	return &msv, nil
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
func (cp CPv1) GetMsvNodes(uuid string) (string, []string) {
	msv, err := GetMayastorCpVolume(uuid, *cp.nodeIPAddresses)
	if err != nil {
		logf.Log.Info("failed to get mayastor volume", "uuid", uuid)
		return "", nil
	}
	node := msv.State.Target.Node
	replicas := make([]string, msv.Spec.Num_replicas)

	msvReplicas, err := cp.GetMsvReplicas(uuid)
	if err != nil {
		logf.Log.Info("failed to get mayastor volume replica", "uuid", uuid)
		return node, nil
	}
	for ix, r := range msvReplicas {
		replicas[ix] = r.Node
	}
	return node, replicas
}

func (cp CPv1) ListMsvs() ([]common.MayastorVolume, error) {
	var msvs []common.MayastorVolume
	list, err := ListMayastorCpVolumes(*cp.nodeIPAddresses)
	if err == nil {
		for _, item := range list {
			msvs = append(msvs, cpVolumeToMsv(&item, *cp.nodeIPAddresses))
		}
	}
	return msvs, err
}

func (cp CPv1) SetMsvReplicaCount(uuid string, replicaCount int) error {
	err := scaleMayastorVolume(uuid, replicaCount, *cp.nodeIPAddresses)
	logf.Log.Info("ScaleMayastorVolume", "Num_replicas", replicaCount)
	return err
}

func (cp CPv1) GetMsvState(uuid string) (string, error) {
	return GetMayastorVolumeState(uuid, *cp.nodeIPAddresses)
}

func (cp CPv1) GetMsvReplicas(volName string) ([]common.Replica, error) {
	vol, err := cp.GetMSV(volName)
	if err != nil {
		logf.Log.Info("Failed to get replicas", "uuid", volName, "error", err)
		return nil, err
	}
	return vol.Status.Replicas, nil
}

func (cp CPv1) GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	var children []common.NexusChild
	cpVolumeChildren, err := GetMayastorVolumeChildren(volName, *cp.nodeIPAddresses)
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

func (cp CPv1) GetMsvNexusState(uuid string) (string, error) {
	return GetMayastorVolumeChildState(uuid, *cp.nodeIPAddresses)
}

func (cp CPv1) IsMsvPublished(uuid string) bool {
	return IsMmayastorVolumePublished(uuid, *cp.nodeIPAddresses)
}

func (cp CPv1) IsMsvDeleted(uuid string) bool {
	return IsMayastorVolumeDeleted(uuid, *cp.nodeIPAddresses)
}

func (cp CPv1) CheckForMsvs() (bool, error) {
	return CheckForMayastorVolumes(*cp.nodeIPAddresses)
}

func (cp CPv1) CheckAllMsvsAreHealthy() error {
	return CheckAllMayastorVolumesAreHealthy(*cp.nodeIPAddresses)
}

func (cp CPv1) DeleteMsv(volName string) error {
	return fmt.Errorf("delete of mayastor volume not supported %v", volName)
}
