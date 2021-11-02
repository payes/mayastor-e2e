package k8sclient

import (
	"context"
	"fmt"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	coreV1 "k8s.io/api/core/v1"
)

type NodeLocation struct {
	NodeName     string
	IPAddress    string
	MayastorNode bool
	MasterNode   bool
}

func GetNodeLocs() ([]NodeLocation, error) {

	nodeList, err := gKubeInt.CoreV1().Nodes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes, error: %v", err)
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
			return nil, fmt.Errorf("node lacks expected fields")
		}
	}
	return NodeLocs, nil
}

func GetMayastorNodeNames() ([]string, error) {
	var nodeNames []string
	nodes, err := GetNodeLocs()
	if err != nil {
		return nodeNames, fmt.Errorf("failed to get node locations, error: %v", err)
	}

	for _, node := range nodes {
		if node.MayastorNode {
			nodeNames = append(nodeNames, node.NodeName)
		}
	}
	return nodeNames, err
}

func GetMayastorNodeIPs() ([]string, error) {
	var nodeIPs []string
	nodes, err := GetNodeLocs()
	if err != nil {
		return nodeIPs, fmt.Errorf("failed to get node locations, error: %v", err)
	}

	for _, node := range nodes {
		if node.MayastorNode {
			nodeIPs = append(nodeIPs, node.IPAddress)
		}
	}
	return nodeIPs, err
}
