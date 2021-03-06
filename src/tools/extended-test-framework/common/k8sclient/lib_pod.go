package k8sclient

import (
	// container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	// volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"

	"context"
	"fmt"
	"mayastor-e2e/tools/extended-test-framework/common"
	"strings"
	"time"

	errors "github.com/pkg/errors"
	coreV1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// k8sNodeLabelKeyHostname is the label key used by Kubernetes
	// to store the hostname on the node resource.
	K8sNodeLabelKeyHostname = "kubernetes.io/hostname"
)

type Pod struct {
	object *corev1.Pod
}

// PodBuilder is the builder object for Pod
type PodBuilder struct {
	pod  *Pod
	errs []error
}

func IsPodRunning(podName string, nameSpace string) bool {
	pod, err := gKubeInt.CoreV1().Pods(nameSpace).Get(context.TODO(), podName, metaV1.GetOptions{})
	if err != nil {
		return false
	}
	return pod.Status.Phase == v1.PodRunning
}

func IsPodFailed(podName string, nameSpace string) bool {
	pod, err := gKubeInt.CoreV1().Pods(nameSpace).Get(context.TODO(), podName, metaV1.GetOptions{})
	if err != nil {
		return false
	}
	return pod.Status.Phase == v1.PodFailed
}

// NewPodBuilder returns new instance of Builder
func NewPodBuilder() *PodBuilder {
	return &PodBuilder{pod: &Pod{object: &corev1.Pod{}}}
}

// WithTolerationsForTaints sets the Spec.Tolerations with provided taints.
func (b *PodBuilder) WithTolerationsForTaints(taints ...corev1.Taint) *PodBuilder {

	tolerations := []corev1.Toleration{}
	for i := range taints {
		var toleration corev1.Toleration
		toleration.Key = taints[i].Key
		toleration.Effect = taints[i].Effect
		if len(taints[i].Value) == 0 {
			toleration.Operator = corev1.TolerationOpExists
		} else {
			toleration.Value = taints[i].Value
			toleration.Operator = corev1.TolerationOpEqual
		}
		tolerations = append(tolerations, toleration)
	}

	b.pod.object.Spec.Tolerations = append(
		b.pod.object.Spec.Tolerations,
		tolerations...,
	)
	return b
}

// WithName sets the Name field of Pod with provided value.
func (b *PodBuilder) WithName(name string) *PodBuilder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing Pod name"),
		)
		return b
	}
	b.pod.object.Name = name
	return b
}

// WithNamespace sets the Namespace field of Pod with provided value.
func (b *PodBuilder) WithNamespace(namespace string) *PodBuilder {
	if len(namespace) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing namespace"),
		)
		return b
	}
	b.pod.object.Namespace = namespace
	return b
}

// WithNamespace sets the Namespace field of Pod with provided value.
func (b *PodBuilder) WithLabels(labels map[string]string) *PodBuilder {
	if len(labels) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing labels"),
		)
		return b
	}
	b.pod.object.Labels = labels
	return b
}

func (b *PodBuilder) WithAppLabel(applabel string) *PodBuilder {
	label := make(map[string]string)
	label["app"] = applabel
	b.pod.object.Labels = label
	return b
}

// WithRestartPolicy sets the RestartPolicy field in Pod with provided arguments
func (b *PodBuilder) WithRestartPolicy(
	restartPolicy corev1.RestartPolicy,
) *PodBuilder {
	b.pod.object.Spec.RestartPolicy = restartPolicy
	return b
}

// WithNodeName sets the NodeName field of Pod with provided value.
func (b *PodBuilder) WithNodeName(nodeName string) *PodBuilder {
	if len(nodeName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing Pod node name"),
		)
		return b
	}
	b.pod.object.Spec.NodeName = nodeName
	return b
}

// WithNodeSelectorHostnameNew sets the Pod NodeSelector to the provided hostname value
// This function replaces (resets) the NodeSelector to use only hostname selector
func (b *PodBuilder) WithNodeSelectorHostnameNew(hostname string) *PodBuilder {
	if len(hostname) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing Pod hostname"),
		)
		return b
	}

	b.pod.object.Spec.NodeSelector = map[string]string{
		K8sNodeLabelKeyHostname: hostname,
	}

	return b
}

// WithContainers sets the Containers field in Pod with provided arguments
func (b *PodBuilder) WithContainers(containers []corev1.Container) *PodBuilder {
	if len(containers) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing containers"),
		)
		return b
	}
	b.pod.object.Spec.Containers = containers
	return b
}

