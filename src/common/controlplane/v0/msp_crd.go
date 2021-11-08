package v0

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	v1alpha1Api "mayastor-e2e/common/custom_resources/api/types/v1alpha1"
)

func crdToMsp(crdMsp *v1alpha1Api.MayastorPool) common.MayastorPool {
	return common.MayastorPool{
		Name: crdMsp.Name,
		Spec: common.MayastorPoolSpec{
			Node:  crdMsp.Spec.Node,
			Disks: crdMsp.Spec.Disks,
		},
		Status: common.MayastorPoolStatus{
			Capacity: crdMsp.Status.Capacity,
			Used:     crdMsp.Status.Used,
			Disks:    crdMsp.Status.Disks,
			Spec: common.MayastorPoolSpec{
				Disks: crdMsp.Spec.Disks,
				Node:  crdMsp.Spec.Node,
			},
			State:  crdMsp.Status.State,
			Avail:  crdMsp.Status.Avail,
			Reason: crdMsp.Status.Reason,
		},
	}
}

// GetMsPool Get pointer to a mayastor node custom resource
func (cp CPv0p8) GetMsPool(poolName string) (*common.MayastorPool, error) {
	crdMsp, err := custom_resources.GetMsPool(poolName)
	if err != nil {
		return nil, fmt.Errorf("GetMsPool: %v", err)
	}

	msp := crdToMsp(&crdMsp)
	return &msp, nil
}

func (cp CPv0p8) ListMsPools() ([]common.MayastorPool, error) {
	var msps []common.MayastorPool
	crs, err := custom_resources.ListMsPools()
	if err == nil {
		for _, cr := range crs {
			msps = append(msps, crdToMsp(&cr))
		}
	}
	return msps, err
}
