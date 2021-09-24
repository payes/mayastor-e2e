package tc

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/test_conductor/wm/client"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/wm/client/workload_monitor"

	"mayastor-e2e/tools/extended-test-framework/test_conductor/wm/models"

	"github.com/go-openapi/strfmt"
	"mayastor-e2e/tools/extended-test-framework/common"
)

func AddWorkload(clientset kubernetes.Clientset, client *client.Etfw, name string, namespace string) error {

	tcpod, err := getPod(clientset, "test-conductor", common.EtfwNamespace)
	if err != nil {
		return fmt.Errorf("failed to get tc pod %s, error: %v\n", name, err)
	}

	pod, err := getPod(clientset, name, namespace)
	if err != nil {
		return fmt.Errorf("failed to get pod %s, error: %v\n", name, err)
	}

	workload_spec := models.WorkloadSpec{}
	workload_spec.Violations = []models.WorkloadViolationEnum{models.WorkloadViolationEnumRESTARTED}
	workload_params := workload_monitor.NewPutWorkloadByRegistrantParams()

	workload_params.Rid = strfmt.UUID(tcpod.ObjectMeta.UID)
	workload_params.Wid = strfmt.UUID(pod.ObjectMeta.UID)
	workload_params.Body = &workload_spec
	pPutWorkloadOk, err := client.WorkloadMonitor.PutWorkloadByRegistrant(workload_params)

	if err != nil {
		return fmt.Errorf("failed to put workload %v %v\n", err, pPutWorkloadOk)
	} else {
		logf.Log.Info("put workload",
			"name", string(pPutWorkloadOk.Payload.Name),
			"namespace", pPutWorkloadOk.Payload.Namespace,
			"wid", pPutWorkloadOk.Payload.ID)
	}
	return nil
}

func AddWorkloadsInNamespace(clientset kubernetes.Clientset, client *client.Etfw, namespace string) error {
	tcpod, err := getPod(clientset, "test-conductor", common.EtfwNamespace)
	if err != nil {
		return fmt.Errorf("failed to get tc pod, error: %v\n", err)
	}

	podlist, err := getPodsInNamespace(clientset, namespace)
	if err != nil {
		return fmt.Errorf("failed to get pods in namespace %s, error: %v\n", namespace, err)
	}

	for _, pod := range podlist.Items {
		workload_spec := models.WorkloadSpec{}
		workload_spec.Violations = []models.WorkloadViolationEnum{models.WorkloadViolationEnumRESTARTED}
		workload_params := workload_monitor.NewPutWorkloadByRegistrantParams()

		workload_params.Rid = strfmt.UUID(tcpod.ObjectMeta.UID)
		workload_params.Wid = strfmt.UUID(pod.ObjectMeta.UID)
		workload_params.Body = &workload_spec
		pPutWorkloadOk, err := client.WorkloadMonitor.PutWorkloadByRegistrant(workload_params)

		if err != nil {
			return fmt.Errorf("failed to put workload %v %v\n", err, pPutWorkloadOk)
		} else {
			logf.Log.Info("put workload",
				"name", string(pPutWorkloadOk.Payload.Name),
				"namespace", pPutWorkloadOk.Payload.Namespace,
				"wid", pPutWorkloadOk.Payload.ID)
		}
	}
	return nil
}

func DeleteWorkload(clientset kubernetes.Clientset, client *client.Etfw, name string, namespace string) error {

	tcpod, err := getPod(clientset, "test-conductor", common.EtfwNamespace)
	if err != nil {
		return fmt.Errorf("failed to get tc pod %s, error: %v\n", name, err)
	}

	pod, err := getPod(clientset, name, namespace)
	if err != nil {
		return fmt.Errorf("failed to get pod %s, error: %v\n", name, err)
	}

	workload_params := workload_monitor.NewDeleteWorkloadByRegistrantParams()

	workload_params.Rid = strfmt.UUID(tcpod.ObjectMeta.UID)
	workload_params.Wid = strfmt.UUID(pod.ObjectMeta.UID)
	pDeleteWorkloadOk, err := client.WorkloadMonitor.DeleteWorkloadByRegistrant(workload_params)

	if err != nil {
		return fmt.Errorf("failed to delete workload %v %v\n", err, pDeleteWorkloadOk)
	} else {
		logf.Log.Info("deleted workload",
			"name", string(pDeleteWorkloadOk.Payload.Name),
			"namespace", pDeleteWorkloadOk.Payload.Namespace,
			"wid", pDeleteWorkloadOk.Payload.ID)
	}
	return nil
}

func DeleteWorkloads(clientset kubernetes.Clientset, client *client.Etfw) error {

	tcpod, err := getPod(clientset, "test-conductor", common.EtfwNamespace)
	if err != nil {
		return fmt.Errorf("failed to get tc pod, error: %v", err)
	}

	workload_params := workload_monitor.NewDeleteWorkloadsByRegistrantParams()

	workload_params.Rid = strfmt.UUID(tcpod.ObjectMeta.UID)
	pDeleteWorkloadsOk, err := client.WorkloadMonitor.DeleteWorkloadsByRegistrant(workload_params)

	if err != nil {
		return fmt.Errorf("failed to delete workloads %v %v\n", err, pDeleteWorkloadsOk)
	} else {
		logf.Log.Info("deleted workloads",
			"details", string(pDeleteWorkloadsOk.Payload.Details),
			"items", pDeleteWorkloadsOk.Payload.ItemsAffected,
			"result", pDeleteWorkloadsOk.Payload.Result)
	}
	return nil
}
