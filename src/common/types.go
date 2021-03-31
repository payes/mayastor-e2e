package common

type ShareProto string

const (
	ShareProtoNvmf  ShareProto = "nvmf"
	ShareProtoIscsi ShareProto = "iscsi"
)

type VolumeType int

const (
	VolFileSystem VolumeType = iota
	VolRawBlock   VolumeType = iota
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
