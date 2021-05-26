package k8stest

// Utility functions for manipulation of nodes.
import (
	"context"
	"errors"
	"fmt"
	"mayastor-e2e/common"
	"os/exec"

	. "github.com/onsi/gomega"

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

// returns vector of populated NodeLocation structs
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

// TODO remove dependency on kubectl
// label is a string in the form "key=value"
// function still succeeds if label already present
func LabelNode(nodename string, label string, value string) {
	labelAssign := fmt.Sprintf("%s=%s", label, value)
	cmd := exec.Command("kubectl", "label", "node", nodename, labelAssign, "--overwrite=true")
	cmd.Dir = ""
	_, err := cmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred())
}

// TODO remove dependency on kubectl
// function still succeeds if label not present
func UnlabelNode(nodename string, label string) {
	cmd := exec.Command("kubectl", "label", "node", nodename, label+"-")
	cmd.Dir = ""
	_, err := cmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred())
}

// EnsureNodeLabels  add the label openebs.io/engine=mayastor to  all worker nodes so that K8s runs mayastor on them
// returns error is accessing the list of nodes fails.
func EnsureNodeLabels() error {
	nodes, err := GetNodeLocs()
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if !node.MasterNode {
			LabelNode(node.NodeName, common.MayastorEngineLabel, common.MayastorEngineLabelValue)
		}
	}
	return nil
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
