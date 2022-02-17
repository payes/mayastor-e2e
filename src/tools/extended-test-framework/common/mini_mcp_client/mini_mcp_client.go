package mini_mcp_client

// This is a light-weight client for accessing the control plane
// using a limited number of the APIs.
// In time this will be replaced with a full implementation using OpenAPI.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// RestPort is the port on which the control plane is listening
const RestPort = "30011"

func sendRequest(reqType, url string, data interface{}) (string, error) {
	client := &http.Client{}
	reqData := new(bytes.Buffer)
	if err := json.NewEncoder(reqData).Encode(data); err != nil {
		return "", err
	}
	req, err := http.NewRequest(reqType, url, reqData)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

// a cut-down version of the Mayastor Volume returned by the control plane
// with only fields we are interested in.
type SparseVolume struct {
	Spec struct {
		Num_replicas int `json:"num_replicas"`
	} `json:"spec"`
	State struct {
		Target struct {
			Children []struct {
				State string `json:"state"`
			} `json:"children"`
			Uuid string `json:"uuid"`
		} `json:"target"`
		Uuid   string `json:"uuid"`
		Status string `json:"status"`
	} `json:"state"`
}

// Get the volume uuid and status.
// Errors if there are not exactly 1 MS volume in the cluster.
func GetOnlyVolume(serverAddr string) (string, string, error) {
	var vollist []SparseVolume
	url := "http://" + serverAddr + ":" + RestPort + "/v0/volumes"
	resp, err := sendRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal([]byte(resp), &vollist)
	if err != nil {
		return "", "", err
	}
	if len(vollist) != 1 {
		return "", "", fmt.Errorf("invalid number of volumes %d", len(vollist))
	}
	return vollist[0].State.Uuid, vollist[0].State.Status, nil
}

func GetVolumes(serverAddr string) ([]SparseVolume, error) {
	var vollist []SparseVolume
	url := "http://" + serverAddr + ":" + RestPort + "/v0/volumes"
	resp, err := sendRequest("GET", url, nil)
	if err != nil {
		return vollist, err
	}
	err = json.Unmarshal([]byte(resp), &vollist)
	return vollist, err
}

// GetVolume gets the given volume
func GetVolume(serverAddr string, uuid string) (SparseVolume, error) {
	var vol SparseVolume
	url := "http://" + serverAddr + ":" + RestPort + "/v0/volumes/" + uuid
	resp, err := sendRequest("GET", url, nil)
	if err != nil {
		return vol, err
	}
	err = json.Unmarshal([]byte(resp), &vol)
	return vol, err
}

// GetVolumeStatus gets the volume status of the given volume
func GetVolumeStatus(serverAddr string, uuid string) (string, error) {
	vol, err := GetVolume(serverAddr, uuid)
	if err != nil {
		return "", err
	}
	return vol.State.Status, nil
}

// SetVolumeReplicas set the number of replicas for the volume
func SetVolumeReplicas(serverAddr string, volumeid string, replicas int) error {
	url := "http://" + serverAddr + ":" + RestPort + "/v0/volumes/" + volumeid + "/replica_count/" + strconv.Itoa(replicas)
	_, err := sendRequest("PUT", url, nil)
	return err
}

// GetVolumeReplicas get the number of replicas for the volume by counting the children
func GetVolumeReplicaCount(serverAddr string, uuid string) (int, error) {
	var vol SparseVolume
	url := "http://" + serverAddr + ":" + RestPort + "/v0/volumes/" + uuid
	resp, err := sendRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}
	err = json.Unmarshal([]byte(resp), &vol)
	if err != nil {
		return 0, err
	}
	return len(vol.State.Target.Children), nil
}
