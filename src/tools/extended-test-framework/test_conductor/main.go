package main

import (
	"os"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/tests"
)

func banner() {
	logf.Log.Info("test_conductor started")
}

func main() {
	logger := zap.New(zap.UseDevMode(true))
	logf.SetLogger(logger)
	banner()

	testConductor, err := tc.NewTestConductor()
	if err != nil {
		logf.Log.Info("failed to create test conductor", "error", err)
		os.Exit(1)
	}

	if err = tc.SendTestPlan(testConductor.TestDirectorClient, "test name 2", testConductor.Config.TestPlan); err != nil {
		logf.Log.Info("failed to send test plan", "error", err)
		os.Exit(1)
	}

	if err = tc.GetTestPlans(testConductor.TestDirectorClient); err != nil {
		logf.Log.Info("failed to get test plan", "error", err)
		os.Exit(1)
	}

	switch {
	case testConductor.Config.Test == "steady_state":
		if err = tests.SteadyStateTest(testConductor); err != nil {
			logf.Log.Info("steady state failed", "error", err)
		}
	}
}
