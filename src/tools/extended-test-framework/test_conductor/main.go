package main

import (
	"context"
	"fmt"

	errors "github.com/pkg/errors"

	"mayastor-e2e/tools/extended-test-framework/client"
	"mayastor-e2e/tools/extended-test-framework/client/test_director"

	"mayastor-e2e/tools/extended-test-framework/models"

	"mayastor-e2e/lib"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"time"

	v1 "k8s.io/api/core/v1"
)

var gClientSet kubernetes.Clientset

var nameSpace = "default"

// TestConductor object for test conductor context
type TestConductor struct {
	pTestDirectorClient *client.Extended
}

func banner() {
	fmt.Println("test_conductor started")
}

func getPod(podName string, namespace string) (*v1.Pod, error) {
	pods, err := gClientSet.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		fmt.Println("list failed")
		return nil, err
	}
	for _, pod := range pods.Items {
		if pod.Name == podName {
			fmt.Println("found pod ", podName)
			return &pod, nil
		}
	}
	fmt.Println("not found pod ", podName)
	return nil, errors.New("pod not found")
}

func getTestPlans(client *client.Extended) {
	testPlanParams := test_director.NewGetTestPlansParams()
	pTestPlansOk, err := client.TestDirector.GetTestPlans(testPlanParams)

	if err != nil {
		fmt.Printf("failed to get plans %v %v\n", err, pTestPlansOk)
	} else {
		fmt.Printf("got plans payload %v items %d\n", pTestPlansOk.Payload, len(pTestPlansOk.Payload))
		for _, tp := range pTestPlansOk.Payload {
			fmt.Printf("plan name %s\n", *tp.Name)
			fmt.Printf("plan key %v\n", tp.Key)
		}
	}
}

func sendTestPlan(client *client.Extended, name string, id *models.JiraKey, isActive bool) {

	testPlanRunSpec := models.TestRunSpec{}
	testPlanRunSpec.TestKey = id
	testPlanRunSpec.Data = "test"

	testPlanParams := test_director.NewPutTestPlanByIDParams()
	testPlanParams.ID = string(*id)
	testPlanParams.Body = &testPlanRunSpec

	pPutTestPlansOk, err := client.TestDirector.PutTestPlanByID(testPlanParams)

	if err != nil {
		fmt.Printf("failed to put plans %v %v\n", err, pPutTestPlansOk)
	} else {
		fmt.Printf("put plans payload %v\n", pPutTestPlansOk.Payload)
		fmt.Printf("plan name: %s ", *pPutTestPlansOk.Payload.Name)
		fmt.Printf("plan ID: %s\n", pPutTestPlansOk.Payload.Key)
	}
}

func installMayastor(clientset kubernetes.Clientset) error {
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
	if err = lib.CreatePools(clientset, GetConfig().PoolDevice); err != nil {
		return fmt.Errorf("cannot create mayastor pools %v", err)
	}
	return nil
}

func (testConductor TestConductor) runTest(clientset kubernetes.Clientset) error {
	if err := installMayastor(clientset); err != nil {
		return err
	}
	time.Sleep(600 * time.Second)
	return nil
}

func main() {

	testConductor := TestConductor{}

	banner()
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("failed to get cluster config")
		return
	}

	// read config file
	config := GetConfig()

	fmt.Printf("config name %s\n", config.ConfigName)
	if restConfig == nil {
		fmt.Println("failed to get kubeint")
		return
	}
	gClientSet = *kubernetes.NewForConfigOrDie(restConfig)

	time.Sleep(10 * time.Second)

	workloadMonitorPod, err := getPod("workload-monitor", "default")
	if err != nil {
		fmt.Println("failed to get workload-monitor")
		return
	}
	fmt.Println("worload-monitor pod IP is", workloadMonitorPod.Status.PodIP)

	// find the test_director
	testDirectorPod, err := getPod("test-director", nameSpace)
	if err != nil {
		fmt.Println("failed to get test-director pod")
		return
	}

	fmt.Println("test-director pod IP is", testDirectorPod.Status.PodIP)
	testDirectorLoc := testDirectorPod.Status.PodIP + ":8080"

	transportConfig := client.DefaultTransportConfig().WithHost(testDirectorLoc)
	testConductor.pTestDirectorClient = client.NewHTTPClientWithConfig(nil, transportConfig)

	var jk models.JiraKey = "MQ-002"
	sendTestPlan(testConductor.pTestDirectorClient, "test name 2", &jk, false)

	getTestPlans(testConductor.pTestDirectorClient)

	jk = "MQ-003"
	sendTestPlan(testConductor.pTestDirectorClient, "test name 3", &jk, true)

	getTestPlans(testConductor.pTestDirectorClient)

	if err = testConductor.runTest(gClientSet); err != nil {
		fmt.Println("run test failed")
	}

}
