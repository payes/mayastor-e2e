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

	if err = tc.SendTestPlanRunning(testConductor.TestDirectorClient, "test plan", testConductor.Config.TestPlan); err != nil {
		logf.Log.Info("failed to send test plan", "error", err)
		os.Exit(1)
	}

	if err = tc.GetTestPlans(testConductor.TestDirectorClient); err != nil {
		logf.Log.Info("failed to get test plan", "error", err)
		os.Exit(1)
	}

	if err = tests.ReplicaPerturbationTest(testConductor); err != nil {
		logf.Log.Info("replica perturbation test failed", "error", err)
		if err = tests.SendTestCompletedFail(testConductor, err.Error()); err != nil {
			logf.Log.Info("replica perturbation test failed to report error", "error", err)
		}
		os.Exit(1)
	}
}
