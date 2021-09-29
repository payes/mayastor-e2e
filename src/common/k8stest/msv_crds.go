package k8stest

// Utility functions for Mayastor CRDs
import (
	"fmt"
	"mayastor-e2e/common/custom_resources"
	v1alpha1Api "mayastor-e2e/common/custom_resources/api/types/v1alpha1"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/gomega"
)

type CrMsv struct{}

func crdToMsv(crdMsv *v1alpha1Api.MayastorVolume) MayastorVolume {
	var nexusChildren []NexusChild
	for _, inChild := range crdMsv.Status.Nexus.Children {
		nexusChildren = append(nexusChildren, NexusChild(inChild))
	}
	var replicas []Replica
	for _, crdReplica := range crdMsv.Status.Replicas {
		replicas = append(replicas, Replica(crdReplica))
	}

	return MayastorVolume{
		Name: crdMsv.GetName(),
		Spec: MayastorVolumeSpec{
			Protocol:      crdMsv.Spec.Protocol,
			ReplicaCount:  crdMsv.Spec.ReplicaCount,
			RequiredBytes: crdMsv.Spec.RequiredBytes,
		},
		Status: MayastorVolumeStatus{
			Nexus: Nexus{
				Children:  nexusChildren,
				DeviceUri: crdMsv.Status.Nexus.DeviceUri,
				Node:      crdMsv.Status.Nexus.Node,
				State:     crdMsv.Status.Nexus.State,
			},
			Reason:   crdMsv.Status.Reason,
			Replicas: replicas,
			Size:     crdMsv.Status.Size,
			State:    crdMsv.Status.State,
		},
	}
}

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func (mc CrMsv) getMSV(uuid string) (*MayastorVolume, error) {
	crdMsv, err := custom_resources.GetMsVol(uuid)
	if err != nil {
		return nil, fmt.Errorf("GetMSV: %v", err)
	}

	// pending means still being created
	if crdMsv.Status.State == "pending" {
		return nil, nil
	}

	//logf.Log.Info("GetMSV", "msv", msv)
	// Note: msVol.Node can be unassigned here if the volume is not mounted
	if crdMsv.Status.State == "" {
		return nil, fmt.Errorf("GetMSV: state not defined, got msv.Status=\"%v\"", crdMsv.Status)
	}

	if len(crdMsv.Status.Replicas) < 1 {
		return nil, fmt.Errorf("GetMSV: msv.Status.Replicas=\"%v\"", crdMsv.Status.Replicas)
	}

	msv := crdToMsv(crdMsv)
	return &msv, nil
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func (mc CrMsv) getMsvNodes(uuid string) (string, []string) {
	msv, err := custom_resources.GetMsVol(uuid)
	Expect(err).ToNot(HaveOccurred())
	node := msv.Status.Nexus.Node
	replicas := make([]string, len(msv.Status.Replicas))
	for ix, r := range msv.Status.Replicas {
		replicas[ix] = r.Node
	}
	return node, replicas
}

func (mc CrMsv) deleteMsv(volName string) error {
	return custom_resources.CRD_DeleteMsVol(volName)
}

func (mc CrMsv) listMsvs() ([]MayastorVolume, error) {
	var msvs []MayastorVolume
	crs, err := custom_resources.CRD_ListMsVols()
	if err == nil {
		for _, cr := range crs {
			msvs = append(msvs, crdToMsv(&cr))
		}
	}
	return msvs, err
}

func (mc CrMsv) setMsvReplicaCount(uuid string, replicaCount int) error {
	err := custom_resources.CRD_UpdateMsVolReplicaCount(uuid, replicaCount)
	logf.Log.Info("SetMsvReplicaCount", "ReplicaCount", replicaCount)
	return err
}

func (mc CrMsv) getMsvState(uuid string) (string, error) {
	return custom_resources.CRD_GetMsVolState(uuid)
}

func (mc CrMsv) getMsvReplicas(volName string) ([]Replica, error) {
	var replicas []Replica
	crdReplicas, err := custom_resources.CRD_GetMsVolReplicas(volName)
	if err == nil {
		for _, cr := range crdReplicas {
			replicas = append(replicas, Replica(cr))
		}
	}
	return replicas, err
}

func (mc CrMsv) getMsvNexusChildren(volName string) ([]NexusChild, error) {
	var children []NexusChild
	crdChildren, err := custom_resources.CRD_GetMsVolNexusChildren(volName)
	if err == nil {
		for _, cr := range crdChildren {
			children = append(children, NexusChild(cr))
		}
	}
	return children, err
}

func (mc CrMsv) getMsvNexusState(uuid string) (string, error) {
	return custom_resources.CRD_GetMsVolNexusState(uuid)
}

func (mc CrMsv) isMsvPublished(uuid string) bool {
	return custom_resources.CRD_IsMsVolPublished(uuid)
}

func (mc CrMsv) isMsvDeleted(uuid string) bool {
	return custom_resources.CRD_IsMsVolDeleted(uuid)
}

func (mc CrMsv) checkForMsvs() (bool, error) {
	return custom_resources.CRD_CheckForMsVols()
}

func (mc CrMsv) checkAllMsvsAreHealthy() error {
	return custom_resources.CRD_CheckAllMsVolsAreHealthy()
}

func MakeCrMsv() CrMsv {
	return CrMsv{}
}
