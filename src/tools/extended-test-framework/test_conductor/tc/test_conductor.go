package tc

import (
	"fmt"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/common"
	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
	td "mayastor-e2e/tools/extended-test-framework/test_conductor/td/client"
	wm "mayastor-e2e/tools/extended-test-framework/test_conductor/wm/client"
)

// TestConductor object for test conductor context
type TestConductor struct {
	TestDirectorClient    *td.Etfw
	WorkloadMonitorClient *wm.Etfw
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

	workloadMonitorPod, err := k8sclient.WaitForPodReady("workload-monitor", common.EtfwNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get workload-monitor, error: %v", err)
	}
	logf.Log.Info("workload-monitor", "pod IP", workloadMonitorPod.Status.PodIP)
	workloadMonitorLoc := workloadMonitorPod.Status.PodIP + ":8080"

	// find the test_director
	testDirectorPod, err := k8sclient.WaitForPodReady("test-director", common.EtfwNamespace)
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
