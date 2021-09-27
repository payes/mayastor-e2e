package tc

import (
	"context"
	"fmt"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/common"
	td "mayastor-e2e/tools/extended-test-framework/test_conductor/td/client"
	wm "mayastor-e2e/tools/extended-test-framework/test_conductor/wm/client"

	"mayastor-e2e/lib"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"time"

	v1 "k8s.io/api/core/v1"
)

// TestConductor object for test conductor context
type TestConductor struct {
	TestDirectorClient    *td.Etfw
	WorkloadMonitorClient *wm.Etfw
	Clientset             kubernetes.Clientset
	Config                ExtendedTestConfig
}

const SourceInstance = "test-conductor"

func NewTestConductor() (*TestConductor, error) {

	var testConductor TestConductor
	// read config file
	config, err := GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config")
	}
	logf.Log.Info("config", "name", config.ConfigName)
	testConductor.Config = config

	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config")
	}
	if restConfig == nil {
		return nil, fmt.Errorf("rest config is nil")
	}
	testConductor.Clientset = *kubernetes.NewForConfigOrDie(restConfig)

	workloadMonitorPod, err := waitForPod(testConductor.Clientset, "workload-monitor", common.EtfwNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get workload-monitor, error: %v", err)
	}
	logf.Log.Info("workload-monitor", "pod IP", workloadMonitorPod.Status.PodIP)
	workloadMonitorLoc := workloadMonitorPod.Status.PodIP + ":8080"

	// find the test_director
	testDirectorPod, err := waitForPod(testConductor.Clientset, "test-director", common.EtfwNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get test-director, error: %v", err)
	}

	logf.Log.Info("test-director", "pod IP", testDirectorPod.Status.PodIP)
	testDirectorLoc := testDirectorPod.Status.PodIP + ":8080"

	transportConfigTd := td.DefaultTransportConfig().WithHost(testDirectorLoc)
	testConductor.TestDirectorClient = td.NewHTTPClientWithConfig(nil, transportConfigTd)

	transportConfigWm := wm.DefaultTransportConfig().WithHost(workloadMonitorLoc)
	testConductor.WorkloadMonitorClient = wm.NewHTTPClientWithConfig(nil, transportConfigWm)

	return &testConductor, nil
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

func getPodsInNamespace(clientset kubernetes.Clientset, namespace string) (*v1.PodList, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("pod list failed, error: %v", err)
	}
	return pods, nil
}
