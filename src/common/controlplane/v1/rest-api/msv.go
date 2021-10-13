package rest_api

// Utility functions for Mayastor CRDs
import (
	"fmt"
	"mayastor-e2e/common"
	openapiClient "mayastor-e2e/generated/openapi"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func (cp CPv1RestApi) GetMSV(uuid string) (*common.MayastorVolume, error) {
	vol, err, _ := cp.oa.getVolume(uuid)

	if err != nil {
		return nil, fmt.Errorf("GetMSV: %v", err)
	}

	// still being created
	volSpec, ok := vol.GetSpecOk()
	if !ok {
		logf.Log.Info("GetMSV !GetSpecOk()", "uuid", uuid)
		return nil, nil
	}

	volState, ok := vol.GetStateOk()
	if !ok {
		logf.Log.Info("GetMSV !vol.GetStateOk()", "uuid", uuid)
		return nil, nil
	}

	if volState.GetStatus() == openapiClient.VOLUMESTATUS_UNKNOWN {
		return nil, fmt.Errorf("GetMSV: state not defined, got msv.Status=\"%v\"", vol.GetState().Status)
	}

	if volSpec.GetNumReplicas() < 1 {
		return nil, fmt.Errorf("GetMsv msv.Spec.NumReplicas=\"%v\"", volSpec.GetNumReplicas())
	}

	nexus, ok := volState.GetTargetOk()
	if ok {
		if len(nexus.Children) < 1 {
			return nil, fmt.Errorf("GetMSV: nexus children =\"%v\"", nexus.Children)
		}
	}

	msv := cp.oa.volToMsv(vol)
	return &msv, nil
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func (cp CPv1RestApi) GetMsvNodes(uuid string) (string, []string) {
	var nexusNode string
	var replicaNodes []string

	vol, err, _ := cp.oa.getVolume(uuid)
	if err == nil {
		nexusNode = vol.State.Target.Node
		for _, replica := range vol.State.ReplicaTopology {
			replicaNodes = append(replicaNodes, replica.GetNode())
		}
	}
	return nexusNode, replicaNodes
}

func (cp CPv1RestApi) DeleteMsv(uuid string) error {
	err, _ := cp.oa.deleteVolume(uuid)
	return err
}

func (cp CPv1RestApi) ListMsvs() ([]common.MayastorVolume, error) {
	vols, err, _ := cp.oa.getVolumes()
	if err != nil {
		return nil, fmt.Errorf("ListMsvs: %v", err)
	}

	var msvs []common.MayastorVolume
	if err == nil {
		for _, vol := range vols {
			msvs = append(msvs, cp.oa.volToMsv(vol))
		}
	}
	return msvs, err
}

func (cp CPv1RestApi) SetMsvReplicaCount(uuid string, replicaCount int) error {
	err, _ := cp.oa.putReplicaCount(uuid, replicaCount)
	return err
}

func (cp CPv1RestApi) GetMsvState(uuid string) (string, error) {
	vol, err, _ := cp.oa.getVolume(uuid)
	var volState string

	if err == nil {
		volState = string(vol.State.GetStatus())
	}

	return volState, nil
}

func (cp CPv1RestApi) GetMsvReplicas(uuid string) ([]common.Replica, error) {
	var replicas []common.Replica

	vol, err, _ := cp.oa.getVolume(uuid)
	if err == nil {
		for rUuid, rt := range vol.State.ReplicaTopology {
			r, _, _ := cp.oa.getReplica(rUuid)
			replicas = append(replicas, common.Replica{
				Node:    rt.GetNode(),
				Offline: rt.GetState() != "Online",
				Pool:    rt.GetPool(),
				Uri:     r.Uri,
			})
		}
	}
	return replicas, err
}

func (cp CPv1RestApi) GetMsvNexusChildren(uuid string) ([]common.NexusChild, error) {
	var children []common.NexusChild

	vol, err, _ := cp.oa.getVolume(uuid)
	if err == nil {
		for _, child := range vol.State.Target.Children {
			children = append(children,
				common.NexusChild{
					State: string(child.GetState()),
					Uri:   child.GetUri(),
				})
		}
	}
	return children, err
}

func (cp CPv1RestApi) GetMsvNexusState(uuid string) (string, error) {
	var nexusState string

	vol, err, _ := cp.oa.getVolume(uuid)
	if err == nil {
		nexusState = string(vol.State.Target.GetState())
	}
	return nexusState, err
}

func (cp CPv1RestApi) IsMsvPublished(uuid string) bool {
	vol, err, _ := cp.oa.getVolume(uuid)
	if err == nil {
		return vol.Spec.Target.Node != ""
	}
	return false
}

func (cp CPv1RestApi) IsMsvDeleted(uuid string) bool {
	_, err, responseStatusCode := cp.oa.getVolume(uuid)
	return err != nil && responseStatusCode == 404
}

func (cp CPv1RestApi) CheckForMsvs() (bool, error) {
	vols, err, _ := cp.oa.getVolumes()
	return err == nil && vols != nil && len(vols) != 0, err
}

func (cp CPv1RestApi) CheckAllMsvsAreHealthy() error {
	allHealthy := true
	vols, err, _ := cp.oa.getVolumes()

	if err == nil {
		for _, vol := range vols {
			if vol.State.GetStatus() != openapiClient.VOLUMESTATUS_ONLINE {
				allHealthy = false
				logf.Log.Info("CheckAllMsvsAreHealthy", "vol", vol)
			}
		}
		if !allHealthy {
			err = fmt.Errorf("all MSVs were not healthy")
		}
	}

	return err
}
