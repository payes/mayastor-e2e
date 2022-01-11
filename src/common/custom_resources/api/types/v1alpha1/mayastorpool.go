package v1alpha1

import metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type MayastorPoolSpec struct {
	Disks []string `json:"disks"`
	Node  string   `json:"node"`
}

type MayastorPoolStatus struct {
	Available uint64 `json:"available"`
	Capacity  uint64 `json:"capacity"`
	State     string `json:"state"`
	Used      uint64 `json:"used"`
}

type MayastorPool struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MayastorPoolSpec   `json:"spec"`
	Status MayastorPoolStatus `json:"status"`
}

type MayastorPoolList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`

	Items []MayastorPool `json:"items"`
}
