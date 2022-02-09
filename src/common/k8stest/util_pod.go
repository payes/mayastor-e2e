package k8stest

import (
	// container "github.com/openebs/maya/pkg/kubernetes/container/v1alpha1"
	// volume "github.com/openebs/maya/pkg/kubernetes/volume/v1alpha1"

	"context"
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
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
	// timeout and sleep time in seconds
	timeout       = 300 // timeout in seconds
	timeSleepSecs = 10  // sleep time in seconds
)

type Pod struct {
	object *corev1.Pod
}

// PodBuilder is the builder object for Pod
type PodBuilder struct {
	pod  *Pod
	errs []error
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

func (b *PodBuilder) WithVolumeDeviceOrMount(volType common.VolumeType) *PodBuilder {
	volMounts := coreV1.VolumeMount{
		Name:      "ms-volume",
		MountPath: common.FioFsMountPoint,
	}
	volDevices := coreV1.VolumeDevice{
		Name:       "ms-volume",
		DevicePath: common.FioBlockFilename,
	}
	if volType == common.VolRawBlock {
		b.WithVolumeDevice(volDevices)
	} else {
		b.WithVolumeMount(volMounts)
	}

	return b
}

// Build returns the Pod API instance
func (b *PodBuilder) Build() (*corev1.Pod, error) {
	if e2e_config.GetConfig().Platform.HostNetworkingRequired {
		b.pod.object.Spec.HostNetwork = true
	}
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.pod.object, nil
}

// ListPod return lis of pods in the given namespace
func ListPod(ns string) (*v1.PodList, error) {
	pods, err := gTestEnv.KubeInt.CoreV1().Pods(ns).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, errors.New("failed to list pods")
	}
	return pods, nil
}

func VerifyPodsOnNode(podLabelsList []string, nodeName string, namespace string) error {
	for _, label := range podLabelsList {
		var err error
		for ix := 0; ix < timeout/timeSleepSecs; ix++ {
			nodeList, err := GetNodeListForPods("app="+label, namespace)
			if err != nil {
				logf.Log.Info("VerifyPodsOnNode", "podLabel", label, "NodList", nodeList, "error", err)
			} else if len(nodeList) == 1 && nodeList[nodeName] == v1.PodRunning {
				break
			}
			time.Sleep(timeSleepSecs * time.Second)
		}
		if err != nil {
			return fmt.Errorf("failed to verify pod on node %s, podLabel: %s, error: %v", nodeName, label, err)
		}
	}
	return nil
}
