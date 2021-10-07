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
func ensure_msv() {
	once.Do(func() {
		controlplane.SetIpNodeAddresses(GetMayastorNodeIPAddresses())
	})
}

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*common.MayastorVolume, error) {
	ensure_msv()
	return controlplane.GetMSV(uuid)
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	ensure_msv()
	return controlplane.GetMsvNodes(uuid)
}

func DeleteMsv(volName string) error {
	ensure_msv()
	return controlplane.DeleteMsv(volName)
}

func ListMsvs() ([]common.MayastorVolume, error) {
	ensure_msv()
	return controlplane.ListMsvs()
}

func SetMsvReplicaCount(uuid string, replicaCount int) error {
	ensure_msv()
	return controlplane.SetMsvReplicaCount(uuid, replicaCount)
}

func GetMsvState(uuid string) (string, error) {
	ensure_msv()
	return controlplane.GetMsvState(uuid)
}

func GetMsvReplicas(volName string) ([]common.Replica, error) {
	ensure_msv()
	return controlplane.GetMsvReplicas(volName)
}

func GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	ensure_msv()
	return controlplane.GetMsvNexusChildren(volName)
}

func GetMsvNexusState(uuid string) (string, error) {
	ensure_msv()
	return controlplane.GetMsvNexusState(uuid)
}

func IsMsvPublished(uuid string) bool {
	ensure_msv()
	return controlplane.IsMsvPublished(uuid)
}

func IsMsvDeleted(uuid string) bool {
	ensure_msv()
	return controlplane.IsMsvDeleted(uuid)
}

func CheckForMsvs() (bool, error) {
	ensure_msv()
	return controlplane.CheckForMsvs()
}

func CheckAllMsvsAreHealthy() error {
	ensure_msv()
	return controlplane.CheckAllMsvsAreHealthy()
}
