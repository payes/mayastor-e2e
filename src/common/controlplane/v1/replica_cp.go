package v1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mayastor-e2e/common"
	"net/http"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type mayastorCpReplica struct {
	Node  string `json:"node"`
	Pool  string `json:"pool"`
	Share string `json:"share"`
	Size  int64  `json:"size"`
	State string `json:"state"`
	Thin  bool   `json:"thin"`
	Uri   string `json:"uri"`
	Uuid  string `json:"uuid"`
}

func listMayastorCpReplicas(address []string) ([]mayastorCpReplica, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("mayastor nodes not found")
	}
	var jsonResponse []byte
	var err error
	for _, addr := range address {
		url := fmt.Sprintf("http://%s:%s/v0/replicas", addr, common.PluginPort)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logf.Log.Info("Error in GET request", "node IP", addr, "url", url, "error", err)
		}
		req.Header.Add("Accept", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logf.Log.Info("Error while making GET request", "url", url, "error", err)
		}
		defer resp.Body.Close()
		jsonResponse, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logf.Log.Info("Error while reading data", "error", err)
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	var response []mayastorCpReplica
	err = json.Unmarshal(jsonResponse, &response)
	if err != nil {
		logf.Log.Info("Failed to unmarshal (get replicas)", "string", string(jsonResponse))
		return nil, err
	}
	return response, nil
}
func getMayastorCpReplica(replicaUuid string, address []string) (mayastorCpReplica, error) {
	if len(address) == 0 {
		return mayastorCpReplica{}, fmt.Errorf("mayastor nodes not found")
	}
	var jsonResponse []byte
	var err error
	for _, addr := range address {
		url := fmt.Sprintf("http://%s:%s/v0/replicas/%s", addr, common.PluginPort, replicaUuid)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logf.Log.Info("Error in GET request", "node IP", addr, "url", url, "error", err)
		}
		req.Header.Add("Accept", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logf.Log.Info("Error while making GET request", "url", url, "error", err)
		}
		defer resp.Body.Close()
		jsonResponse, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logf.Log.Info("Error while reading data", "error", err)
		} else {
			break
		}
	}
	if err != nil {
		return mayastorCpReplica{}, err
	}
	var response mayastorCpReplica
	err = json.Unmarshal(jsonResponse, &response)
	if err != nil {
		logf.Log.Info("Failed to unmarshal (get replicas)", "string", string(jsonResponse))
		return mayastorCpReplica{}, err
	}
	return response, nil
}
