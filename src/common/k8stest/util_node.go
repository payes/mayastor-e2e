package k8stest

// Utility functions for manipulation of nodes.
import (
	"context"
	"errors"
	"fmt"
	"mayastor-e2e/common"
	"os/exec"

	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type NodeLocation struct {
	NodeName     string
	IPAddress    string
	MayastorNode bool
	MasterNode   bool
}

func GetNodeLocsMap() (map[string]NodeLocation, error) {
	NodeLocsMap := make(map[string]NodeLocation)
	nodeLocs, err := GetNodeLocs()
	if err != nil {
		return nil, err
	}
	for _, node := range nodeLocs {
		NodeLocsMap[node.NodeName] = node
	}
	return NodeLocsMap, nil
}

// GetNodeLocs returns vector of populated NodeLocation structs
func GetNodeLocs() ([]NodeLocation, error) {
	nodeList := coreV1.NodeList{}

	if gTestEnv.K8sClient.List(context.TODO(), &nodeList, &client.ListOptions{}) != nil {
		return nil, errors.New("failed to list nodes")
	}
	NodeLocs := make([]NodeLocation, 0, len(nodeList.Items))
	for _, k8snode := range nodeList.Items {
		addrstr := ""
		namestr := ""
		mayastorNode := false
		masterNode := false
		for label, value := range k8snode.Labels {
			if label == "openebs.io/engine" && value == "mayastor" {
				mayastorNode = true
			}
			if label == "node-role.kubernetes.io/master" {
				masterNode = true
			}
		}
		for _, addr := range k8snode.Status.Addresses {
			if addr.Type == coreV1.NodeInternalIP {
				addrstr = addr.Address
			}
			if addr.Type == coreV1.NodeHostName {
				namestr = addr.Address
			}
		}
		if namestr != "" && addrstr != "" {
			NodeLocs = append(NodeLocs, NodeLocation{
				NodeName:     namestr,
				IPAddress:    addrstr,
				MayastorNode: mayastorNode,
				MasterNode:   masterNode,
			})
		} else {
			return nil, errors.New("node lacks expected fields")
		}
	}
	return NodeLocs, nil
}

// GetNodeIPAddress returns IP address of a node
func GetNodeIPAddress(nodeName string) (*string, error) {
	nodeLocs, err := GetNodeLocs()
	if err != nil {
		return nil, err
	}
	for _, nl := range nodeLocs {
		if nodeName == nl.NodeName {
			return &nl.IPAddress, nil
		}
	}
	return nil, fmt.Errorf("node %s not found", nodeName)
}

// GetMayastorNodeIPAddresses return an array of IP addresses for nodes
// running mayastor. On error an empty array is returned.
func GetMayastorNodeIPAddresses() []string {
	var addrs []string
	nodes, err := GetNodeLocs()
	if err != nil {
		return addrs
	}

	for _, node := range nodes {
		if node.MayastorNode {
			addrs = append(addrs, node.IPAddress)
		}
	}
	return addrs
}

func GetMayastorNodeNames() ([]string, error) {
	var nodeNames []string
	nodes, err := GetNodeLocs()
	if err != nil {
		return nodeNames, err
	}

	for _, node := range nodes {
		if node.MayastorNode {
			nodeNames = append(nodeNames, node.NodeName)
		}
	}
	return nodeNames, err
}

// LabelNode add a label to a node
// label is a string in the form "key=value"
// function still succeeds if label already present
func LabelNode(nodename string, label string, value string) error {
	// TODO remove dependency on kubectl
	labelAssign := fmt.Sprintf("%s=%s", label, value)
	cmd := exec.Command("kubectl", "label", "node", nodename, labelAssign, "--overwrite=true")
	cmd.Dir = ""
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to label node %s, label: %s, error: %v", nodename, labelAssign, err)
	}
	return nil
}

// UnlabelNode remove a label from a node
// function still succeeds if label not present
func UnlabelNode(nodename string, label string) error {
	// TODO remove dependency on kubectl
	cmd := exec.Command("kubectl", "label", "node", nodename, label+"-")
	cmd.Dir = ""
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove label from node %s, label: %s, error: %v", nodename, label, err)
	}
	return nil
}

// AllowMasterScheduling removed NoSchedule toleration from master node
func AllowMasterScheduling() error {
	var masterNode string
	nodes, err := GetNodeLocs()
	if err != nil {
		return fmt.Errorf("failed to get nodes, error: %v", err)
	}

	for _, node := range nodes {
		if node.MasterNode {
			masterNode = node.NodeName
			break
		}
	}
	// TODO remove dependency on kubectl
	cmd := exec.Command("kubectl", "taint", "node", masterNode, "node-role.kubernetes.io/master:NoSchedule"+"-")
	cmd.Dir = ""
	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove  no schedule toleration from master node %s, error: %v", masterNode, err)
	}
	return nil
}

// RemoveMasterScheduling adds NoSchedule taint to master node
func RemoveMasterScheduling() error {
	var masterNode string
	nodes, err := GetNodeLocs()
	if err != nil {
		return fmt.Errorf("failed to get nodes, error: %v", err)
	}

	for _, node := range nodes {
		if node.MasterNode {
			masterNode = node.NodeName
			break
		}
	}
	// TODO remove dependency on kubectl
	cmd := exec.Command("kubectl", "taint", "node", masterNode, "node-role.kubernetes.io/master:NoSchedule")
	cmd.Dir = ""
	_, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add  no schedule toleration to master node %s, error: %v", masterNode, err)
	}
	return nil
}

// EnsureNodeLabels  add the label openebs.io/engine=mayastor to  all worker nodes so that K8s runs mayastor on them
// returns error is accessing the list of nodes fails.
func EnsureNodeLabels() error {
	var errors common.ErrorAccumulator
	nodes, err := GetNodeLocs()
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if !node.MasterNode {
			err = LabelNode(node.NodeName, common.MayastorEngineLabel, common.MayastorEngineLabelValue)
			if err != nil {
				errors.Accumulate(err)
			}
		}
	}
	return errors.GetError()
}

func AreNodesReady() (bool, error) {
	nodes, err := gTestEnv.KubeInt.CoreV1().Nodes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, node := range nodes.Items {
		readyStatus, err := IsNodeReady(node.Name, &node)
		if err != nil {
			return false, err
		}
		if !readyStatus {
			return false, nil
		}
	}
	return true, nil
}

func IsNodeReady(nodeName string, node *v1.Node) (bool, error) {
	var err error
	if node == nil {
		node, err = gTestEnv.KubeInt.CoreV1().Nodes().Get(context.TODO(), nodeName, metaV1.GetOptions{})
		if err != nil {
			return false, err
		}
	}
	master := false
	taints := node.Spec.Taints
	for _, taint := range taints {
		if taint.Key == "node-role.kubernetes.io/master" {
			master = true
		}
	}
	for _, nodeCond := range node.Status.Conditions {
		if nodeCond.Reason == "KubeletReady" && nodeCond.Type == v1.NodeReady {
			return true, nil
		} else if master && nodeCond.Type == v1.NodeReady {
			return true, nil
		}
	}
	addrs := node.Status.Addresses
	nodeAddr := ""
	for _, addr := range addrs {
		if addr.Type == v1.NodeInternalIP {
			nodeAddr = addr.Address
		}
	}
	logf.Log.Info("Node not ready", nodeName, nodeAddr)
	return false, nil
}
