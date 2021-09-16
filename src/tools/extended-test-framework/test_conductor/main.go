package main

import (
	"context"
	"fmt"

	errors "github.com/pkg/errors"

	"mayastor-e2e/tools/extended-test-framework/client"
	"mayastor-e2e/tools/extended-test-framework/client/test_director"

	"mayastor-e2e/tools/extended-test-framework/models"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"time"

	v1 "k8s.io/api/core/v1"
)

var nameSpace = "default"

// TestConductor object for test conductor context
type TestConductor struct {
	pTestDirectorClient *client.Extended
	clientset           kubernetes.Clientset
}

func banner() {
	fmt.Println("test_conductor started")
}

func getPod(clientset kubernetes.Clientset, podName string, namespace string) (*v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
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

func main() {
	banner()

	testConductor := TestConductor{}

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
	testConductor.clientset = *kubernetes.NewForConfigOrDie(restConfig)

	time.Sleep(10 * time.Second)

	workloadMonitorPod, err := getPod(testConductor.clientset, "workload-monitor", "default")
	if err != nil {
		fmt.Println("failed to get workload-monitor")
		return
	}
	fmt.Println("worload-monitor pod IP is", workloadMonitorPod.Status.PodIP)

	// find the test_director
	testDirectorPod, err := getPod(testConductor.clientset, "test-director", nameSpace)
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

	if err = testConductor.BasicSoakTest(); err != nil {
		fmt.Printf("run test failed, error: %v\n", err)
	}

}
