package cp_openapi

import (
	"context"
	"mayastor-e2e/common"
	openapiClient "mayastor-e2e/generated/openapi"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type oaClient struct {
	nodes   *[]string
	cfg     []*openapiClient.Configuration
	clients []*openapiClient.APIClient
}

func makeClient(nodeIPs *[]string) oaClient {
	oac := oaClient{
		nodes: nodeIPs,
	}
	return oac
}

// ensure that delayed initialisation is performed
// delayed init is required because at creation time, dependencies cycle restrictions (k8stest)
// make it impossible to initialising the set of clients then.
func (oac *oaClient) ensure() {
	if len(oac.clients) == 0 {
		for _, node := range *oac.nodes {
			cfg := openapiClient.NewConfiguration()
			cfg.Host = node + ":30011"
			cfg.Scheme = `http`
			client := openapiClient.NewAPIClient(cfg)
			oac.clients = append(oac.clients, client)
			oac.cfg = append(oac.cfg, cfg)
		}
	}
}

func (oac oaClient) getReplica(uuid string) (openapiClient.Replica, error) {
	var replica openapiClient.Replica
	var resp *http.Response
	var err error

	oac.ensure()
	for _, client := range oac.clients {
		req := client.ReplicasApi.GetReplica(context.TODO(), uuid)
		replica, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getReplica", "statusCode", resp.StatusCode, "resp", resp)
		} else {
			break
		}
	}
	return replica, err
}

func (oac oaClient) getVolume(uuid string) (openapiClient.Volume, error, int) {
	var volume openapiClient.Volume
	var resp *http.Response
	var err error

	oac.ensure()
	for _, client := range oac.clients {
		req := client.VolumesApi.GetVolume(context.TODO(), uuid)
		volume, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getVolume", "statusCode", resp.StatusCode, "resp", resp)
		} else {
			break
		}
	}
	log.Log.Info("getVolume", "volume", volume, "err", err, "statusCode", resp.StatusCode)
	return volume, err, resp.StatusCode
}

func (oac oaClient) getVolumes() ([]openapiClient.Volume, error) {
	var volumes []openapiClient.Volume
	var resp *http.Response
	var err error

	oac.ensure()
	for _, client := range oac.clients {
		req := client.VolumesApi.GetVolumes(context.TODO())
		volumes, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getVolumes", "statusCode", resp.StatusCode, "resp", resp)
		} else {
			break
		}
	}
	return volumes, err
}

func (oac oaClient) deleteVolume(uuid string) error {
	var resp *http.Response
	var err error

	oac.ensure()
	for _, client := range oac.clients {
		req := client.VolumesApi.DelVolume(context.TODO(), uuid)
		resp, err = req.Execute()
		if err != nil {
			log.Log.Info("deleteVolume", "statusCode", resp.StatusCode, "resp", resp)
		} else {
			break
		}
	}
	return err
}

func (oac oaClient) putReplicaCount(uuid string, replicaCount int) error {
	var resp *http.Response
	var err error

	oac.ensure()
	for _, client := range oac.clients {
		req := client.VolumesApi.PutVolumeReplicaCount(context.TODO(), uuid, int32(replicaCount))
		_, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("putReplicaCount", "statusCode", resp.StatusCode, "resp", resp)
		} else {
			break
		}
	}
	return err
}

func (oac oaClient) volToMsv(vol openapiClient.Volume) common.MayastorVolume {
	var nexusChildren []common.NexusChild
	var replicas []common.Replica

	volSpec := vol.GetSpec()
	volState := vol.GetState()
	nexus, ok := volState.GetTargetOk()
	if !ok {
		nexus = &openapiClient.Nexus{}
	}

	for _, inChild := range nexus.Children {
		nexusChildren = append(nexusChildren, common.NexusChild{
			Uri:   inChild.Uri,
			State: string(inChild.State),
		})
	}

	for k, replicaTopology := range volState.ReplicaTopology {
		uri := ""
		r, err := oac.getReplica(k)
		if err == nil {
			uri = r.GetUri()
		}
		replicas = append(replicas, common.Replica{
			Node:    replicaTopology.GetNode(),
			Offline: replicaTopology.GetState() != "Online",
			Pool:    replicaTopology.GetPool(),
			Uri:     uri,
		})
	}

	return common.MayastorVolume{
		Name: volSpec.GetUuid(),
		Spec: common.MayastorVolumeSpec{
			Protocol:      string(vol.GetSpec().Target.GetProtocol()),
			ReplicaCount:  int(volSpec.GetNumReplicas()),
			RequiredBytes: int(volSpec.GetSize()), // FIXME required bytes should be int64
		},
		Status: common.MayastorVolumeStatus{
			Nexus: common.Nexus{
				Children:  nexusChildren,
				DeviceUri: nexus.DeviceUri,
				Node:      nexus.Node,
				State:     string(nexus.State),
			},
			Reason:   "",
			Replicas: replicas,
			Size:     volState.Size,
			State:    string(volState.Status),
		},
	}
}

/*
func (oac oaClient) getReplicas() (map[string]openapiClient.Replica, error) {
	replicaMap := make(map[string]openapiClient.Replica)
	var err error
	var resp *http.Response

	oac.ensure()
	for _, client := range oac.clients {
		req := client.ReplicasApi.GetReplicas(context.TODO())
		var replicas []openapiClient.Replica
		replicas, resp, err = req.Execute()
		if err == nil {
			for _, r := range replicas {
				replicaMap[r.Uri] = r
			}
			break
		} else {
			logf.Log.Info("getReplicas", "resp"statusCode", resp.StatusCode, ", resp)
		}
	}
	return replicaMap, err
}
*/
