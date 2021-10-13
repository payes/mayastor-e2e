package rest_api

import (
	"context"
	"fmt"
	"net/http"

	"mayastor-e2e/common"
	openapiClient "mayastor-e2e/generated/openapi"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type restApiClient struct {
	nodes   []string
	cfg     []*openapiClient.Configuration
	clients []*openapiClient.APIClient
}

func makeClient(nodeIPs []string) restApiClient {
	oac := restApiClient{
		nodes: nodeIPs,
	}
	for _, node := range oac.nodes {
		cfg := openapiClient.NewConfiguration()
		cfg.Host = node + ":30011"
		cfg.Scheme = `http`
		client := openapiClient.NewAPIClient(cfg)
		oac.clients = append(oac.clients, client)
		oac.cfg = append(oac.cfg, cfg)
	}
	return oac
}

func (oac restApiClient) getReplica(uuid string) (openapiClient.Replica, error, int) {
	var replica openapiClient.Replica
	var resp *http.Response
	var err error

	for _, client := range oac.clients {
		req := client.ReplicasApi.GetReplica(context.TODO(), uuid)
		replica, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getReplica", "statusCode", resp.StatusCode, "error", err)
			err = fmt.Errorf("%v, statusCode=%d", err, resp.StatusCode)
		} else {
			break
		}
	}
	return replica, err, resp.StatusCode
}

func (oac restApiClient) getVolume(uuid string) (openapiClient.Volume, error, int) {
	var volume openapiClient.Volume
	var resp *http.Response
	var err error

	for _, client := range oac.clients {
		req := client.VolumesApi.GetVolume(context.TODO(), uuid)
		volume, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getVolume", "statusCode", resp.StatusCode, "error", err)
			err = fmt.Errorf("%v, statusCode=%d", err, resp.StatusCode)
		} else {
			break
		}
	}
	return volume, err, resp.StatusCode
}

func (oac restApiClient) getVolumes() ([]openapiClient.Volume, error, int) {
	var volumes []openapiClient.Volume
	var resp *http.Response
	var err error

	for _, client := range oac.clients {
		req := client.VolumesApi.GetVolumes(context.TODO())
		volumes, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getVolumes", "statusCode", resp.StatusCode, "error", err)
			err = fmt.Errorf("%v, statusCode=%d", err, resp.StatusCode)
		} else {
			break
		}
	}
	return volumes, err, resp.StatusCode
}

func (oac restApiClient) deleteVolume(uuid string) (error, int) {
	var resp *http.Response
	var err error

	for _, client := range oac.clients {
		req := client.VolumesApi.DelVolume(context.TODO(), uuid)
		resp, err = req.Execute()
		if err != nil {
			log.Log.Info("deleteVolume", "statusCode", resp.StatusCode, "error", err)
			err = fmt.Errorf("%v, statusCode=%d", err, resp.StatusCode)
		} else {
			break
		}
	}
	return err, resp.StatusCode
}

func (oac restApiClient) putReplicaCount(uuid string, replicaCount int) (error, int) {
	var resp *http.Response
	var err error

	for _, client := range oac.clients {
		req := client.VolumesApi.PutVolumeReplicaCount(context.TODO(), uuid, int32(replicaCount))
		_, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("putReplicaCount", "statusCode", resp.StatusCode, "error", err)
			err = fmt.Errorf("%v, statusCode=%d", err, resp.StatusCode)
		} else {
			break
		}
	}
	return err, resp.StatusCode
}

func (oac restApiClient) volToMsv(vol openapiClient.Volume) common.MayastorVolume {
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
		r, err, _ := oac.getReplica(k)
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

func (oac restApiClient) getNode(nodeName string) (openapiClient.Node, error, int) {
	var node openapiClient.Node
	var resp *http.Response
	var err error

	for _, client := range oac.clients {
		req := client.NodesApi.GetNode(context.TODO(), nodeName)
		node, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getNode", "statusCode", resp.StatusCode, "error", err)
			err = fmt.Errorf("%v, statusCode=%d", err, resp.StatusCode)
		} else {
			break
		}
	}
	return node, err, resp.StatusCode
}

func (oac restApiClient) getNodes() ([]openapiClient.Node, error, int) {
	var nodes []openapiClient.Node
	var resp *http.Response
	var err error

	for _, client := range oac.clients {
		req := client.NodesApi.GetNodes(context.TODO())
		nodes, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getNodes", "statusCode", resp.StatusCode, "error", err)
			err = fmt.Errorf("%v, statusCode=%d", err, resp.StatusCode)
		} else {
			break
		}
	}
	return nodes, err, resp.StatusCode
}

func (oac restApiClient) getPool(poolName string) (openapiClient.Pool, error, int) {
	var pool openapiClient.Pool
	var resp *http.Response
	var err error

	for _, client := range oac.clients {
		req := client.PoolsApi.GetPool(context.TODO(), poolName)
		pool, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getPool", "statusCode", resp.StatusCode, "error", err)
			err = fmt.Errorf("%v, statusCode=%d", err, resp.StatusCode)
		} else {
			break
		}
	}
	return pool, err, resp.StatusCode
}

func (oac restApiClient) getPools() ([]openapiClient.Pool, error, int) {
	var pools []openapiClient.Pool
	var resp *http.Response
	var err error

	for _, client := range oac.clients {
		req := client.PoolsApi.GetPools(context.TODO())
		pools, resp, err = req.Execute()
		if err != nil {
			log.Log.Info("getPools", "statusCode", resp.StatusCode, "error", err)
			err = fmt.Errorf("%v, statusCode=%d", err, resp.StatusCode)
		} else {
			break
		}
	}
	return pools, err, resp.StatusCode
}
