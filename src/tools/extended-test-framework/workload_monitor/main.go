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

var gKubeInt kubernetes.Interface
var nameSpace = "default"

func readConfig() {
}

func banner() {
	fmt.Println("workload_monitor started")
}

func getPod(podName string, namespace string) (*v1.Pod, error) {
	pods, err := gKubeInt.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
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
	//testPlanRunSpec.IsActive = true

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
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("failed to get config")
		return
	}

	// read config file
	readConfig()

	gKubeInt = kubernetes.NewForConfigOrDie(restConfig)
	if restConfig == nil {
		fmt.Println("failed to get kubeint")
		return
	}

	time.Sleep(20 * time.Second)

	// find the test_director
	testDirectorPod, err := getPod("test-director", nameSpace)
	if err != nil {
		fmt.Println("failed to get test-director pod")
		return
	}

	fmt.Println("test-director pod IP is", testDirectorPod.Status.PodIP)
	testDirectorLoc := testDirectorPod.Status.PodIP + ":8080"

	transportConfig := client.DefaultTransportConfig().WithHost(testDirectorLoc)
	client := client.NewHTTPClientWithConfig(nil, transportConfig)

	var jk models.JiraKey = "MQ-004"
	sendTestPlan(client, "test name 4", &jk, false)

	getTestPlans(client)

	jk = "MQ-005"
	sendTestPlan(client, "test name 5", &jk, true)

	getTestPlans(client)

	time.Sleep(600 * time.Second)
}
