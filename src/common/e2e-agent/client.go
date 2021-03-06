package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// RestPort is the port on which e2e-agent is listening
const RestPort = "10012"

// NodeList is the list of nodes to be passed to e2e-agent
type NodeList struct {
	Nodes []string `json:"nodes"`
}

type CmdList struct {
	Cmd string `json:"cmd"`
}

type Device struct {
	Device string `json:"device"`
	Table  string `json:"table"`
}

type ControlledDevice struct {
	Device string `json:"device"`
	State  string `json:"state"`
}

func sendRequest(reqType, url string, data interface{}) error {
	_, err := sendRequestGetResponse(reqType, url, data, true)
	return err
}

func sendRequestGetResponse(reqType, url string, data interface{}, verbose bool) (string, error) {
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
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("request returned code %d", resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if verbose {
		fmt.Printf("resp: %s\n", bodyBytes)
	}
	return string(bodyBytes), nil
}

// UngracefulReboot crashes and reboots the host machine
func UngracefulReboot(serverAddr string) error {
	logf.Log.Info("Ungracefully rebooting node", "addr", serverAddr)
	url := "http://" + serverAddr + ":" + RestPort + "/ungracefulReboot"
	return sendRequest("POST", url, nil)
}

// IsAgentReachable checks if the agent pod is in reachable
func IsAgentReachable(serverAddr string) error {
	url := "http://" + serverAddr + ":" + RestPort + "/"
	return sendRequest("GET", url, nil)
}

// GracefulReboot reboots the host gracefully
// It is not yet supported
func GracefulReboot(serverAddr string) error {
	logf.Log.Info("Gracefully rebooting node", "addr", serverAddr)
	url := "http://" + serverAddr + ":" + RestPort + "/gracefulReboot"
	return sendRequest("POST", url, nil)
}

// DropConnectionsFromNodes creates rules to drop connections from other k8s nodes
func DropConnectionsFromNodes(serverAddr string, nodes []string) error {
	logf.Log.Info("Dropping connections from nodes", "addr", serverAddr)
	url := "http://" + serverAddr + ":" + RestPort + "/dropConnectionsFromNodes"
	data := NodeList{
		Nodes: nodes,
	}
	return sendRequest("POST", url, data)
}

// AcceptConnectionsFromNodes removes the rules set by
// DropConnectionsFromNodes so that other k8s nodes can reach this node again
func AcceptConnectionsFromNodes(serverAddr string, nodes []string) error {
	logf.Log.Info("Accepting connections from nodes", "addr", serverAddr)
	url := "http://" + serverAddr + ":" + RestPort + "/acceptConnectionsFromNodes"
	data := NodeList{
		Nodes: nodes,
	}
	return sendRequest("POST", url, data)
}

// DiskPartition performs operation related to disk prtitioning
func DiskPartition(serverAddr string, cmd string) error {
	url := "http://" + serverAddr + ":" + RestPort + "/exec"
	data := CmdList{
		Cmd: cmd,
	}
	return sendRequest("POST", url, data)
}

// CreateFaultyDevice creates a device which returns an error on write IOs
func CreateFaultyDevice(serverAddr, device, table string) error {
	url := "http://" + serverAddr + ":" + RestPort + "/createFaultyDevice"
	data := Device{
		Device: device,
		Table:  table,
	}
	return sendRequest("POST", url, data)
}

// Exec sends the shell command to the e2e-agent
func Exec(serverAddr string, command string) (string, error) {
	logf.Log.Info("Executing command on node", "command", command, "addr", serverAddr)
	url := "http://" + serverAddr + ":" + RestPort + "/exec"
	data := CmdList{
		Cmd: command,
	}
	return sendRequestGetResponse("POST", url, data, false)
}

// ControlDevice sets the specified to the specified state
// by writing to /sys/block/<device e.g. sdb>/device/state
// The only accepted states are "running" and "offline"
func ControlDevice(serverAddr string, device string, state string) (string, error) {
	logf.Log.Info("Controlling device", "device", device, "state", state, "addr", serverAddr)
	url := "http://" + serverAddr + ":" + RestPort + "/devicecontrol"
	data := ControlledDevice{
		Device: device,
		State:  state,
	}
	return sendRequestGetResponse("POST", url, data, false)
}

// KillMayastor use kill -9 against the mayastor
func KillMayastor(serverAddr string) (string, error) {
	logf.Log.Info("Killing Mayastor", "addr", serverAddr)
	url := "http://" + serverAddr + ":" + RestPort + "/killmayastor"
	return sendRequestGetResponse("POST", url, nil, true)
}
