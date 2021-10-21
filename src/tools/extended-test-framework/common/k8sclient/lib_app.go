package k8sclient

import (
	"fmt"
	"strings"
	"time"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// MakeFioContainer returns a container object setup to use e2e-fio and run fio with appropriate permissions.
// Privileged: True, AllowPrivilegeEscalation: True, RunAsUser root,
// parameters:
//		name - name of the container (usually the pod name)
//		args - container arguments, if empty the defaults to "sleep", "1000000"
func MakeFioContainer(name string, args []string) coreV1.Container {
	containerArgs := args
	if len(containerArgs) == 0 {
		containerArgs = []string{"sleep", "1000000"}
	}
	var z64 int64 = 0
	var vTrue bool = true

	sc := coreV1.SecurityContext{
		Privileged:               &vTrue,
		RunAsUser:                &z64,
		AllowPrivilegeEscalation: &vTrue,
	}
	return coreV1.Container{
		Name:            name,
		Image:           GetFioImage(),
		ImagePullPolicy: coreV1.PullIfNotPresent,
		Args:            containerArgs,
		SecurityContext: &sc,
	}
}

func DeployFio(
	fioPodName string,
	pvcName string,
	volumeType VolumeType,
	volSize int,
	fioLoops int,
	thinkTime int,
	thinkTimeBlocks int) error {

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

	if thinkTime > 0 {
		volFioArgs = append(volFioArgs, []string{
			fmt.Sprintf("--thinktime=%d", thinkTime),
		})
	}
	if thinkTimeBlocks > 0 {
		volFioArgs = append(volFioArgs, []string{
			fmt.Sprintf("--thinktime_blocks=%d", thinkTimeBlocks),
		})
	}

	if volumeType == VolFileSystem {
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

	// Create the fio Pod

	// Construct argument list for fio to run a single instance of fio,
	// with multiple jobs, one for each volume.
	var podArgs []string

	// 1) directives for all fio jobs
	podArgs = append(podArgs, []string{"---", "fio"}...)
	podArgs = append(podArgs, GetDefaultFioArguments()...)

	if volumeType == VolFileSystem {
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

	container := MakeFioContainer(fioPodName, podArgs)
	podBuilder := NewPodBuilder().
		WithName(fioPodName).
		WithNamespace(NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(container).
		WithVolumes(volumes).
		WithAppLabel("fio")

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

	pod, err := CreatePod(podObj, NSDefault)
	if err != nil {
		return fmt.Errorf("failed to create fio pod, error: %v", err)
	}
	if pod == nil {
		return fmt.Errorf("failed to create fio pod, pod nil")
	}

	// Wait for the fio Pod to transition to running
	const timoSecs = 1000
	const timoSleepSecs = 10
	for ix := 0; ; ix++ {
		if IsPodRunning(fioPodName, NSDefault) {
			break
		}
		if ix >= timoSecs/timoSleepSecs {
			return fmt.Errorf("timed out waiting for pod to be running")
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	return nil
}
