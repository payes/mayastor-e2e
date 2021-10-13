package controlplane

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane/v0"
	v1cpkp "mayastor-e2e/common/controlplane/v1"
	v1oa "mayastor-e2e/common/controlplane/v1/cp-openapi"
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

type cpInterface interface {
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

	MspGrpcStateToCrdState(int) string

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

var ifc cpInterface

var once sync.Once

func getControlPlane() cpInterface {
	once.Do(func() {
		if common.IsControlPlaneMoac() {
			ifc = v0.MakeCP()
		}
		if common.IsControlPlaneMcp() {
			_ = v1cpkp.MakeCP(&nodeIpAddresses)
			ifc = v1oa.MakeCP(&nodeIpAddresses)
		}
		if ifc == nil {
			panic("failed to set control plane object")
		}
	})
	return ifc
}

func VolStateHealthy() string {
	return getControlPlane().VolStateHealthy()
}

func VolStateDegraded() string {
	return getControlPlane().VolStateDegraded()
}

func ChildStateUnknown() string {
	return getControlPlane().ChildStateUnknown()
}

func ChildStateOnline() string {
	return getControlPlane().ChildStateOnline()
}

func ChildStateDegraded() string {
	return getControlPlane().ChildStateDegraded()
}

func ChildStateFaulted() string {
	return getControlPlane().ChildStateFaulted()
}

func NexusStateUnknown() string {
	return getControlPlane().NexusStateUnknown()
}

func NexusStateOnline() string {
	return getControlPlane().NexusStateOnline()
}

func NexusStateDegraded() string {
	return getControlPlane().NexusStateDegraded()
}

func NexusStateFaulted() string {
	return getControlPlane().NexusStateFaulted()
}

func MspGrpcStateToCrdState(mspState int) string {
	return getControlPlane().MspGrpcStateToCrdState(mspState)
}

//FIXME: MSV These functions are only guaranteed to
// work correctly if invoked from k8stest/msv.go
// which ensures that necessary setup functions
// have been called
// The issue is that for control plane v1 we need
// node addresses and the k8stest pkg provides that.

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*common.MayastorVolume, error) {
	return getControlPlane().GetMSV(uuid)
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	return getControlPlane().GetMsvNodes(uuid)
}

func DeleteMsv(volName string) error {
	return getControlPlane().DeleteMsv(volName)
}

func ListMsvs() ([]common.MayastorVolume, error) {
	return getControlPlane().ListMsvs()
}

func SetMsvReplicaCount(uuid string, replicaCount int) error {
	return getControlPlane().SetMsvReplicaCount(uuid, replicaCount)
}

func GetMsvState(uuid string) (string, error) {
	return getControlPlane().GetMsvState(uuid)
}

func GetMsvReplicas(volName string) ([]common.Replica, error) {
	return getControlPlane().GetMsvReplicas(volName)
}

func GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	return getControlPlane().GetMsvNexusChildren(volName)
}

func GetMsvNexusState(uuid string) (string, error) {
	return getControlPlane().GetMsvNexusState(uuid)
}

func IsMsvPublished(uuid string) bool {
	return getControlPlane().IsMsvPublished(uuid)
}

func IsMsvDeleted(uuid string) bool {
	return getControlPlane().IsMsvDeleted(uuid)
}

func CheckForMsvs() (bool, error) {
	return getControlPlane().CheckForMsvs()
}

func CheckAllMsvsAreHealthy() error {
	return getControlPlane().CheckAllMsvsAreHealthy()
}
