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

	testRunId, failmessage, err := tests.SteadyStateTest(testConductor)
	if err != nil {
		logf.Log.Info("failed to run test", "error", err)
		if failmessage == "" {
			failmessage = err.Error()
		}
	}
	err = tests.ReportResult(testConductor, failmessage, testRunId)
	if err != nil {
		logf.Log.Info("failed to report result", "error", err)
		os.Exit(1)
	}

}
