package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"log_monitor/utils"
	"os"
)

func Init() (kubernetes.Interface, error) {
	rc, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("failed to get config")
		return nil, err
	}

	if rc == nil {
		fmt.Println("failed to get restConfig")
		return nil, err
	}

	gKubeInt := kubernetes.NewForConfigOrDie(rc)
	if rc == nil {
		fmt.Println("failed to get kubeint")
		return nil, err
	}
	return gKubeInt, nil
}

func GetPod(podName, ns string) (v1.Pod, error) {
	pod, err := app.Client.CoreV1().Pods(ns).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return *pod, err
	}
	return *pod, nil
}

func ListPods(ns string) ([]v1.Pod, error) {
	pods, err := app.Client.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func PodStatus(pod *v1.Pod) v1.PodPhase {
	return pod.Status.Phase
}

func execTailPodCommand(pod v1.Pod) {
	cmd := []string{"/bin/sh", "-c", "tail -F /tmp/file-fluentd.log/*.log"}
	req := app.Client.CoreV1().RESTClient().
		Post().
		Namespace(utils.FluentdNS).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: pod.Spec.Containers[0].Name,
			Command:   cmd,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("failed to get config err %v", err)
		return
	}
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		log.Fatalf("failed to execute command %s with err %v", cmd, err)
		return
	}
	go func() {
		fmt.Println("Reading from pod:", pod.Name)
		defer app.PipeReader.Close()
		defer app.PipeWriter.Close()
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: app.PipeWriter,
			Stderr: app.PipeWriter,
			Tty:    true,
		})
		if err != nil {
			log.Println(err, "error from command stream")
		}
	}()
}
