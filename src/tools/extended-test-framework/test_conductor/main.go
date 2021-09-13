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
	"mayastor-e2e/lib"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	v1 "k8s.io/api/core/v1"
)

var gClientSet kubernetes.Clientset
var nameSpace = "default"

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

func installMayastor(clientset kubernetes.Clientset) {
	mayastorNodes, err := lib.GetMayastorNodeNames(clientset)
	if err == nil {
		for _, node := range mayastorNodes {
			fmt.Println("node ", node)
		}
	} else {
		fmt.Println(err)
		return
	}

	numMayastorInstances := len(mayastorNodes)
	//Expect(numMayastorInstances).ToNot(Equal(0))

	logf.Log.Info("Install", "# of mayastor instances", numMayastorInstances)
	/*
	   GenerateYamlFiles()
	   yamlsDir := locations.GetGeneratedYamlsDir()


	   err = k8stest.MkNamespace(common.NSMayastor())
	   Expect(err).ToNot(HaveOccurred())
	   k8stest.KubeCtlApplyYaml("moac-rbac.yaml", yamlsDir)

	   k8stest.KubeCtlApplyYaml("etcd", yamlsDir)
	   k8stest.KubeCtlApplyYaml("nats-deployment.yaml", yamlsDir)
	   k8stest.KubeCtlApplyYaml("csi-daemonset.yaml", yamlsDir)
	   k8stest.KubeCtlApplyYaml("moac-deployment.yaml", yamlsDir)
	   k8stest.KubeCtlApplyYaml("mayastor-daemonset.yaml", yamlsDir)

	   ready, err := k8stest.MayastorReady(2, 540)
	   Expect(err).ToNot(HaveOccurred())
	   Expect(ready).To(BeTrue())

	   crdReady := WaitForPoolCrd()
	   Expect(crdReady).To(BeTrue())

	   // Now create configured pools on all nodes.
	   k8stest.CreateConfiguredPools()

	   // Wait for pools to be online
	   const timoSecs = 120
	*/
}

func runTest(clientset kubernetes.Clientset) {
	installMayastor(clientset)
	time.Sleep(600 * time.Second)
}

func main() {
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
	client := client.NewHTTPClientWithConfig(nil, transportConfig)

	var jk models.JiraKey = "MQ-002"
	sendTestPlan(client, "test name 2", &jk, false)

	getTestPlans(client)

	jk = "MQ-003"
	sendTestPlan(client, "test name 3", &jk, true)

	getTestPlans(client)

	runTest(gClientSet)

}
