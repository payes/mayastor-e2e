package main

import (
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/wm"
)

func main() {
	logger := zap.New(zap.UseDevMode(true))
	logf.SetLogger(logger)

	logf.Log.Info("workload_monitor v1 started")

	workloadMonitor, err := wm.NewWorkloadMonitor()
	if err != nil {
		logf.Log.Info("failed to create test monitor", "error", err)
		return
	}

	workloadMonitor.InstallSignalHandler()
	go workloadMonitor.StartServer()
	go workloadMonitor.StartMonitor()

	workloadMonitor.WaitSignal()
}