// WithContainer sets the Containers field in Pod with provided arguments
func (b *PodBuilder) WithContainer(container corev1.Container) *PodBuilder {
	return b.WithContainers([]corev1.Container{container})
}

// WithVolumes sets the Volumes field in Pod with provided arguments
func (b *PodBuilder) WithVolumes(volumes []corev1.Volume) *PodBuilder {
	if len(volumes) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing volumes"),
		)
		return b
	}
	b.pod.object.Spec.Volumes = volumes
	return b
}

// WithVolume sets the Volumes field in Pod with provided arguments
func (b *PodBuilder) WithVolume(volume corev1.Volume) *PodBuilder {
	return b.WithVolumes([]corev1.Volume{volume})
}

// WithServiceAccountName sets the ServiceAccountName of Pod spec with
// the provided value
func (b *PodBuilder) WithServiceAccountName(serviceAccountName string) *PodBuilder {
	if len(serviceAccountName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing Pod service account name"),
		)
		return b
	}
	b.pod.object.Spec.ServiceAccountName = serviceAccountName
	return b
}

// WithVolumeMounts sets the Volumes field in Pod with provided arguments
func (b *PodBuilder) WithVolumeMounts(volMounts []corev1.VolumeMount) *PodBuilder {
	if len(volMounts) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing VolumeMount"),
		)
		return b
	}
	b.pod.object.Spec.Containers[0].VolumeMounts = volMounts
	return b
}

// WithVolumeMount sets the Volumes field in Pod with provided arguments
func (b *PodBuilder) WithVolumeMount(volMount corev1.VolumeMount) *PodBuilder {
	return b.WithVolumeMounts([]corev1.VolumeMount{volMount})
}

// WithVolumeDevices sets the Volumes field in Pod with provided arguments
func (b *PodBuilder) WithVolumeDevices(volDevices []corev1.VolumeDevice) *PodBuilder {
	if len(volDevices) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Pod object: missing VolumeDevices"),
		)
		return b
	}
	b.pod.object.Spec.Containers[0].VolumeDevices = volDevices
	return b
}

// WithVolumeDevice sets the Volumes field in Pod with provided arguments
func (b *PodBuilder) WithVolumeDevice(volDevice corev1.VolumeDevice) *PodBuilder {
	return b.WithVolumeDevices([]corev1.VolumeDevice{volDevice})
}

func (b *PodBuilder) WithVolumeDeviceOrMount(volType VolumeType) *PodBuilder {
	volMounts := coreV1.VolumeMount{
		Name:      "ms-volume",
		MountPath: FioFsMountPoint,
	}
	volDevices := coreV1.VolumeDevice{
		Name:       "ms-volume",
		DevicePath: FioBlockFilename,
	}
	if volType == VolRawBlock {
		b.WithVolumeDevice(volDevices)
	} else {
		b.WithVolumeMount(volMounts)
	}

	return b
}

// Build returns the Pod API instance
func (b *PodBuilder) Build() (*corev1.Pod, error) {
	//if e2e_config.GetConfig().Platform.HostNetworkingRequired {
	//	b.pod.object.Spec.HostNetwork = true
	//}
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.pod.object, nil
}

// CreatePod Create a Pod in the specified namespace, no options and no context
func CreatePod(podDef *coreV1.Pod, nameSpace string) (*coreV1.Pod, error) {
	logf.Log.Info("Creating", "pod", podDef.Name)
	return gKubeInt.CoreV1().Pods(nameSpace).Create(context.TODO(), podDef, metaV1.CreateOptions{})
}

