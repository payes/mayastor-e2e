package k8stest

import (
	. "github.com/onsi/gomega"
	"mayastor-e2e/common"
	MSVCrd "mayastor-e2e/common/msv/crd"
	MSVKubectl "mayastor-e2e/common/msv/mayastor_kubectl"
	"sync"
)

var once sync.Once

// DO NOT ACCESS DIRECTLY, use GetMsvIfc
var ifc common.MayastorVolumeInterface

func GetMsvIfc() common.MayastorVolumeInterface {
	once.Do(func() {
		if common.IsControlPlaneMoac() {
			ifc = MSVCrd.MakeCrMsv()
		} else if common.IsControlPlaneMcp() {
			ifc = MSVKubectl.MakeCpMsv(GetMayastorNodeIPAddresses())
		} else {
			Expect(false).To(BeTrue())
		}
	})

	return ifc
}

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*common.MayastorVolume, error) {
	return GetMsvIfc().GetMSV(uuid)
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	return GetMsvIfc().GetMsvNodes(uuid)
}

func DeleteMsv(volName string) error {
	return GetMsvIfc().DeleteMsv(volName)
}

func ListMsvs() ([]common.MayastorVolume, error) {
	return GetMsvIfc().ListMsvs()
}

func SetMsvReplicaCount(uuid string, replicaCount int) error {
	return GetMsvIfc().SetMsvReplicaCount(uuid, replicaCount)
}

func GetMsvState(uuid string) (string, error) {
	return GetMsvIfc().GetMsvState(uuid)
}

func GetMsvReplicas(volName string) ([]common.Replica, error) {
	return GetMsvIfc().GetMsvReplicas(volName)
}

func GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	return GetMsvIfc().GetMsvNexusChildren(volName)
}

func GetMsvNexusState(uuid string) (string, error) {
	return GetMsvIfc().GetMsvNexusState(uuid)
}

func IsMsvPublished(uuid string) bool {
	return GetMsvIfc().IsMsvPublished(uuid)
}

func IsMsvDeleted(uuid string) bool {
	return GetMsvIfc().IsMsvDeleted(uuid)
}

func CheckForMsvs() (bool, error) {
	return GetMsvIfc().CheckForMsvs()
}

func CheckAllMsvsAreHealthy() error {
	return GetMsvIfc().CheckAllMsvsAreHealthy()
}
