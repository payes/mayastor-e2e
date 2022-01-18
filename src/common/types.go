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
