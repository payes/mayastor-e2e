package tc

import (
	"fmt"

	"github.com/go-openapi/strfmt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/common/k8s_lib"

	"mayastor-e2e/tools/extended-test-framework/common"
	td "mayastor-e2e/tools/extended-test-framework/common/td/client"
	wm "mayastor-e2e/tools/extended-test-framework/common/wm/client"
)

// TestConductor object for test conductor context
type TestConductor struct {
	TestDirectorClient    *td.Etfw
	WorkloadMonitorClient *wm.Etfw
	Config                ExtendedTestConfig
	TestRunId             strfmt.UUID
}

const SourceInstance = "test-conductor"

func NewTestConductor() (*TestConductor, error) {

	var testConductor TestConductor
	// read config file
	config, err := GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config")
	}
	logf.Log.Info("test", "name", config.TestName)
	testConductor.Config = config

	workloadMonitorPod, err := k8s_lib.WaitForPodReady("workload-monitor", common.EtfwNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get workload-monitor, error: %v", err)
	}
	logf.Log.Info("workload-monitor", "pod IP", workloadMonitorPod.Status.PodIP)
	workloadMonitorLoc := workloadMonitorPod.Status.PodIP + ":8080"

	// find the test_director
	testDirectorPod, err := k8s_lib.WaitForPodReady("test-director", common.EtfwNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get test-director, error: %v", err)
	}

	logf.Log.Info("test-director", "pod IP", testDirectorPod.Status.PodIP)
	testDirectorLoc := testDirectorPod.Status.PodIP + ":8080"

	transportConfigTd := td.DefaultTransportConfig().WithHost(testDirectorLoc)
	testConductor.TestDirectorClient = td.NewHTTPClientWithConfig(nil, transportConfigTd)

	transportConfigWm := wm.DefaultTransportConfig().WithHost(workloadMonitorLoc)
	testConductor.WorkloadMonitorClient = wm.NewHTTPClientWithConfig(nil, transportConfigWm)

	// the test run ID is the same as the uuid of the test conductor pod
	tcpod, err := k8s_lib.GetPod("test-conductor", common.EtfwNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get tc pod uid, error: %v", err)
	}

	testConductor.TestRunId = strfmt.UUID(tcpod.ObjectMeta.UID)

	if testConductor.Config.SendXrayTest == 1 || testConductor.Config.SendEvent == 1 {
		common.WaitTestDirector(testConductor.TestDirectorClient)
	}
	return &testConductor, nil
}
