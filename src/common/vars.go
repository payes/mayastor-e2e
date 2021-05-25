package common

import "mayastor-e2e/common/e2e_config"

var nsMayastor = e2e_config.GetConfig().Platform.MayastorNamespace
var fioImage = e2e_config.GetConfig().E2eFioImage

// NSMayastor return the name of the namespace in which Mayastor is installed
func NSMayastor() string {
	return nsMayastor
}

// default fio arguments for E2E fio runs
var fioArgs = []string{
	"--name=benchtest",
	"--direct=1",
	"--rw=randrw",
	"--ioengine=libaio",
	"--bs=4k",
	"--iodepth=16",
	"--numjobs=1",
	"--verify=crc32",
	"--verify_fatal=1",
	"--verify_async=2",
}

// GetFioArgs return the default argument set for fio - for use with Mayastor
func GetFioArgs() []string {
	return fioArgs
}

func GetFioImage() string {
	return fioImage
}
