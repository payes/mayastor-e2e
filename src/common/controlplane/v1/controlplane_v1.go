package v1

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"regexp"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type CPv1 struct {
	nodeIPAddresses []string
	pluginPath      string
}

func (cp CPv1) Version() string {
	return "1.0.0"
}

func (cp CPv1) MajorVersion() int {
	return 1
}

var regExs = []*regexp.Regexp{
	regexp.MustCompile(`(Request Timeout)`),
	regexp.MustCompile(`(request timed out)`),
}

func (cp CPv1) IsTimeoutError(err error) bool {
	str := fmt.Sprintf("%v", err)
	for _, re := range regExs {
		frags := re.FindSubmatch([]byte(str))
		if len(frags) != 0 {
			return true
		}
	}
	return false
}

func (cp CPv1) VolStateHealthy() string {
	return "Online"
}

func (cp CPv1) VolStateDegraded() string {
	return "Degraded"
}

func (cp CPv1) ChildStateOnline() string {
	return "Online"
}

func (cp CPv1) ChildStateDegraded() string {
	return "Degraded"
}

func (cp CPv1) ChildStateUnknown() string {
	return "Unknown"
}

func (cp CPv1) ChildStateFaulted() string {
	return "Faulted"
}

func (cp CPv1) NexusStateUnknown() string {
	return "Unknown"
}

func (cp CPv1) NexusStateOnline() string {
	return "Online"
}

func (cp CPv1) NexusStateDegraded() string {
	return "Degraded"
}

func (cp CPv1) NexusStateFaulted() string {
	return "Faulted"
}

func (cp CPv1) MspStateOnline() string {
	return "Online"
}

func (cp CPv1) MspGrpcStateToCrdState(mspState int) string {
	switch mspState {
	case 0:
		return "Pending"
	case 1:
		return "Online"
	case 2:
		return "Degraded"
	case 3:
		return "Faulted"
	default:
		return "Offline"
	}
}

func MakeCP(majorVersion int) (CPv1, error) {
	if majorVersion != 1 {
		return CPv1{}, fmt.Errorf("incompatible version %d, expected 1", majorVersion)
	}
	addrs, err := GetMasterNodeIPs()
	logf.Log.Info("Control Plane v1 - using kubectl plugin")
	return CPv1{
		nodeIPAddresses: addrs,
		pluginPath:      fmt.Sprintf("%s/%s", e2e_config.GetConfig().KubectlPluginDir, common.KubectlMayastorPlugin),
	}, err
}

func (cp CPv1) NodeStateOnline() string {
	return "Online"
}

func (cp CPv1) NodeStateOffline() string {
	return "Offline"
}

func (cp CPv1) NodeStateUnknown() string {
	return "Unknown"
}

func (cp CPv1) NodeStateEmpty() string {
	return ""
}
