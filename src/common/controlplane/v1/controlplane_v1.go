package v1

import (
	"fmt"
	"regexp"
)

type CPv1 struct {
	nodeIPAddresses *[]string
}

func (cp CPv1) Version() string {
	return "1.0.0"
}

func (cp CPv1) MajorVersion() int {
	return 1
}

var re = regexp.MustCompile(`(request timed out)`)

func (cp CPv1) IsTimeoutError(err error) bool {
	str := fmt.Sprintf("%v", err)
	frags := re.FindSubmatch([]byte(str))
	return len(frags) == 2 && string(frags[1]) == "request timed out"
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

func MakeCP(addr *[]string) CPv1 {
	return CPv1{
		nodeIPAddresses: addr,
	}
}

func (cp CPv1) NodeStateOnline() string {
	return "Online"
}

func (cp CPv1) NodeStateOffline() string {
	return "Offline"
}