// DeletePod Delete a Pod in the specified namespace, no options and no context
func DeletePod(name string, nameSpace string, timeoutSecs int) error {
	logf.Log.Info("Deleting", "pod", name)
	const timoSleepSecs = 10
	err := gKubeInt.CoreV1().Pods(nameSpace).Delete(context.TODO(), name, metaV1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Failed to delete pod, error: %+v", err)
	}
	for i := 0; ; i += timoSleepSecs {
		exists, err := GetPodExistsByName(name, nameSpace)
		if err != nil {
			return fmt.Errorf("Error determining if pod %s exists, error: %v", name, err)
		}
		if !exists {
			logf.Log.Info("Deleted", "pod", name)
			return nil
		}
		if i >= timeoutSecs {
			return fmt.Errorf("Timed out waiting for pod %s to be deleted", name)
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
}

// CheckPodAndDelete: Delete a Pod if it completed with no container errors
func CheckPodAndDelete(podName string, nameSpace string, timeoutSecs int) error {
	pod, err := gKubeInt.CoreV1().Pods(nameSpace).Get(context.TODO(), podName, metaV1.GetOptions{})
	if err != nil {
		return err
	}
	if pod.Status.Phase != v1.PodSucceeded {
		return fmt.Errorf("Pod failed, status %s", pod.Status.Phase)
	}
	return DeletePod(podName, nameSpace, timeoutSecs)
}

// ListPod return lis of pods in the given namespace
func ListPod(ns string) (*v1.PodList, error) {
	pods, err := gKubeInt.CoreV1().Pods(ns).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, errors.New("failed to list pods")
	}
	return pods, nil
}

func GetPod(podName string, namespace string) (*v1.Pod, error) {
	pods, err := gKubeInt.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("pod list failed, error: %v", err)
	}
	for _, pod := range pods.Items {
		if pod.Name == podName {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("pod %s not found", podName)
}

func GetPodIfPresent(podName string, namespace string) (*v1.Pod, bool, error) {
	pods, err := gKubeInt.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, false, fmt.Errorf("pod list failed, error: %v", err)
	}
	for _, pod := range pods.Items {
		if pod.Name == podName {
			return &pod, true, nil
		}
	}
	return nil, false, nil
}

func GetPodByUuid(uuid string) (*v1.Pod, bool, error) {
	pods, err := gKubeInt.CoreV1().Pods("").List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, false, fmt.Errorf("list failed, error: %v", err)
	}
	for _, pod := range pods.Items {
		if uuid == string(pod.ObjectMeta.UID) {
			return &pod, true, nil
		}
	}
	return nil, false, nil
}

func GetPodStatus(uuid string) (v1.PodPhase, bool, error) {
	pod, present, err := GetPodByUuid(uuid)
	if err != nil {
		return v1.PodUnknown, present, fmt.Errorf("get pod failed, error: %v", err)
	}
	if !present || pod == nil {
		return v1.PodUnknown, false, err
	}
	return pod.Status.Phase, present, err
}

func GetPodExistsByUuid(uuid string) (bool, error) {
	_, present, err := GetPodByUuid(uuid)
	if err != nil {
		return present, fmt.Errorf("get pod failed, error: %v", err)
	}
	return present, err
}

func GetPodExistsByName(name string, namespace string) (bool, error) {
	_, present, err := GetPodIfPresent(name, namespace)
	if err != nil {
		return present, fmt.Errorf("get pod failed, error: %v", err)
	}
	return present, err
}

func GetPodNameAndNamespaceFromUuid(uuid string) (string, string, error) {
	pod, present, err := GetPodByUuid(uuid)
	if err != nil {
		return "", "", fmt.Errorf("failed to get pod name from uuid, error: %v", err)
	}
	if !present || pod == nil {
		return "", "", fmt.Errorf("failed to find pod with uuid %s", uuid)
	}
	return pod.Name, pod.Namespace, nil
}

func GetPodsInNamespace(namespace string) (*v1.PodList, error) {
	pods, err := gKubeInt.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("pod list failed, error: %v", err)
	}
	return pods, nil
}

func WaitForPodReady(podName string, namespace string) (*v1.Pod, error) {
	// Wait for the Pod to transition to running
	const timoSecs = 120
	const timoSleepSecs = 10
	for ix := 0; ; ix++ {
		if ix >= timoSecs/timoSleepSecs {
			return nil, fmt.Errorf("timed out waiting for pod %s to be running", podName)
		}
		time.Sleep(timoSleepSecs * time.Second)
		pods, err := gKubeInt.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("list failed, error: %v", err)
		}
		for _, pod := range pods.Items {
			if podName == string(pod.Name) && pod.Status.Phase == v1.PodRunning {
				return &pod, nil
			}
		}
	}
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
	podApi := gKubeInt.CoreV1().Pods

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
			if len(newPodNames) >= len(podNames) {
				logf.Log.Info("All pods have been restarted.")
				return nil
			}
		}
	}
	logf.Log.Info("Restart pods failed", "oldpods", podNames, "newpods", newPodNames)
	return fmt.Errorf("restart failed incomplete error=%v", err)
}

// List mayastor pod names, conditionally
//  1) No timestamp - all mayastor pods
//	2) With timestamp - all mayastor pods created after the timestamp which are Running.
func listMayastorPods(timestamp *time.Time) ([]string, error) {
	var podNames []string
	podApi := gKubeInt.CoreV1().Pods
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
