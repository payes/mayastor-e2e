package mayastorclient

import (
	"mayastor-e2e/common/k8stest"
	mayastorGrpc "mayastor-e2e/common/mayastorclient/grpc"
)

var null = mayastorGrpc.Null{}

const mayastorPort = 10124

func getClusterMayastorNodeIPAddrs() ([]string, error) {
	var nodeAddrs []string
	nodes, err := k8stest.GetNodeLocs()
	if err != nil {
		return nodeAddrs, err
	}

	for _, node := range nodes {
		if node.MayastorNode {
			nodeAddrs = append(nodeAddrs, node.IPAddress)
		}
	}
	return nodeAddrs, err
}
