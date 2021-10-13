package common

import (
	"fmt"
	"mayastor-e2e/common/e2e_config"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

var nsMayastor = e2e_config.GetConfig().Platform.MayastorNamespace
var fioImage = e2e_config.GetConfig().E2eFioImage
var fsxImage = e2e_config.GetConfig().E2eFsxImage
var registry = e2e_config.GetConfig().Registry
var k8sClient *client.Client

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
	return fmt.Sprintf("%s/%s", registry, fioImage)
}

func GetFsxImage() string {
	return fmt.Sprintf("%s/%s", registry, fsxImage)
}

func GetK8sClient() (*client.Client, error) {
	if k8sClient == nil {
		return nil, fmt.Errorf("incorrect code initialisation sequence, GetK8sClient called before test environment has been setup")
	}
	return k8sClient, nil
}

func SetK8sClient(client *client.Client) {
	k8sClient = client
}

var DefaultReplicaCount = e2e_config.GetConfig().DefaultReplicaCount
