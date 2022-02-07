package k8stest

// Utility functions for test pods.
import (
	"context"
	"errors"
	"fmt"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/mayastorclient"
	"os/exec"
	"strings"
	"time"

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

// IsPodWithLabelsRunning expects that at any time only one application pod will be in running state
// if there are more then one pod in terminating state then it will return the last terminating pod.
func IsPodWithLabelsRunning(labels, namespace string) (string, bool, error) {
	var podName string
	pods, err := gTestEnv.KubeInt.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{LabelSelector: labels})
	if err != nil {
		return "", false, err
	}
	if len(pods.Items) == 0 {
		return "", false, nil
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			return pod.Name, true, nil
		}
		podName = pod.Name
	}
	return podName, false, nil
}

// ForceDeleteTerminatingPods force deletes the pod this function is required because
// sometimes after powering off the node some pods stuck in terminating state.
func ForceDeleteTerminatingPods(labels, namespace string) error {
	deletionTime := int64(0)
	pods, err := gTestEnv.KubeInt.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{LabelSelector: labels})
	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		return nil
	}
	for _, pod := range pods.Items {
		if pod.DeletionTimestamp != nil {
			err = gTestEnv.KubeInt.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, metaV1.DeleteOptions{GracePeriodSeconds: &deletionTime})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func GetNodeListForPods(labels, namespace string) (map[string]v1.PodPhase, error) {
	pods, err := gTestEnv.KubeInt.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{LabelSelector: labels})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, nil
	}
	nodeList := map[string]v1.PodPhase{}
	for _, pod := range pods.Items {
		nodeList[pod.Spec.NodeName] = pod.Status.Phase
	}
	return nodeList, nil
}

func IsPodRunning(podName string, nameSpace string) bool {
	var pod coreV1.Pod
	if gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: podName, Namespace: nameSpace}, &pod) != nil {
		return false
	}
	return pod.Status.Phase == v1.PodRunning
}

// WaitPodRunning wait for pod to transition to running with timeout,
// returns true of the pod is running, false otherwise.
func WaitPodRunning(podName string, nameSpace string, timeoutSecs int) bool {
	const sleepTime = 3
	for ix := 0; ix < (timeoutSecs+sleepTime-1)/sleepTime && !IsPodRunning(podName, nameSpace); ix++ {
		time.Sleep(sleepTime * time.Second)
	}
	return IsPodRunning(podName, nameSpace)
}

func GetPodScheduledStatus(podName string, nameSpace string) (coreV1.ConditionStatus, string, error) {
	var pod coreV1.Pod
	if gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: podName, Namespace: nameSpace}, &pod) != nil {
		return coreV1.ConditionUnknown, "", fmt.Errorf("failed to get pod")
	}
	status := pod.Status
	for _, condition := range status.Conditions {
		if condition.Type == coreV1.PodScheduled {
			return condition.Status, condition.Reason, nil
		}
	}
	return coreV1.ConditionUnknown, "", fmt.Errorf("failed to find pod scheduled condition")
}

// CreatePod Create a Pod in the specified namespace, no options and no context
func CreatePod(podDef *coreV1.Pod, nameSpace string) (*coreV1.Pod, error) {
	logf.Log.Info("Creating", "pod", podDef.Name)
	return gTestEnv.KubeInt.CoreV1().Pods(nameSpace).Create(context.TODO(), podDef, metaV1.CreateOptions{})
}

// DeletePod Delete a Pod in the specified namespace, no options and no context
func DeletePod(podName string, nameSpace string) error {
	logf.Log.Info("Deleting", "pod", podName)
	return gTestEnv.KubeInt.CoreV1().Pods(nameSpace).Delete(context.TODO(), podName, metaV1.DeleteOptions{})
}

//CreateFioPodDef  deprecated use MakeFioContainer and NewPodBuilder instead
/// Create a test fio pod in default namespace, no options and no context
/// for filesystem,  mayastor volume is mounted on /volume
/// for raw-block, mayastor volume is mounted on /dev/sdm
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
					Name:            podName,
					Image:           common.GetFioImage(),
					ImagePullPolicy: coreV1.PullAlways,
					Args:            []string{"sleep", "1000000"},
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
	if e2e_config.GetConfig().Platform.HostNetworkingRequired {
		podDef.Spec.HostNetwork = true
	}
	if volType == common.VolRawBlock {
		podDef.Spec.Containers[0].VolumeDevices = volDevices
	}
	if volType == common.VolFileSystem {
		podDef.Spec.Containers[0].VolumeMounts = volMounts
	}
	return &podDef
}

