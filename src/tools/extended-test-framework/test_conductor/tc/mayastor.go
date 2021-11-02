package tc

import (
	"fmt"

	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
)

func InstallMayastor(pool_device string) error {
	var err error
	if err = k8sclient.CreateNamespace("mayastor"); err != nil {
		return fmt.Errorf("cannot create namespace %v", err)
	}
	if err = k8sclient.DeployYaml("moac-rbac.yaml"); err != nil {
		return fmt.Errorf("cannot create moac-rbac %v", err)
	}
	if err = k8sclient.DeployYaml("etcd/statefulset.yaml"); err != nil {
		return fmt.Errorf("cannot create etcd stateful set %v", err)
	}
	if err = k8sclient.DeployYaml("etcd/svc-headless.yaml"); err != nil {
		return fmt.Errorf("cannot create etcd svc-headless %v", err)
	}
	if err = k8sclient.DeployYaml("etcd/svc.yaml"); err != nil {
		return fmt.Errorf("cannot create etcd svc %v", err)
	}
	if err = k8sclient.DeployYaml("nats-deployment.yaml"); err != nil {
		return fmt.Errorf("cannot create nats-deployment %v", err)
	}
	if err = k8sclient.DeployYaml("csi-daemonset.yaml"); err != nil {
		return fmt.Errorf("cannot create csi daemonset %v", err)
	}
	if err = k8sclient.DeployYaml("moac-deployment.yaml"); err != nil {
		return fmt.Errorf("cannot create moac deployment %v", err)
	}
	if err = k8sclient.DeployYaml("mayastor-daemonset.yaml"); err != nil {
		return fmt.Errorf("cannot create mayastor daemonset %v", err)
	}
	if err = k8sclient.CreatePools(pool_device); err != nil {
		return fmt.Errorf("cannot create mayastor pools %v", err)
	}
	return nil
}
