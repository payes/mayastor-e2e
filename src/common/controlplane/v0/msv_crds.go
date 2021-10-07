package v0

// Utility functions for Mayastor CRDs
import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	v1alpha1Api "mayastor-e2e/common/custom_resources/api/types/v1alpha1"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/gomega"
)

func crdToMsv(crdMsv *v1alpha1Api.MayastorVolume) common.MayastorVolume {
	var nexusChildren []common.NexusChild
	for _, inChild := range crdMsv.Status.Nexus.Children {
		nexusChildren = append(nexusChildren, common.NexusChild(inChild))
	}
	var replicas []common.Replica
	for _, crdReplica := range crdMsv.Status.Replicas {
		replicas = append(replicas, common.Replica(crdReplica))
	}

	return common.MayastorVolume{
		Name: crdMsv.GetName(),
		Spec: common.MayastorVolumeSpec{
			Protocol:      crdMsv.Spec.Protocol,
			ReplicaCount:  crdMsv.Spec.ReplicaCount,
			RequiredBytes: crdMsv.Spec.RequiredBytes,
		},
		Status: common.MayastorVolumeStatus{
			Nexus: common.Nexus{
				Children:  nexusChildren,
				DeviceUri: crdMsv.Status.Nexus.DeviceUri,
				Node:      crdMsv.Status.Nexus.Node,
				State:     crdMsv.Status.Nexus.State,
				Uuid:      crdMsv.GetName(),
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
func (cp CPv0p8) GetMSV(uuid string) (*common.MayastorVolume, error) {
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
func (cp CPv0p8) GetMsvNodes(uuid string) (string, []string) {
	msv, err := custom_resources.GetMsVol(uuid)
	Expect(err).ToNot(HaveOccurred())
	node := msv.Status.Nexus.Node
	replicas := make([]string, len(msv.Status.Replicas))
	for ix, r := range msv.Status.Replicas {
		replicas[ix] = r.Node
	}
	return node, replicas
}

func (cp CPv0p8) DeleteMsv(volName string) error {
	return custom_resources.CRD_DeleteMsVol(volName)
}

func (cp CPv0p8) ListMsvs() ([]common.MayastorVolume, error) {
	var msvs []common.MayastorVolume
	crs, err := custom_resources.CRD_ListMsVols()
	if err == nil {
		for _, cr := range crs {
			msvs = append(msvs, crdToMsv(&cr))
		}
	}
	return msvs, err
}

func (cp CPv0p8) SetMsvReplicaCount(uuid string, replicaCount int) error {
	err := custom_resources.CRD_UpdateMsVolReplicaCount(uuid, replicaCount)
	logf.Log.Info("SetMsvReplicaCount", "ReplicaCount", replicaCount)
	return err
}

func (cp CPv0p8) GetMsvState(uuid string) (string, error) {
	return custom_resources.CRD_GetMsVolState(uuid)
}

func (cp CPv0p8) GetMsvReplicas(volName string) ([]common.Replica, error) {
	var replicas []common.Replica
	crdReplicas, err := custom_resources.CRD_GetMsVolReplicas(volName)
	if err == nil {
		for _, cr := range crdReplicas {
			replicas = append(replicas, common.Replica(cr))
		}
	}
	return replicas, err
}

func (cp CPv0p8) GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	var children []common.NexusChild
	crdChildren, err := custom_resources.CRD_GetMsVolNexusChildren(volName)
	if err == nil {
		for _, cr := range crdChildren {
			children = append(children, common.NexusChild(cr))
		}
	}
	return children, err
}

func (cp CPv0p8) GetMsvNexusState(uuid string) (string, error) {
	return custom_resources.CRD_GetMsVolNexusState(uuid)
}

func (cp CPv0p8) IsMsvPublished(uuid string) bool {
	return custom_resources.CRD_IsMsVolPublished(uuid)
}

func (cp CPv0p8) IsMsvDeleted(uuid string) bool {
	return custom_resources.CRD_IsMsVolDeleted(uuid)
}

func (cp CPv0p8) CheckForMsvs() (bool, error) {
	return custom_resources.CRD_CheckForMsVols()
}

func (cp CPv0p8) CheckAllMsvsAreHealthy() error {
	return custom_resources.CRD_CheckAllMsVolsAreHealthy()
}