//CreateFioPod deprecated use MakeFioContainer, NewPodBuilder and CreatePod instead
/// Create a test fio pod in default namespace, no options and no context
/// mayastor volume is mounted on /volume
func CreateFioPod(podName string, volName string, volType common.VolumeType, nameSpace string) (*coreV1.Pod, error) {
	logf.Log.Info("Creating fio pod definition", "name", podName, "volume type", volType)
	podDef := CreateFioPodDef(podName, volName, volType, nameSpace)
	return CreatePod(podDef, common.NSDefault)
}

//CheckForTestPods Check if any test pods exist in the default and e2e related namespaces .
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

// isPodHealthCheckCandidate is a filter function for health check on pod,
func isPodHealthCheckCandidate(podName string, namespace string) bool {
	if namespace == common.NSMayastor() {
		return !strings.HasPrefix(podName, "mayastor-etcd")
	}
	return true
}

//CheckTestPodsHealth Check test pods in a namespace for restarts and failed/unknown state
func CheckTestPodsHealth(namespace string) error {
	podApi := gTestEnv.KubeInt.CoreV1().Pods
	var errorStrings []string
	podList, err := podApi(namespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return errors.New("failed to list pods")
	}

	for _, pod := range podList.Items {
		if !isPodHealthCheckCandidate(pod.Name, namespace) {
			continue
		}
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

	// Keeping this commented out code for the time being.
	//	podApi := gTestEnv.KubeInt.CoreV1().Pods
	//	pod, err := podApi(nameSpace).Get(context.TODO(), podName, metaV1.GetOptions{})
	//	if err != nil {
	//		return coreV1.PodUnknown, err
	//	}
	//	return pod.Status.Phase, err
	//
	return CheckPodContainerCompleted(podName, nameSpace)
}

// GetPodHostIp retrieve the IP address  of the node hosting a pod
func GetPodHostIp(podName string, nameSpace string) (string, error) {
	podApi := gTestEnv.KubeInt.CoreV1().Pods
	pod, err := podApi(nameSpace).Get(context.TODO(), podName, metaV1.GetOptions{})
	if err != nil {
		return "", err
	}
	return pod.Status.HostIP, err
}

func CheckPodContainerCompleted(podName string, nameSpace string) (coreV1.PodPhase, error) {
	podApi := gTestEnv.KubeInt.CoreV1().Pods
	pod, err := podApi(nameSpace).Get(context.TODO(), podName, metaV1.GetOptions{})
	if err != nil {
		return coreV1.PodUnknown, err
	}
	containerStatuses := pod.Status.ContainerStatuses
	for _, containerStatus := range containerStatuses {
		if containerStatus.Name == podName {
			if !containerStatus.Ready {
				if containerStatus.State.Terminated != nil &&
					containerStatus.State.Terminated.Reason == "Completed" {
					if containerStatus.State.Terminated.ExitCode == 0 {
						return coreV1.PodSucceeded, nil
					} else {
						return coreV1.PodFailed, nil
					}
				}
			}
		}
	}
	return pod.Status.Phase, err
}

var mayastorInitialPodCount int

func SetMayastorInitialPodCount(count int) {
	mayastorInitialPodCount = count
}

func GetMayastorInitialPodCount() int {
	return mayastorInitialPodCount
}

// List mayastor pod names, conditionally
//  1) No timestamp - all mayastor pods
//	2) With timestamp - all mayastor pods created after the timestamp which are Running.
func listMayastorPods(timestamp *time.Time) ([]string, error) {
	var podNames []string
	podApi := gTestEnv.KubeInt.CoreV1().Pods
	pods, err := podApi(common.NSMayastor()).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return podNames, err
	}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, "mayastor-etcd") {
			continue
		}
		// If timestamp != nil, then we return the list of pods which are both
		//	1. running
		//  2. created after the timestamp

		if timestamp != nil {
			if pod.Status.Phase != v1.PodRunning {
				continue
			}
			cs := pod.GetCreationTimestamp()
			if !cs.After(*timestamp) {
				continue
			}
		}
		podNames = append(podNames, pod.Name)
	}
	return podNames, nil
}

