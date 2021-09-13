package lib

import (
	"context"
	"fmt"

	errors "github.com/pkg/errors"

	//"mayastor-e2e/tools/extended-test-framework/client"
	//"mayastor-e2e/tools/extended-test-framework/client/test_director"

	//"mayastor-e2e/tools/extended-test-framework/models"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/rest"

	//"time"

	coreV1 "k8s.io/api/core/v1"
)

type NodeLocation struct {
	NodeName     string
	IPAddress    string
	MayastorNode bool
	MasterNode   bool
}

func GetNodeLocs(k8sclient kubernetes.Clientset) ([]NodeLocation, error) {
	//nodeList := coreV1.NodeList{}

	nodeList, err := k8sclient.CoreV1().Nodes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		text := fmt.Sprintf("failed to list nodes: %v", err)
		return nil, errors.New(text)
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

func GetMayastorNodeNames(k8sclient kubernetes.Clientset) ([]string, error) {
	var nodeNames []string
	nodes, err := GetNodeLocs(k8sclient)
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
