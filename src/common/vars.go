package common

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

var NSMayastor = "mayastor"

func GetFioArgs() []string {
	return fioArgs
}
