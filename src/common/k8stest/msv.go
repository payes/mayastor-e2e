package k8stest

import (
	. "github.com/onsi/gomega"
	"sync"
)

var once sync.Once

type MayastorVolumeSpec struct {
	Protocol      string `json:"protocol"`
	ReplicaCount  int    `json:"replicaCount"`
	RequiredBytes int    `json:"requiredBytes"`
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
	Nexus    Nexus     `json:"nexus"`
	Reason   string    `json:"reason"`
	Replicas []Replica `json:"replicas"`
	Size     int64     `json:"size"`
	State    string    `json:"state"`
}

type MayastorVolume struct {
	Name   string               `json:"name"`
	Spec   MayastorVolumeSpec   `json:"spec"`
	Status MayastorVolumeStatus `json:"status"`
}

type MayastorVolumeInterface interface {
	getMSV(uuid string) (*MayastorVolume, error)
	getMsvNodes(uuid string) (string, []string)
	deleteMsv(volName string) error
	listMsvs() ([]MayastorVolume, error)
	setMsvReplicaCount(uuid string, replicaCount int) error
	getMsvState(uuid string) (string, error)
	getMsvReplicas(volName string) ([]Replica, error)
	getMsvNexusChildren(volName string) ([]NexusChild, error)
	getMsvNexusState(uuid string) (string, error)
	isMsvPublished(uuid string) bool
	isMsvDeleted(uuid string) bool
	checkForMsvs() (bool, error)
	checkAllMsvsAreHealthy() error
}

// DO NOT ACCESS DIRECTLY, use GetMsvIfc
var ifc MayastorVolumeInterface

func GetMsvIfc() MayastorVolumeInterface {
	once.Do(func() {
		if IsControlPlaneMoac() {
			ifc = MakeCrMsv()
		} else if IsControlPlaneMcp() {
			Expect(false).To(BeTrue())
		} else {
			Expect(false).To(BeTrue())
		}
	})

	return ifc
}
