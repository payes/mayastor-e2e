package v1alpha1

import metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type MayastorVolumeSpec struct {
	ReplicaCount   int      `json:"replicaCount"`
	Local          bool     `json:"local"`
	PreferredNodes []string `json:"preferredNodes"`
	RequiredNodes  []string `json:"requiredNodes"`
	RequiredBytes  int      `json:"requiredBytes"`
	LimitBytes     int      `json:"limitBytes"`
	Protocol       string   `json:"protocol"`
}

type NexusChild struct {
	Uri   string `json:"uri"`
	State string `json:"state"`
}

type Nexus struct {
	Node      string       `json:"node"`
	DeviceUri string       `json:"deviceUri"`
	State     string       `json:"state"`
	Children  []NexusChild `json:"children"`
}

type Replica struct {
	Node    string `json:"node"`
	Pool    string `json:"pool"`
	Uri     string `json:"uri"`
	Offline bool   `json:"offline"`
}

type MayastorVolumeStatus struct {
	Size        int64     `json:"size"`
	State       string    `json:"state"`
	Reason      string    `json:"reason"`
	TargetNodes []string  `json:"targetNodes"`
	Nexus       Nexus     `json:"Nexus"`
	Replicas    []Replica `json:"replicas"`
}

type MayastorVolume struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MayastorVolumeSpec   `json:"spec"`
	Status MayastorVolumeStatus `json:"status"`
}

type MayastorVolumeList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`

	Items []MayastorVolume `json:"items"`
}
