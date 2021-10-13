package v1

import (
	"context"
	"fmt"
	"k8s.io/api/core/v1"
	"mayastor-e2e/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetMasterNodeIPs return the IP addresses of master nodes,
// Both v1 control plane implementations require access to the master node
// - control plane using kubectl uses the rest api to retrieve a specific replica
//		the plugin does not return information for replicas
// - control plane using rest api uses setup openapi clients
func GetMasterNodeIPs() ([]string, error) {
	var addrs []string
	nodeList := v1.NodeList{}

	k8sclient, err := common.GetK8sClient()
	if err == nil {
		err = (*k8sclient).List(context.TODO(), &nodeList, &client.ListOptions{})
		if err == nil {
			if len(nodeList.Items) > 0 {
				for _, k8sNode := range nodeList.Items {
					for label := range k8sNode.Labels {
						if label == "node-role.kubernetes.io/master" {
							for _, addr := range k8sNode.Status.Addresses {
								if addr.Type == v1.NodeInternalIP {
									addrs = []string{addr.Address}
								}
							}
						}
					}
				}
			} else {
				err = fmt.Errorf("got empty list of k8s nodes")
			}
		}
	}
	return addrs, err
}
