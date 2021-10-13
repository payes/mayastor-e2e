package rest_api

import (
	"fmt"
	"mayastor-e2e/generated/openapi"
	"regexp"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	cpV1 "mayastor-e2e/common/controlplane/v1"
)

type CPv1RestApi struct {
	oa restApiClient
}

func (cp CPv1RestApi) Version() string {
	return "1.0.0"
}

func (cp CPv1RestApi) MajorVersion() int {
	return 1
}

func (cp CPv1RestApi) VolStateHealthy() string {
	return string(openapi.VOLUMESTATUS_ONLINE)
}

func (cp CPv1RestApi) VolStateDegraded() string {
	return string(openapi.VOLUMESTATUS_DEGRADED)
}

func (cp CPv1RestApi) ChildStateOnline() string {
	return string(openapi.CHILDSTATE_ONLINE)
}

func (cp CPv1RestApi) ChildStateDegraded() string {
	return string(openapi.CHILDSTATE_DEGRADED)
}

func (cp CPv1RestApi) ChildStateUnknown() string {
	return string(openapi.CHILDSTATE_UNKNOWN)
}

func (cp CPv1RestApi) ChildStateFaulted() string {
	return string(openapi.CHILDSTATE_FAULTED)
}

func (cp CPv1RestApi) NexusStateUnknown() string {
	return string(openapi.CHILDSTATE_UNKNOWN)
}

func (cp CPv1RestApi) NexusStateOnline() string {
	return string(openapi.NEXUSSTATE_ONLINE)
}

func (cp CPv1RestApi) NexusStateDegraded() string {
	return string(openapi.NEXUSSTATE_DEGRADED)
}

func (cp CPv1RestApi) NexusStateFaulted() string {
	return string(openapi.NEXUSSTATE_FAULTED)
}

func (cp CPv1RestApi) MspStateOnline() string {
	return string(openapi.POOLSTATUS_ONLINE)
}

func (cp CPv1RestApi) MspGrpcStateToCrdState(mspState int) string {
	switch mspState {
	case 0:
		// FIXME: unmappable state. All known states are mapped below.
		return "Pending"
	case 1:
		return string(openapi.POOLSTATUS_ONLINE)
	case 2:
		return string(openapi.POOLSTATUS_DEGRADED)
	case 3:
		return string(openapi.POOLSTATUS_FAULTED)
	default:
		return string(openapi.POOLSTATUS_UNKNOWN)
	}
}

func (cp CPv1RestApi) NodeStateOffline() string {
	return string(openapi.NODESTATUS_OFFLINE)
}

func (cp CPv1RestApi) NodeStateOnline() string {
	return string(openapi.NODESTATUS_ONLINE)
}

func (cp CPv1RestApi) NodeStateUnknown() string {
	return string(openapi.NODESTATUS_UNKNOWN)
}

func (cp CPv1RestApi) NodeStateEmpty() string {
	return ""
}

func MakeCP(majorVersion int) (CPv1RestApi, error) {
	if majorVersion != 1 {
		return CPv1RestApi{}, fmt.Errorf("incompatible major version %d, expected 1", majorVersion)
	}
	addrs, err := cpV1.GetMasterNodeIPs()
	cp := CPv1RestApi{
		oa: makeClient(addrs),
	}
	logf.Log.Info("Control Plane v1 - Rest API")
	return cp, err
}

var re = regexp.MustCompile(`(statusCode=408)`)

func (cp CPv1RestApi) IsTimeoutError(err error) bool {
	str := fmt.Sprintf("%v", err)
	frags := re.FindSubmatch([]byte(str))
	return len(frags) == 2 && string(frags[1]) == "statusCode=408"
}