// RestartMayastorPods shortcut to reinstalling mayastor, especially useful on platforms
// like volterra, for example calling this function after patching the installation to
// use different mayastor images, should allow us to have reasonable confidence that
// mayastor has been restarted with those images.
// Deletes all mayastor pods except for mayastor etcd pods,
// then waits upto specified time for new pods to be provisioned
// Simply deleting the pods and then waiting for daemonset ready checks do not work due to k8s latencies,
// for example it has been observed the mayastor-csi pods are deemed ready
// because they enter terminating state after we've checked for readiness
// Caller must perform readiness checks after calling this function.
func RestartMayastorPods(timeoutSecs int) error {
	var err error
	podApi := gTestEnv.KubeInt.CoreV1().Pods

	podNames, err := listMayastorPods(nil)
	if err != nil {
		return err
	}

	logf.Log.Info("Restarting", "pods", podNames)
	now := time.Now()
	time.Sleep(1 * time.Second)
	for _, podName := range podNames {
		delErr := podApi(common.NSMayastor()).Delete(context.TODO(), podName, metaV1.DeleteOptions{})
		if delErr != nil {
			logf.Log.Info("Failed to delete", "pod", podName, "error", delErr)
			err = delErr
		} else {
			logf.Log.Info("Deleted", "pod", podName)
		}
	}

	if err != nil {
		return err
	}
	var newPodNames []string
	const sleepTime = 10
	// Wait (with timeout) for all pods to have restarted
	logf.Log.Info("Waiting for all pods to restart", "timeoutSecs", timeoutSecs)
	for ix := 1; ix < (timeoutSecs+sleepTime-1)/sleepTime; ix++ {
		time.Sleep(sleepTime * time.Second)
		newPodNames, err = listMayastorPods(&now)
		if err == nil {
			logf.Log.Info("Restarted", "pods", newPodNames)
			if len(newPodNames) >= GetMayastorInitialPodCount() {
				logf.Log.Info("All pods have been restarted.")
				return nil
			}
		}
	}
	logf.Log.Info("Restart pods failed", "oldpods", podNames, "newpods", newPodNames)
	return fmt.Errorf("restart failed incomplete error=%v", err)
}

func collectNatsPodNames() ([]string, error) {
	var podNames []string
	podApi := gTestEnv.KubeInt.CoreV1().Pods
	pods, err := podApi(common.NSMayastor()).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return podNames, err
	}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, "nats") {
			podNames = append(podNames, pod.Name)
		}
	}
	return podNames, nil
}

// RestartNatsPods restart the nats pods
func RestartNatsPods(timeoutSecs int) error {
	var err error
	podApi := gTestEnv.KubeInt.CoreV1().Pods

	podNames, err := collectNatsPodNames()
	if err != nil {
		return err
	}

	for _, podName := range podNames {
		delErr := podApi(common.NSMayastor()).Delete(context.TODO(), podName, metaV1.DeleteOptions{})
		if delErr != nil {
			logf.Log.Info("Failed to delete", "pod", podName, "error", delErr)
			err = delErr
		} else {
			logf.Log.Info("Deleted", "pod", podName)
		}
	}

	if err != nil {
		return err
	}
	const sleepTime = 5
	// Wait (with timeout) for all pods to have restarted
	// For this to work we rely on the fact that for daemonsets and deployments,
	// when a pod is deleted, k8s spins up a new pod with a different name.
	// So the check is comparison between
	//      1) the list of nats pods deleted
	//      2) a freshly generated list of nats pods
	// - the size of the fresh list >= size of the deleted list
	// - the names of the pods deleted do not occur in the fresh list
	for ix := 1; ix < (timeoutSecs+sleepTime-1)/sleepTime; ix++ {
		newPodNames, err := collectNatsPodNames()
		if err == nil {
			if len(podNames) <= len(newPodNames) {
				return nil
			}
			time.Sleep(sleepTime * time.Second)
		}
	}
	return fmt.Errorf("restart failed in some nebulous way! ")
}

