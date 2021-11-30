package k8s_lib

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var gKubeInt kubernetes.Interface

func init() {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("failed to get in-cluster config, attempting out-of-cluster")
		restConfig = config.GetConfigOrDie()
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
