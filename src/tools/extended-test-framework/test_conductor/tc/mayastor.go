package tc

import (
	"fmt"

	"mayastor-e2e/lib"

	"k8s.io/client-go/kubernetes"
)

func InstallMayastor(clientset kubernetes.Clientset, pool_device string) error {
	var err error
	if err = lib.CreateNamespace(clientset, "mayastor"); err != nil {
		return fmt.Errorf("cannot create namespace %v", err)
	}
	if err = lib.DeployYaml(clientset, "moac-rbac.yaml"); err != nil {
		return fmt.Errorf("cannot create moac-rbac %v", err)
	}
	if err = lib.DeployYaml(clientset, "etcd/statefulset.yaml"); err != nil {
		return fmt.Errorf("cannot create etcd stateful set %v", err)
	}
	if err = lib.DeployYaml(clientset, "etcd/svc-headless.yaml"); err != nil {
		return fmt.Errorf("cannot create etcd svc-headless %v", err)
	}
	if err = lib.DeployYaml(clientset, "etcd/svc.yaml"); err != nil {
		return fmt.Errorf("cannot create etcd svc %v", err)
	}
	if err = lib.DeployYaml(clientset, "nats-deployment.yaml"); err != nil {
		return fmt.Errorf("cannot create nats-deployment %v", err)
	}
	if err = lib.DeployYaml(clientset, "csi-daemonset.yaml"); err != nil {
		return fmt.Errorf("cannot create csi daemonset %v", err)
	}
	if err = lib.DeployYaml(clientset, "moac-deployment.yaml"); err != nil {
		return fmt.Errorf("cannot create moac deployment %v", err)
	}
	if err = lib.DeployYaml(clientset, "mayastor-daemonset.yaml"); err != nil {
		return fmt.Errorf("cannot create mayastor daemonset %v", err)
	}
	if err = lib.CreatePools(clientset, pool_device); err != nil {
		return fmt.Errorf("cannot create mayastor pools %v", err)
	}
	return nil
}
