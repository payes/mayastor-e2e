package main

import (
	"os"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/tests"
)

func main() {
	logger := zap.New(zap.UseDevMode(true))
	logf.SetLogger(logger)

	testConductor, err := tc.NewTestConductor()
	if err != nil {
		logf.Log.Info("failed to create test conductor", "error", err)
		os.Exit(1)
	}

	if err = tests.NonSteadyStateTest(testConductor); err != nil {
		logf.Log.Info("non steady state test failed", "error", err)
		if err = tests.SendTestCompletedFail(testConductor, err.Error()); err != nil {
			logf.Log.Info("non steady state test failed to report error", "error", err)
		}
		os.Exit(1)
	}
}
