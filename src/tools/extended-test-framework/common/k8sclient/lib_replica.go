package k8sclient

import "mayastor-e2e/common/mayastorclient"

// ListReplicasInCluster use mayastorclient to enumerate the set of mayastor replicas present in the cluster
func ListReplicasInCluster() ([]mayastorclient.MayastorReplica, error) {
	nodeAddrs, err := GetMayastorNodeIPs()
	if err == nil {
		return mayastorclient.ListReplicas(nodeAddrs)
	}
	return []mayastorclient.MayastorReplica{}, err
}

// RmReplicasInCluster use mayastorclient to remove mayastor replicas present in the cluster
func RmReplicasInCluster() error {
	nodeAddrs, err := GetMayastorNodeIPs()
	if err == nil {
		return mayastorclient.RmNodeReplicas(nodeAddrs)
	}
	return err
}
