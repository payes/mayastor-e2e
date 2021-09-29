package main

import (
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/wm"
)

func banner() {
	logf.Log.Info("workload_monitor v1 started")
}

func main() {
	banner()

	workloadMonitor, err := wm.NewWorkloadMonitor()
	if err != nil {
		logf.Log.Info("failed to create test monitor", "error", err)
		return
	}
	logger := zap.New(zap.UseDevMode(true))
	logf.SetLogger(logger)

	workloadMonitor.InstallSignalHandler()
	go workloadMonitor.StartServer()
	go workloadMonitor.StartMonitor()

	workloadMonitor.WaitSignal()
}
