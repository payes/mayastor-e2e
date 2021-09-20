package util

import (
	"context"
	"fmt"
	"time"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	v1 "k8s.io/api/core/v1"
)

var gKubeInt kubernetes.Interface

func init() {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("failed to get config")
		return
	}

	if restConfig == nil {
		fmt.Println("failed to get restConfig")
		return
	}

	gKubeInt = kubernetes.NewForConfigOrDie(restConfig)
	if restConfig == nil {
		fmt.Println("failed to get kubeint")
		return
	}
}

func GetPodByUuid(uuid string) (*v1.Pod, bool, error) {
	pods, err := gKubeInt.CoreV1().Pods("").List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, false, fmt.Errorf("list failed, error: %v", err)
	}
	for _, pod := range pods.Items {
		if uuid == string(pod.ObjectMeta.UID) {
			fmt.Println("found pod ", pod.Name)
			return &pod, true, nil
		}
	}
	fmt.Println("not found pod ", uuid)
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

func GetPodExists(uuid string) (bool, error) {
	_, present, err := GetPodByUuid(uuid)
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
				fmt.Println("found pod ", pod.Name)
				return &pod, nil
			}
		}
	}
}
