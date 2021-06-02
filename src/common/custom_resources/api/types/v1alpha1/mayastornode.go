package v1alpha1

import metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type MayastorNodeSpec struct {
	GrpcEndpoint string `json:"grpcEndpoint"`
}

type MayastorNode struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MayastorNodeSpec `json:"spec"`
	Status string           `json:"status"`
}

type MayastorNodeList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`

	Items []MayastorNode `json:"items"`
}
