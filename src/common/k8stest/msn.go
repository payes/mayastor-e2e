package k8stest

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
)

// GetMSN Get pointer to a mayastor volume custom resource
// returns nil and no error if the msn is in pending state.
func GetMSN(nodeName string) (*common.MayastorNode, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.GetMSN(nodeName)
}

func ListMsns() ([]common.MayastorNode, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.ListMsns()
}

func GetMsNodeStatus(nodeName string) (string, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.GetMsNodeStatus(nodeName)
}
