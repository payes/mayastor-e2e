package k8stest

// Utility functions for Mayastor CRDs
import (
	"fmt"
	"mayastor-e2e/common/custom_resources"
	"strings"

	"mayastor-e2e/common/custom_resources/api/types/v1alpha1"

	. "github.com/onsi/gomega"
)

type MayastorVolumeDetails struct {
	Spec   v1alpha1.MayastorVolumeSpec   `json:"spec"`
	Status v1alpha1.MayastorVolumeStatus `json:"status"`
}

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*MayastorVolumeDetails, error) {
	if !IsControlPlaneMcp() {
		msv, err := custom_resources.GetMsVol(uuid)
		if err != nil {
			return nil, fmt.Errorf("GetMSV: %v", err)
		}

		// pending means still being created
		if msv.Status.State == "pending" {
			return nil, nil
		}

		//logf.Log.Info("GetMSV", "msv", msv)
		// Note: msVol.Node can be unassigned here if the volume is not mounted
		if msv.Status.State == "" {
			return nil, fmt.Errorf("GetMSV: state not defined, got msv.Status=\"%v\"", msv.Status)
		}

		if len(msv.Status.Replicas) < 1 {
			return nil, fmt.Errorf("GetMSV: msv.Status.Replicas=\"%v\"", msv.Status.Replicas)
		}
		return &MayastorVolumeDetails{
			Spec:   msv.Spec,
			Status: msv.Status,
		}, nil
	} else {
		msv, err := GetMayastorVolume(uuid)
		if err != nil {
			return nil, fmt.Errorf("GetMayastorVolume: %v", err)
		}

		// pending means still being created
		if strings.ToLower(msv.State.Status) != "pending" {
			return nil, nil
		}
		if msv.Spec.Num_replicas < 1 {
			return nil, fmt.Errorf("GetMayastorVolume: msv.Status.Replicas=\"%v\"", msv.Spec.Num_replicas)
		}
		return GetMSVDetailsFromPlugin(msv), nil
	}
}

// Cross fit msv details from kubectl plugin to msv crd
func GetMSVDetailsFromPlugin(msv *MayastorVolume) *MayastorVolumeDetails {
	var ms_volume *MayastorVolumeDetails
	ms_volume.Spec.Protocol = msv.Spec.Protocol
	ms_volume.Spec.RequiredBytes = int(msv.Spec.Size)
	ms_volume.Spec.ReplicaCount = msv.Spec.Num_replicas
	ms_volume.Status.State = msv.Spec.State
	ms_volume.Status.Size = msv.Spec.Size
	for _, children := range msv.State.Child.Children {
		ms_volume.Status.Nexus.Children = append(ms_volume.Status.Nexus.Children,
			v1alpha1.NexusChild{
				State: children.State,
				Uri:   children.Uri,
			})
	}
	return ms_volume
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	msv, err := custom_resources.GetMsVol(uuid)
	Expect(err).ToNot(HaveOccurred())
	node := msv.Status.Nexus.Node
	replicas := make([]string, len(msv.Status.Replicas))
	for ix, r := range msv.Status.Replicas {
		replicas[ix] = r.Node
	}
	return node, replicas
}
