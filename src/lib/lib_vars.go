package lib

var nsMayastor = "mayastor"       //e2e_config.GetConfig().Platform.MayastorNamespace
var fioImage = "mayadata/e2e-fio" //e2e_config.GetConfig().E2eFioImage
var fsxImage = "mayadata/e2e-fsx" //e2e_config.GetConfig().E2eFsxImage

// NSMayastor return the name of the namespace in which Mayastor is installed
func NSMayastor() string {
	return nsMayastor
}

// default fio arguments for E2E fio runs
var fioArgs = []string{
	"--name=benchtest",
	"--numjobs=1",
}

var fioParams = []string{
	"--direct=1",
	"--rw=randrw",
	"--ioengine=libaio",
	"--bs=4k",
	"--iodepth=16",
	"--verify=crc32",
	"--verify_fatal=1",
	"--verify_async=2",
}

// GetFioArgs return the default command line for fio - for use with Mayastor,
// for single volume
func GetFioArgs() []string {
	return append(fioArgs, fioParams...)
}

// GetDefaultFioArguments return the default settings (arguments) for fio - for use with Mayastor
func GetDefaultFioArguments() []string {
	return fioParams
}

func GetFioImage() string {
	return fioImage
}

func GetFsxImage() string {
	return fsxImage
}

var DefaultReplicaCount = 2 // e2e_config.GetConfig().DefaultReplicaCount
