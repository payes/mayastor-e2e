package v1alpha1

import metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type MayastorPoolSpec struct {
	Node  string   `json:"node"`
	Disks []string `json:"disks"`
}

type MayastorPoolStatus struct {
	Capacity int64    `json:"capacity"`
	Disks    []string `json:"disks"`
	Reason   string   `json:"reason"`
	State    string   `json:"state"`
	Used     int64    `json:"used"`
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
