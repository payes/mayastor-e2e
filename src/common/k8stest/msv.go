package k8stest

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
)

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*common.MayastorVolume, error) {
	return controlplane.GetMSV(uuid)
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	return controlplane.GetMsvNodes(uuid)
}

func DeleteMsv(volName string) error {
	return controlplane.DeleteMsv(volName)
}

func ListMsvs() ([]common.MayastorVolume, error) {
	return controlplane.ListMsvs()
}

func SetMsvReplicaCount(uuid string, replicaCount int) error {
	return controlplane.SetMsvReplicaCount(uuid, replicaCount)
}

func GetMsvState(uuid string) (string, error) {
	return controlplane.GetMsvState(uuid)
}

func GetMsvReplicas(volName string) ([]common.Replica, error) {
	return controlplane.GetMsvReplicas(volName)
}

func GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	return controlplane.GetMsvNexusChildren(volName)
}

func GetMsvNexusState(uuid string) (string, error) {
	return controlplane.GetMsvNexusState(uuid)
}

func IsMsvPublished(uuid string) bool {
	return controlplane.IsMsvPublished(uuid)
}

func IsMsvDeleted(uuid string) bool {
	return controlplane.IsMsvDeleted(uuid)
}

func CheckForMsvs() (bool, error) {
	return controlplane.CheckForMsvs()
}

func CheckAllMsvsAreHealthy() error {
	return controlplane.CheckAllMsvsAreHealthy()
}