func restartMayastor(restartTOSecs int, readyTOSecs int, poolsTOSecs int) error {
	var err error
	ready := false

	// note: CleanUp removes replicas
	CleanUp()
	_, _ = DeleteAllPoolFinalizers()
	_ = DeleteAllPools()

	err = RestartMayastorPods(restartTOSecs)
	if err != nil {
		return fmt.Errorf("RestartMayastorPods failed %v", err)
	}

	ready, err = MayastorReady(5, readyTOSecs)
	if err != nil {
		return fmt.Errorf("failure waiting for mayastor to be ready %v", err)
	}
	if !ready {
		return fmt.Errorf("mayastor is not ready after deleting all pods")
	}

	// Pause to allow things to settle.
	time.Sleep(30 * time.Second)

	_, _ = DeleteAllPoolFinalizers()
	// Pool device is not specified so do not re-create the pools
	if len(e2e_config.GetConfig().PoolDevice) != 0 {
		_ = DeleteAllPools()
		err = CreateConfiguredPools()
		if err != nil {
			return err
		}
	}

	const sleepTime = 10
	for ix := 0; ix < (poolsTOSecs+sleepTime-1)/sleepTime; ix++ {
		time.Sleep(sleepTime * time.Second)
		err = custom_resources.CheckAllMsPoolsAreOnline()
		if err == nil {
			break
		}
	}

	err = custom_resources.CheckAllMsPoolsAreOnline()
	if err != nil {
		return fmt.Errorf("not all pools are online %v", err)
	}

	if mayastorclient.CanConnect() {
		err := RmReplicasInCluster()
		if err != nil {
			return fmt.Errorf("RmReplicasInCluster failed %v", err)
		}
	} else {
		logf.Log.Info("WARNING: gRPC calls to mayastor are not enabled, unable to clear orphan replicas")
	}
	return err
}

// RestartMayastor this function "restarts" mayastor by
//	- cleaning up all mayastor resource artefacts,
//  - deleting all mayastor pods
func RestartMayastor(restartTOSecs int, readyTOSecs int, poolsTOSecs int) error {
	var err error
	const restartRetries = 3
	// try to restart upto N times
	// chiefly this is a fudge to get restart to work on the volterra platform
	for ix := 1; ix <= restartRetries; ix++ {

		logf.Log.Info("Restarting mayastor", "try", fmt.Sprintf("%d/%d", ix, restartRetries))

		if err = restartMayastor(restartTOSecs, restartTOSecs, poolsTOSecs); err != nil {
			logf.Log.Info("Restarting failed, failed to restart pods", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if err = CheckTestPodsHealth(common.NSMayastor()); err != nil {
			logf.Log.Info("Restarting failed, pods are not healthy", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if err = ResourceCheck(); err != nil {
			logf.Log.Info("Restarting failed, resource check failed", "error", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// @here all restarts and checks passed
		logf.Log.Info("Restarted mayastor successfully")
		break
	}

	return err
}

func GetCoreAgentNodeName() (string, error) {
	podApi := gTestEnv.KubeInt.CoreV1().Pods
	pods, err := podApi(common.NSMayastor()).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return "", err
	}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, "core-agent") && pod.Status.Phase == "Running" {
			return pod.Spec.NodeName, nil
		}
	}
	return "", nil
}

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
		Image:           common.GetFioImage(),
		ImagePullPolicy: coreV1.PullIfNotPresent,
		Args:            containerArgs,
		SecurityContext: &sc,
	}
}

// MakeFsxContainer returns a container object setup to use e2e-fsx and run fsx with appropriate permissions.
// Privileged: True, AllowPrivilegeEscalation: True, RunAsUser root
func MakeFsxContainer(name string, args []string) coreV1.Container {
	var z64 int64 = 0
	var vTrue bool = true

	sc := coreV1.SecurityContext{
		Privileged:               &vTrue,
		RunAsUser:                &z64,
		AllowPrivilegeEscalation: &vTrue,
	}
	return coreV1.Container{
		Name:            name,
		Image:           common.GetFsxImage(),
		ImagePullPolicy: coreV1.PullIfNotPresent,
		Args:            args,
		SecurityContext: &sc,
	}
}
