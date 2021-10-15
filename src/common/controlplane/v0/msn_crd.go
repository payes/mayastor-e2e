package v0

// Utility functions for Mayastor CRDs
import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	v1alpha1Api "mayastor-e2e/common/custom_resources/api/types/v1alpha1"
)

func crdToMsn(crdMsn *v1alpha1Api.MayastorNode) common.MayastorNode {
	return common.MayastorNode{
		Name: crdMsn.GetName(),
		Spec: common.MayastorNodeSpec{
			ID:           crdMsn.GetName(),
			GrpcEndpoint: crdMsn.Spec.GrpcEndpoint,
		},
		State: common.MayastorNodeState{
			GrpcEndpoint: crdMsn.Spec.GrpcEndpoint,
			Status:       crdMsn.Status,
		},
	}
}

// GetMSN Get pointer to a mayastor node custom resource
func (cp CPv0p8) GetMSN(node string) (*common.MayastorNode, error) {
	crdMsn, err := custom_resources.GetMsNode(node)
	if err != nil {
		return nil, fmt.Errorf("GetMSN: %v", err)
	}

	msn := crdToMsn(crdMsn)
	return &msn, nil
}

func (cp CPv0p8) ListMsns() ([]common.MayastorNode, error) {
	var msns []common.MayastorNode
	crs, err := custom_resources.ListMsNodes()
	if err == nil {
		for _, cr := range crs {
			msns = append(msns, crdToMsn(&cr))
		}
	}
	return msns, err
}

// GetMsNodeStatus Get pointer to a mayastor node custom resource
func (cp CPv0p8) GetMsNodeStatus(node string) (string, error) {
	crdMsn, err := custom_resources.GetMsNode(node)
	if err != nil {
		return "", fmt.Errorf("GetMSN: %v", err)
	}

	return crdMsn.Status, nil
}
