package k8stest

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
	"sync"
)

var once sync.Once

// FIXME: MSV
// This two-step initialisation of the control plane interface
// ensures that the node IP addresses are set for control plane
// before any MSV related call is made.
// Facilitates abstraction without introducing a dependency on k8stest
func EnsureNodeAddressesAreSet() {
	once.Do(func() {
		controlplane.SetIpNodeAddresses(GetMayastorNodeIPAddresses())
	})
}

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*common.MayastorVolume, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.GetMSV(uuid)
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	EnsureNodeAddressesAreSet()
	return controlplane.GetMsvNodes(uuid)
}

func DeleteMsv(volName string) error {
	EnsureNodeAddressesAreSet()
	return controlplane.DeleteMsv(volName)
}

func ListMsvs() ([]common.MayastorVolume, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.ListMsvs()
}

func SetMsvReplicaCount(uuid string, replicaCount int) error {
	EnsureNodeAddressesAreSet()
	return controlplane.SetMsvReplicaCount(uuid, replicaCount)
}

func GetMsvState(uuid string) (string, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.GetMsvState(uuid)
}

func GetMsvReplicas(volName string) ([]common.Replica, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.GetMsvReplicas(volName)
}

func GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.GetMsvNexusChildren(volName)
}

func GetMsvNexusState(uuid string) (string, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.GetMsvNexusState(uuid)
}

func IsMsvPublished(uuid string) bool {
	EnsureNodeAddressesAreSet()
	return controlplane.IsMsvPublished(uuid)
}

func IsMsvDeleted(uuid string) bool {
	EnsureNodeAddressesAreSet()
	return controlplane.IsMsvDeleted(uuid)
}

func CheckForMsvs() (bool, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.CheckForMsvs()
}

func CheckAllMsvsAreHealthy() error {
	EnsureNodeAddressesAreSet()
	return controlplane.CheckAllMsvsAreHealthy()
}
