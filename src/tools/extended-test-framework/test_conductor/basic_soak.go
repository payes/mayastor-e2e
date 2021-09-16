package main

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/lib"
	"strings"
	"time"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

func deployFio(clientset kubernetes.Clientset, fioPodName string, pvcName string, volumeType common.VolumeType, volSize int, fioLoops int) error {

	vname := "ms-volume"
	volume := coreV1.Volume{
		Name: vname,
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
	}
	var volumes []coreV1.Volume
	var volMounts []coreV1.VolumeMount
	var volDevices []coreV1.VolumeDevice
	var volFioArgs [][]string
	volumes = append(volumes, volume)

	if volumeType == common.VolFileSystem {
		mount := coreV1.VolumeMount{
			Name:      vname,
			MountPath: "/volume",
		}
		volMounts = append(volMounts, mount)
		volFioArgs = append(volFioArgs, []string{
			fmt.Sprintf("--name=%s", vname),
			fmt.Sprintf("--filename=/volume-%s.test", vname),
		})
	} else {
		device := coreV1.VolumeDevice{
			Name:       vname,
			DevicePath: "/dev/sdm",
		}
		volDevices = append(volDevices, device)
		volFioArgs = append(volFioArgs, []string{
			fmt.Sprintf("--name=%s", vname),
			"--filename=/dev/sdm",
		})
	}

	volFioArgs = append(volFioArgs, []string{
		fmt.Sprintf("--filename=%s", common.FioBlockFilename),
	})

	// Create the fio Pod

	// Construct argument list for fio to run a single instance of fio,
	// with multiple jobs, one for each volume.
	var podArgs []string

	// 1) directives for all fio jobs
	podArgs = append(podArgs, []string{"---", "fio"}...)
	podArgs = append(podArgs, common.GetDefaultFioArguments()...)

	if volumeType == common.VolFileSystem {
		// for FS play safe use filesize which is 75% of volume size
		podArgs = append(podArgs, fmt.Sprintf("--size=%dm", (volSize*75)/100))
	}

	if fioLoops != 0 {
		podArgs = append(podArgs, fmt.Sprintf("--loops=%d", fioLoops))
	}

	// 2) per volume directives
	for _, v := range volFioArgs {
		podArgs = append(podArgs, v...)
	}

	// e2e-fio commandline is
	logf.Log.Info(fmt.Sprintf("commandline: %s", strings.Join(podArgs[1:], " ")))
	podArgs = append(podArgs, "&")
	logf.Log.Info("pod", "args", podArgs)

	container := lib.MakeFioContainer(fioPodName, podArgs)
	podBuilder := lib.NewPodBuilder().
		WithName(fioPodName).
		WithNamespace(common.NSDefault).
		WithContainer(container).
		WithVolumes(volumes)

	if len(volDevices) != 0 {
		podBuilder.WithVolumeDevices(volDevices)
	}

	if len(volMounts) != 0 {
		podBuilder.WithVolumeMounts(volMounts)
	}

	podObj, err := podBuilder.Build()
	if err != nil {
		return fmt.Errorf("failed to build fio test pod object, error: %v", err)
	}

	pod, err := lib.CreatePod(clientset, podObj, common.NSDefault)
	if err != nil {
		return fmt.Errorf("failed to create fio pod, error: %v", err)
	}
	if pod == nil {
		return fmt.Errorf("failed to create fio pod, pod nil")
	}

	// Wait for the fio Pod to transition to running
	const timoSecs = 120
	const timoSleepSecs = 10
	for ix := 0; ; ix++ {
		if lib.IsPodRunning(clientset, fioPodName, common.NSDefault) {
			break
		}
		if ix >= timoSecs/timoSleepSecs {
			return fmt.Errorf("timed out waiting for pod to be running")
		}
		time.Sleep(timoSleepSecs * time.Second)
	}

	logf.Log.Info("Waiting for run to complete", "timeout", timoSecs)
	return nil
}

func (testConductor TestConductor) BasicSoakTest() error {
	if err := InstallMayastor(testConductor.clientset); err != nil {
		return fmt.Errorf("failed to install mayastor %v", err)
	}
	var protocol common.ShareProto = common.ShareProtoNvmf
	//var volumeType common.VolumeType
	var mode storageV1.VolumeBindingMode = storageV1.VolumeBindingImmediate
	// create storage class
	err := lib.NewScBuilder().
		WithName("basic-soak-sc" /*scName*/).
		WithReplicas(3 /*common.DefaultReplicaCount*/).
		WithProtocol(protocol).
		WithNamespace(common.NSDefault).
		WithVolumeBindingMode(mode).
		BuildAndCreate(testConductor.clientset)
	if err != nil {
		return fmt.Errorf("failed to create sc %v", err)
	}
	fmt.Printf("created sc\n")

	// create PV
	pvcname, err := lib.MkPVC(testConductor.clientset, 64, "basic-soak-pvc", "basic-soak-sc", common.VolRawBlock, "default")
	if err != nil {
		return fmt.Errorf("failed to create pvc %v", err)
	}
	fmt.Printf("created pvc %s\n", pvcname)

	// deploy fio
	err = deployFio(testConductor.clientset, "basic-soak-fio", "basic-soak-pvc", common.VolRawBlock, 64, 1)
	if err != nil {
		return fmt.Errorf("failed to deploy pod %v", err)
	}
	// alert workload monitor
	time.Sleep(600 * time.Second)
	return err
}
