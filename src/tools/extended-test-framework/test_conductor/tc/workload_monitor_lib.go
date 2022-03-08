package tc

import (
	"fmt"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/common/wm/client"
	"mayastor-e2e/tools/extended-test-framework/common/wm/client/workload_monitor"

	"mayastor-e2e/tools/extended-test-framework/common/wm/models"

	"mayastor-e2e/tools/extended-test-framework/common"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"

	"github.com/go-openapi/strfmt"
	"k8s.io/apimachinery/pkg/types"
)

func AddWorkload(
	client *client.Etfw,
	name string,
	namespace string,
	violations []models.WorkloadViolationEnum) error {

	tcpod, err := k8sclient.GetPod("test-conductor", common.EtfwNamespace)
	if err != nil {
		return fmt.Errorf("failed to get test-conductor pod, error: %v\n", err)
	}

	pod, err := k8sclient.GetPod(name, namespace)
	if err != nil {
		return fmt.Errorf("failed to get pod %s, error: %v\n", name, err)
	}

	workload_spec := models.WorkloadSpec{}
	workload_spec.Violations = violations
	workload_params := workload_monitor.NewPutWorkloadByRegistrantParams()

	workload_params.Rid = strfmt.UUID(tcpod.ObjectMeta.UID)
	workload_params.Wid = strfmt.UUID(pod.ObjectMeta.UID)
	workload_params.Body = &workload_spec

	for i := 0; i < 5; i++ {
		var pPutWorkloadOk *workload_monitor.PutWorkloadByRegistrantOK
		pPutWorkloadOk, err = client.WorkloadMonitor.PutWorkloadByRegistrant(workload_params)
		if err != nil {
			logf.Log.Info("failed to put workload",
				"error", err.Error(),
				"name", name,
				"namespace", namespace)
		} else {
			logf.Log.Info("put workload",
				"name", string(pPutWorkloadOk.Payload.Name),
				"namespace", pPutWorkloadOk.Payload.Namespace,
				"wid", pPutWorkloadOk.Payload.ID)
			break
		}
		time.Sleep(10 * time.Second)
	}
	return err
}

func AddCommonWorkloads(
	client *client.Etfw,
	violations []models.WorkloadViolationEnum,
) error {
	return AddCommonWorkloadsWithExclusions(client, violations, []string{})
}

func AddCommonWorkloadsWithExclusions(
	client *client.Etfw,
	violations []models.WorkloadViolationEnum,
	exclusions []string) error {

	const namespace = "mayastor"

	tcpod, err := k8sclient.GetPod("test-conductor", common.EtfwNamespace)
	if err != nil {
		return fmt.Errorf("failed to get test-conductor pod, error: %v\n", err)
	}

	if err = AddWorkload(
		client,
		"test-conductor",
		common.EtfwNamespace,
		violations); err != nil {
		return fmt.Errorf("failed to inform workload monitor of test-conductor, error: %v", err)
	}

	err = AddNamespaceWorkloads(client, tcpod.ObjectMeta.UID, violations, namespace, exclusions)
	if err != nil {
		return fmt.Errorf("failed to add pods in namespace %s, error: %v\n", namespace, err)
	}
	return err
}

func AddNamespaceWorkloads(
	client *client.Etfw,
	requester_uuid types.UID,
	violations []models.WorkloadViolationEnum,
	namespace string,
	exclusion_labels []string) error {

	podlist, err := k8sclient.GetPodsInNamespace(namespace)
	if err != nil {
		return fmt.Errorf("failed to get pods in namespace %s, error: %v\n", namespace, err)
	}

	for _, pod := range podlist.Items {
		var skip bool
		for _, label := range exclusion_labels {
			value, ok := pod.ObjectMeta.Labels["app"]
			if ok && value == label {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		workload_spec := models.WorkloadSpec{}
		workload_spec.Violations = violations
		workload_params := workload_monitor.NewPutWorkloadByRegistrantParams()

		workload_params.Rid = strfmt.UUID(requester_uuid)
		workload_params.Wid = strfmt.UUID(pod.ObjectMeta.UID)
		workload_params.Body = &workload_spec
		for i := 0; i < 5; i++ {
			var pPutWorkloadOk *workload_monitor.PutWorkloadByRegistrantOK
			pPutWorkloadOk, err = client.WorkloadMonitor.PutWorkloadByRegistrant(workload_params)
			if err != nil {
				logf.Log.Info("failed to put workload",
					"error", err.Error(),
					"name", string(pPutWorkloadOk.Payload.Name),
					"namespace", pPutWorkloadOk.Payload.Namespace,
					"wid", pPutWorkloadOk.Payload.ID)
			} else {
				logf.Log.Info("put workload",
					"name", string(pPutWorkloadOk.Payload.Name),
					"namespace", pPutWorkloadOk.Payload.Namespace,
					"wid", pPutWorkloadOk.Payload.ID)
				break
			}
			time.Sleep(10 * time.Second)
		}
	}
	return err
}

func DeleteWorkload(client *client.Etfw, name string, namespace string) error {

	tcpod, err := k8sclient.GetPod("test-conductor", common.EtfwNamespace)
	if err != nil {
		return fmt.Errorf("failed to get tc pod %s, error: %v\n", name, err)
	}

	pod, err := k8sclient.GetPod(name, namespace)
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

func DeleteWorkloads(client *client.Etfw) error {

	tcpod, err := k8sclient.GetPod("test-conductor", common.EtfwNamespace)
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
