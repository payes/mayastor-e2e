package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const REST_PORT = "10000"

type NodeList struct {
	Nodes []string `json:"nodes"`
}

func sendRequest(url string, data interface{}) error {
	client := &http.Client{}
	reqData := new(bytes.Buffer)
	if err := json.NewEncoder(reqData).Encode(data); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, reqData)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var responseObject string
	if err := json.Unmarshal(bodyBytes, &responseObject); err != nil {
		return err
	}
	return nil
}

// UngracefulReboot crashes and reboots the host machine
func UngracefulReboot(serverAddr string) error {
	url := "http://" + serverAddr + ":" + REST_PORT + "/ungracefulReboot"
	return sendRequest(url, nil)
}

// GracefulReboot reboots the host gracefully
// It is not yet supported
func GracefulReboot(serverAddr string) error {
	url := "http://" + serverAddr + ":" + REST_PORT + "/gracefulReboot"
	return sendRequest(url, nil)
}

// DropConnectionsFromNodes creates rules to drop connections from other k8s nodes
func DropConnectionsFromNodes(serverAddr string, nodes []string) error {
	url := "http://" + serverAddr + ":" + REST_PORT + "/dropConnectionsFromNodes"
	data := NodeList{
		Nodes: nodes,
	}
	return sendRequest(url, data)
}

// AcceptConnectionsFromNodes removes the rules set by
// DropConnectionsFromNodes so that other k8s nodes can reach this node again
func AcceptConnectionsFromNodes(serverAddr string, nodes []string) error {
	url := "http://" + serverAddr + ":" + REST_PORT + "/acceptConnectionsFromNodes"
	data := NodeList{
		Nodes: nodes,
	}
	return sendRequest(url, data)
}
