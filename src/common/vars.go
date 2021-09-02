package common

import "mayastor-e2e/common/e2e_config"

var nsMayastor = e2e_config.GetConfig().Platform.MayastorNamespace
var fioImage = e2e_config.GetConfig().E2eFioImage
var fsxImage = e2e_config.GetConfig().E2eFsxImage

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

var DefaultReplicaCount = e2e_config.GetConfig().DefaultReplicaCount
