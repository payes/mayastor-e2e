package rest_api

import (
	"mayastor-e2e/common"
	"mayastor-e2e/generated/openapi"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func oaNodeToMsn(oaNode *openapi.Node) common.MayastorNode {
	nodeSpec := oaNode.GetSpec()
	nodeState := oaNode.GetState()
	msn := common.MayastorNode{
		Name: oaNode.GetId(),
		Spec: common.MayastorNodeSpec{
			GrpcEndpoint: nodeSpec.GrpcEndpoint,
			ID:           nodeSpec.Id,
		},
		State: common.MayastorNodeState{
			GrpcEndpoint: nodeState.GrpcEndpoint,
			ID:           nodeState.Id,
			Status:       string(nodeState.GetStatus()),
		},
	}

	return msn
}

func (cp CPv1RestApi) GetMSN(nodeName string) (*common.MayastorNode, error) {
	oaNode, err, statusCode := cp.oa.getNode(nodeName)

	if err != nil {
		logf.Log.Info("getNode failed", "statusCode", statusCode)
		return nil, err
	}

	msn := oaNodeToMsn(&oaNode)
	return &msn, err
}

func (cp CPv1RestApi) ListMsns() ([]common.MayastorNode, error) {
	oaNodes, err, statusCode := cp.oa.getNodes()
	if err != nil {
		logf.Log.Info("getNodes failed", "statusCode", statusCode)
		return nil, err
	}
	var msns []common.MayastorNode
	for _, oaNode := range oaNodes {
		msns = append(msns, oaNodeToMsn(&oaNode))
	}
	return msns, err
}

func (cp CPv1RestApi) GetMsNodeStatus(nodeName string) (string, error) {
	rNode, err, _ := cp.oa.getNode(nodeName)

	if err == nil {
		return string(rNode.State.GetStatus()), err
	}
	return "", err
}
