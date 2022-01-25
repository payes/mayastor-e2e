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

	test_err := tests.SteadyStateTest(testConductor)

	tests.SendTestRunFinished(testConductor, test_err)
}
