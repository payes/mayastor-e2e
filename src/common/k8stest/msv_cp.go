package k8stest

// Utility functions for Mayastor control plane volume
import (
	"fmt"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/gomega"
)

type CpMsv struct{}

func cpVolumeToMsv(cpMsv *MayastorCpVolume) MayastorVolume {
	var nexusChildren []NexusChild
	for _, children := range cpMsv.State.Child.Children {
		nexusChildren = append(nexusChildren, NexusChild{
			State: children.State,
			Uri:   children.Uri,
		})
	}

	// FIX-ME !!!!!!!!!!!!!!!!
	// var replicas []Replica
	// for _, cpReplica := range cpMsv.State.Replicas {
	// 	replicas = append(replicas, Replica(cpReplica))
	// }

	return MayastorVolume{
		Name: cpMsv.Spec.Uuid,
		Spec: MayastorVolumeSpec{
			Protocol:      cpMsv.Spec.Protocol,
			ReplicaCount:  cpMsv.Spec.Num_replicas,
			RequiredBytes: int(cpMsv.Spec.Size),
		},
		Status: MayastorVolumeStatus{
			Nexus: Nexus{
				Children:  nexusChildren,
				DeviceUri: cpMsv.State.Child.DeviceUri,
				Node:      cpMsv.State.Child.Node,
				State:     cpMsv.State.Child.State,
			},
			// FIX-ME !!!!!!!!!!!!!!
			//Replicas: replicas,
			Size:  cpMsv.State.Size,
			State: cpMsv.State.Status,
		},
	}
}

// GetMSV Get pointer to a mayastor control plane volume
// returns nil and no error if the msv is in pending state.
func (mc CpMsv) getMSV(uuid string) (*MayastorVolume, error) {
	cpMsv, err := GetMayastorVolume(uuid)
	if err != nil {
		return nil, fmt.Errorf("GetMSV: %v", err)
	}

	// pending means still being created
	if cpMsv.State.Status == "pending" {
		return nil, nil
	}

	//logf.Log.Info("GetMSV", "msv", msv)
	// Note: msVol.Node can be unassigned here if the volume is not mounted
	if cpMsv.State.Status == "" {
		return nil, fmt.Errorf("GetMSV: state not defined, got msv.Status=\"%v\"", cpMsv.State)
	}

	if cpMsv.Spec.Num_replicas < 1 {
		return nil, fmt.Errorf("GetMSV: msv.Spec.Num_replicas=\"%v\"", cpMsv.Spec.Num_replicas)
	}

	msv := cpVolumeToMsv(cpMsv)
	return &msv, nil
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
func (mc CpMsv) getMsvNodes(uuid string) (string, []string) {
	msv, err := GetMayastorVolume(uuid)
	Expect(err).ToNot(HaveOccurred())
	node := msv.State.Child.Node
	replicas := make([]string, msv.Spec.Num_replicas)
	//FIX-ME !!!!!!!!!!!!!
	// for ix, r := range msv.Status.Replicas {
	// 	replicas[ix] = r.Node
	// }
	return node, replicas
}

func (mc CpMsv) listMsvs() ([]MayastorVolume, error) {
	var msvs []MayastorVolume
	list, err := ListMayastorVolumes()
	if err == nil {
		for _, item := range list {
			msvs = append(msvs, cpVolumeToMsv(&item))
		}
	}
	return msvs, err
}

func (mc CpMsv) setMsvReplicaCount(uuid string, replicaCount int) error {
	err := ScaleMayastorVolume(uuid, replicaCount)
	logf.Log.Info("ScaleMayastorVolume", "Num_replicas", replicaCount)
	return err
}

func (mc CpMsv) getMsvState(uuid string) (string, error) {
	return GetMayastorVolumeState(uuid)
}

func (mc CpMsv) getMsvReplicas(volName string) ([]Replica, error) {
	var replicas []Replica
	// FIX_ME !!!!!!!!!!!!!!!!!
	return replicas, nil
}

func (mc CpMsv) getMsvNexusChildren(volName string) ([]NexusChild, error) {
	var children []NexusChild
	cpVolumeChildren, err := GetMayastorVolumeChildren(volName)
	if err == nil {
		for _, child := range cpVolumeChildren {
			children = append(children, NexusChild{
				State: child.State,
				Uri:   child.Uri,
			})
		}
	}
	return children, err
}

func (mc CpMsv) getMsvNexusState(uuid string) (string, error) {
	return GetMayastorVolumeChildState(uuid)
}

func (mc CpMsv) isMsvPublished(uuid string) bool {
	return IsMmayastorVolumePublished(uuid)
}

func (mc CpMsv) isMsvDeleted(uuid string) bool {
	return IsMayastorVolumeDeleted(uuid)
}

func (mc CpMsv) checkForMsvs() (bool, error) {
	return CheckForMayastorVolumes()
}

func (mc CpMsv) checkAllMsvsAreHealthy() error {
	return CheckAllMayastorVolumesAreHealthy()
}

func (mc CpMsv) deleteMsv(volName string) error {
	return fmt.Errorf("Delete of mayastor volume not supported")
}

func MakeCpMsv() CpMsv {
	return CpMsv{}
}
