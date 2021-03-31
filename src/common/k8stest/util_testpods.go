package k8stest

// Utility functions for test pods.
import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"mayastor-e2e/common"

	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// FIXME: this function runs fio with a bunch of parameters which are not configurable.
// sizeMb should be 0 for fio to use the entire block device
func RunFio(podName string, duration int, filename string, sizeMb int, args ...string) ([]byte, error) {
	argRuntime := fmt.Sprintf("--runtime=%d", duration)
	argFilename := fmt.Sprintf("--filename=%s", filename)

	logf.Log.Info("RunFio",
		"podName", podName,
		"duration", duration,
		"filename", filename,
		"args", args)

	cmdArgs := []string{
		"exec",
		"-it",
		podName,
		"--",
		"fio",
		"--name=benchtest",
		"--verify=crc32",
		"--verify_fatal=1",
		"--verify_async=2",
		argFilename,
		"--direct=1",
		"--rw=randrw",
		"--ioengine=libaio",
		"--bs=4k",
		"--iodepth=16",
		"--numjobs=1",
		"--time_based",
		argRuntime,
	}

	if sizeMb != 0 {
		sizeArgs := []string{fmt.Sprintf("--size=%dm", sizeMb)}
		cmdArgs = append(cmdArgs, sizeArgs...)
	}

	if args != nil {
		cmdArgs = append(cmdArgs, args...)
	}

	cmd := exec.Command(
		"kubectl",
		cmdArgs...,
	)
	cmd.Dir = ""
	output, err := cmd.CombinedOutput()
	if err != nil {
		logf.Log.Info("Running fio failed", "error", err)
	}
	return output, err
}

func IsPodRunning(podName string, nameSpace string) bool {
	var pod coreV1.Pod
	if gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: podName, Namespace: nameSpace}, &pod) != nil {
		return false
	}
	return pod.Status.Phase == v1.PodRunning
}

/// Create a Pod in default namespace, no options and no context
func CreatePod(podDef *coreV1.Pod, nameSpace string) (*coreV1.Pod, error) {
	logf.Log.Info("Creating", "pod", podDef.Name)
	return gTestEnv.KubeInt.CoreV1().Pods(nameSpace).Create(context.TODO(), podDef, metaV1.CreateOptions{})
}

/// Delete a Pod in default namespace, no options and no context
func DeletePod(podName string, nameSpace string) error {
	logf.Log.Info("Deleting", "pod", podName)
	return gTestEnv.KubeInt.CoreV1().Pods(nameSpace).Delete(context.TODO(), podName, metaV1.DeleteOptions{})
}

/// Create a test fio pod in default namespace, no options and no context
/// for filesystem,  mayastor volume is mounted on /volume
/// for rawblock, mayastor volume is mounted on /dev/sdm
func CreateFioPodDef(podName string, volName string, volType common.VolumeType, nameSpace string) *coreV1.Pod {
	volMounts := []coreV1.VolumeMount{
		{
			Name:      "ms-volume",
			MountPath: common.FioFsMountPoint,
		},
	}
	volDevices := []coreV1.VolumeDevice{
		{
			Name:       "ms-volume",
			DevicePath: common.FioBlockFilename,
		},
	}

	podDef := coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      podName,
			Namespace: nameSpace,
			Labels:    map[string]string{"app": "fio"},
		},
		Spec: coreV1.PodSpec{
			RestartPolicy: coreV1.RestartPolicyNever,
			Containers: []coreV1.Container{
				{
					Name:  podName,
					Image: "mayadata/e2e-fio",
					Args:  []string{"sleep", "1000000"},
				},
			},
			Volumes: []coreV1.Volume{
				{
					Name: "ms-volume",
					VolumeSource: coreV1.VolumeSource{
						PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
							ClaimName: volName,
						},
					},
				},
			},
		},
	}
	if volType == common.VolRawBlock {
		podDef.Spec.Containers[0].VolumeDevices = volDevices
	} else {
		podDef.Spec.Containers[0].VolumeMounts = volMounts
	}
	return &podDef
}

/// Create a test fio pod in default namespace, no options and no context
/// mayastor volume is mounted on /volume
func CreateFioPod(podName string, volName string, volType common.VolumeType, nameSpace string) (*coreV1.Pod, error) {
	logf.Log.Info("Creating fio pod definition", "name", podName, "volume type", volType)
	podDef := CreateFioPodDef(podName, volName, volType, nameSpace)
	return CreatePod(podDef, common.NSDefault)
}

// Check if any test pods exist in the default and e2e related namespaces .
func CheckForTestPods() (bool, error) {
	logf.Log.Info("CheckForTestPods")
	foundPods := false

	nameSpaces, err := gTestEnv.KubeInt.CoreV1().Namespaces().List(context.TODO(), metaV1.ListOptions{})
	if err == nil {
		for _, ns := range nameSpaces.Items {
			if strings.HasPrefix(ns.Name, common.NSE2EPrefix) || ns.Name == common.NSDefault {
				pods, err := gTestEnv.KubeInt.CoreV1().Pods(ns.Name).List(context.TODO(), metaV1.ListOptions{})
				if err == nil && pods != nil && len(pods.Items) != 0 {
					logf.Log.Info("CheckForTestPods",
						"Pods", pods.Items)
					foundPods = true
				}
			}
		}
	}

	return foundPods, err
}

// Check test pods in a namespace for restarts and failed/unknown state
func CheckTestPodsHealth(namespace string) error {
	podApi := gTestEnv.KubeInt.CoreV1().Pods
	var errorStrings []string
	podList, err := podApi(namespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return errors.New("failed to list pods")
	}

	for _, pod := range podList.Items {
		containerStatuses := pod.Status.ContainerStatuses
		for _, containerStatus := range containerStatuses {
			if containerStatus.RestartCount != 0 {
				logf.Log.Info(pod.Name, "restarts", containerStatus.RestartCount)
				errorStrings = append(errorStrings, fmt.Sprintf("%s restarted %d times", pod.Name, containerStatus.RestartCount))
			}
			if pod.Status.Phase == coreV1.PodFailed || pod.Status.Phase == coreV1.PodUnknown {
				logf.Log.Info(pod.Name, "phase", pod.Status.Phase)
				errorStrings = append(errorStrings, fmt.Sprintf("%s phase is %v", pod.Name, pod.Status.Phase))
			}
		}
	}

	if len(errorStrings) != 0 {
		return errors.New(strings.Join(errorStrings[:], "; "))
	}
	return nil
}

func CheckPodCompleted(podName string, nameSpace string) (coreV1.PodPhase, error) {
	podApi := gTestEnv.KubeInt.CoreV1().Pods
	pod, err := podApi(nameSpace).Get(context.TODO(), podName, metaV1.GetOptions{})
	if err != nil {
		return coreV1.PodUnknown, err
	}
	return pod.Status.Phase, err
}
