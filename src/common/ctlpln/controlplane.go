package ctlpln

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/ctlpln/v0"
	"mayastor-e2e/common/ctlpln/v1"
	"sync"
)

// To circumvent circular dependency on k8stest and avoid hidden initialisation dependencies
// declare an array of nodeIpAddresses the address of which is passed to relevant control
// plane interface implementations as required.
// The array can be updated by a subsequent calls
// This works because we only ever use a single instance of a control plane interface
var nodeIpAddresses []string

func SetIpNodeAddresses(address []string) {
	nodeIpAddresses = nodeIpAddresses[:0]
	nodeIpAddresses = append(nodeIpAddresses, address...)
}

type ControlPlaneInterface interface {
	// Version

	MajorVersion() int
	Version() string

	// Resource state strings abstraction

	VolStateHealthy() string
	VolStateDegraded() string
	ChildStateUnknown() string
	ChildStateOnline() string
	ChildStateDegraded() string
	ChildStateFaulted() string
	NexusStateUnknown() string
	NexusStateOnline() string
	NexusStateDegraded() string
	NexusStateFaulted() string

	// MSV abstraction

	GetMSV(uuid string) (*common.MayastorVolume, error)
	GetMsvNodes(uuid string) (string, []string)
	DeleteMsv(volName string) error
	ListMsvs() ([]common.MayastorVolume, error)
	SetMsvReplicaCount(uuid string, replicaCount int) error
	GetMsvState(uuid string) (string, error)
	GetMsvReplicas(volName string) ([]common.Replica, error)
	GetMsvNexusChildren(volName string) ([]common.NexusChild, error)
	GetMsvNexusState(uuid string) (string, error)
	IsMsvPublished(uuid string) bool
	IsMsvDeleted(uuid string) bool
	CheckForMsvs() (bool, error)
	CheckAllMsvsAreHealthy() error
}

var ifc ControlPlaneInterface

var once sync.Once

func GetControlPlane() ControlPlaneInterface {
	once.Do(func() {
		if common.IsControlPlaneMoac() {
			ifc = v0.MakeCP()
		}
		if common.IsControlPlaneMcp() {
			ifc = v1.MakeCP(&nodeIpAddresses)
		}
		if ifc == nil {
			panic("failed to set control plane object")
		}
	})
	return ifc
}

func VolStateHealthy() string {
	return GetControlPlane().VolStateHealthy()
}

func VolStateDegraded() string {
	return GetControlPlane().VolStateDegraded()
}

func ChildStateOnline() string {
	return GetControlPlane().ChildStateOnline()
}

func ChildStateDegraded() string {
	return GetControlPlane().ChildStateDegraded()
}

func ChildStateUnknown() string {
	return GetControlPlane().ChildStateUnknown()
}

func ChildStateFaulted() string {
	return GetControlPlane().ChildStateFaulted()
}
