package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MayastorPoolSpec   `json:"spec"`
	Status MayastorPoolStatus `json:"status"`
}

type MayastorPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MayastorPool `json:"items"`
}
