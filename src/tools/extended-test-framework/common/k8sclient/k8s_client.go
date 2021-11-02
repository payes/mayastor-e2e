package k8sclient

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
