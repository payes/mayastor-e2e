package k8stest

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/ctlpln"
	"sync"
)

var once sync.Once

// This two-step initialisation of the control plane interface
// ensures that the node IP addresses are set for control plane
// before any MSV related call is made.
// Facilitates abstraction without introducing a dependency on k8stest
func getMsvIfc() ctlpln.ControlPlaneInterface {
	once.Do(func() {
		ctlpln.SetIpNodeAddresses(GetMayastorNodeIPAddresses())
	})

	return ctlpln.GetControlPlane()
}

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*common.MayastorVolume, error) {
	return getMsvIfc().GetMSV(uuid)
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	return getMsvIfc().GetMsvNodes(uuid)
}

func DeleteMsv(volName string) error {
	return getMsvIfc().DeleteMsv(volName)
}

func ListMsvs() ([]common.MayastorVolume, error) {
	return getMsvIfc().ListMsvs()
}

func SetMsvReplicaCount(uuid string, replicaCount int) error {
	return getMsvIfc().SetMsvReplicaCount(uuid, replicaCount)
}

func GetMsvState(uuid string) (string, error) {
	return getMsvIfc().GetMsvState(uuid)
}

func GetMsvReplicas(volName string) ([]common.Replica, error) {
	return getMsvIfc().GetMsvReplicas(volName)
}

func GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	return getMsvIfc().GetMsvNexusChildren(volName)
}

func GetMsvNexusState(uuid string) (string, error) {
	return getMsvIfc().GetMsvNexusState(uuid)
}

func IsMsvPublished(uuid string) bool {
	return getMsvIfc().IsMsvPublished(uuid)
}

func IsMsvDeleted(uuid string) bool {
	return getMsvIfc().IsMsvDeleted(uuid)
}

func CheckForMsvs() (bool, error) {
	return getMsvIfc().CheckForMsvs()
}

func CheckAllMsvsAreHealthy() error {
	return getMsvIfc().CheckAllMsvsAreHealthy()
}
