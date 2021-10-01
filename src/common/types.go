package common

type ShareProto string

const (
	ShareProtoNvmf  ShareProto = "nvmf"
	ShareProtoIscsi ShareProto = "iscsi"
)

type FileSystemType string

const (
	NoneFsType FileSystemType = ""
	Ext4FsType FileSystemType = "ext4"
	XfsFsType  FileSystemType = "xfs"
)

type VolumeType int

const (
	VolFileSystem VolumeType = iota
	VolRawBlock   VolumeType = iota
	VolTypeNone   VolumeType = iota
)

func (volType VolumeType) String() string {
	switch volType {
	case VolFileSystem:
		return "FileSystem"
	case VolRawBlock:
		return "RawBlock"
	default:
		return "Unknown"
	}
}

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
	GetMSV(uuid string) (*MayastorVolume, error)
	GetMsvNodes(uuid string) (string, []string)
	DeleteMsv(volName string) error
	ListMsvs() ([]MayastorVolume, error)
	SetMsvReplicaCount(uuid string, replicaCount int) error
	GetMsvState(uuid string) (string, error)
	GetMsvReplicas(volName string) ([]Replica, error)
	GetMsvNexusChildren(volName string) ([]NexusChild, error)
	GetMsvNexusState(uuid string) (string, error)
	IsMsvPublished(uuid string) bool
	IsMsvDeleted(uuid string) bool
	CheckForMsvs() (bool, error)
	CheckAllMsvsAreHealthy() error
}
