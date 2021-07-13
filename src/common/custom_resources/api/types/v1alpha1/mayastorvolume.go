package v1alpha1

import metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type MayastorVolumeSpec struct {
	LimitBytes     int      `json:"limitBytes"`
	Local          bool     `json:"local"`
	PreferredNodes []string `json:"preferredNodes"`
	Protocol       string   `json:"protocol"`
	ReplicaCount   int      `json:"replicaCount"`
	RequiredBytes  int      `json:"requiredBytes"`
	RequiredNodes  []string `json:"requiredNodes"`
}

type NexusChild struct {
	State string `json:"state"`
	Uri   string `json:"uri"`
}

type Nexus struct {
	Children  []NexusChild `json:"children"`
	DeviceUri string       `json:"deviceUri"`
	Node      string       `json:"node"`
	State     string       `json:"state"`
}

type Replica struct {
	Node    string `json:"node"`
	Offline bool   `json:"offline"`
	Pool    string `json:"pool"`
	Uri     string `json:"uri"`
}

type MayastorVolumeStatus struct {
	Nexus       Nexus     `json:"nexus"`
	Reason      string    `json:"reason"`
	Replicas    []Replica `json:"replicas"`
	Size        int64     `json:"size"`
	State       string    `json:"state"`
	TargetNodes []string  `json:"targetNodes"`
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
