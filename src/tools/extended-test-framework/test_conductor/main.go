package main

import (
	"context"
	"fmt"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/client"

	"mayastor-e2e/lib"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"time"

	v1 "k8s.io/api/core/v1"
)

var nameSpace = "default"

// TestConductor object for test conductor context
type TestConductor struct {
	pTestDirectorClient    *client.Extended
	pWorkloadMonitorClient *client.Extended
	clientset              kubernetes.Clientset
	config                 ExtendedTestConfig
}

func banner() {
	logf.Log.Info("test_conductor started")
}

func waitForPod(clientset kubernetes.Clientset, podName string, namespace string) (*v1.Pod, error) {
	// Wait for the fio Pod to transition to running
	const timoSecs = 120
	const timoSleepSecs = 10
	for ix := 0; ; ix++ {
		if lib.IsPodRunning(clientset, podName, namespace) {
			break
		}
		if ix >= timoSecs/timoSleepSecs {
			return nil, fmt.Errorf("timed out waiting for pod %s to be running", podName)
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	return getPod(clientset, podName, namespace)
}

func getPod(clientset kubernetes.Clientset, podName string, namespace string) (*v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("pod list failed, error: %v", err)
	}
	for _, pod := range pods.Items {
		if pod.Name == podName {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("pod %s not found", podName)
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
	config, err := GetConfig()
	if err != nil {
		fmt.Println("failed to get config")
		return
	}
	testConductor.config = config

	logf.Log.Info("config", "name", config.ConfigName)
	if restConfig == nil {
		logf.Log.Info("failed to get kubeint")
		return
	}
	testConductor.clientset = *kubernetes.NewForConfigOrDie(restConfig)

	time.Sleep(10 * time.Second)

	workloadMonitorPod, err := waitForPod(testConductor.clientset, "workload-monitor", "default")
	if err != nil {
		logf.Log.Info("failed to get workload-monitor")
		return
	}
	logf.Log.Info("worload-monitor", "pod IP", workloadMonitorPod.Status.PodIP)
	workloadMonitorLoc := workloadMonitorPod.Status.PodIP + ":8080"

	// find the test_director
	testDirectorPod, err := waitForPod(testConductor.clientset, "test-director", nameSpace)
	if err != nil {
		logf.Log.Info("failed to get test-director pod")
		return
	}

	logf.Log.Info("test-director", "pod IP", workloadMonitorPod.Status.PodIP)
	testDirectorLoc := testDirectorPod.Status.PodIP + ":8080"

	transportConfig := client.DefaultTransportConfig().WithHost(testDirectorLoc)
	testConductor.pTestDirectorClient = client.NewHTTPClientWithConfig(nil, transportConfig)

	transportConfig = client.DefaultTransportConfig().WithHost(workloadMonitorLoc)
	testConductor.pWorkloadMonitorClient = client.NewHTTPClientWithConfig(nil, transportConfig)

	if err = sendTestPlan(testConductor.pTestDirectorClient, "test name 2", "MQ-002", false); err != nil {
		logf.Log.Info("failed to send test plan", "error", err)
		return
	}

	if err = getTestPlans(testConductor.pTestDirectorClient); err != nil {
		logf.Log.Info("failed to get test plan", "error", err)
		return
	}

	if err = sendTestPlan(testConductor.pTestDirectorClient, "test name 3", "MQ-003", true); err != nil {
		logf.Log.Info("failed to send test plan", "error", err)
		return
	}

	if err = getTestPlans(testConductor.pTestDirectorClient); err != nil {
		logf.Log.Info("failed to get test plan", "error", err)
		return
	}

	if err = testConductor.BasicSoakTest(); err != nil {
		logf.Log.Info("Basic soak test failed", "error", err)
	}
}
