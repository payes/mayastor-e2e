package k8stest

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
)

func GetMsPool(poolName string) (*common.MayastorPool, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.GetMsPool(poolName)
}

func ListMsPools() ([]common.MayastorPool, error) {
	EnsureNodeAddressesAreSet()
	return controlplane.ListMsPools()
}
