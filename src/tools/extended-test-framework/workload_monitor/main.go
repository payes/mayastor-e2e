package main

import (
	"fmt"
	"log"

	"os"
	"time"

	"github.com/go-openapi/loads"
	flags "github.com/jessevdk/go-flags"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/models"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/restapi"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/restapi/operations"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/util"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/client"

	"mayastor-e2e/tools/extended-test-framework/common"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type TestMonitor struct {
	pTestDirectorClient *client.Etfw
	clientset           kubernetes.Clientset
}

func banner() {
	logf.Log.Info("workload_monitor v1 started")
}

/*
	// PodPending means the pod has been accepted by the system, but one or more of the containers
	// has not been started. This includes time before being bound to a node, as well as time spent
	// pulling images onto the host.
	PodPending PodPhase = "Pending"
	// PodRunning means the pod has been bound to a node and all of the containers have been started.
	// At least one container is still running or is in the process of being restarted.
	PodRunning PodPhase = "Running"
	// PodSucceeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	PodSucceeded PodPhase = "Succeeded"
	// PodFailed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	PodFailed PodPhase = "Failed"
	// PodUnknown means that for some reason the state of the pod could not be obtained, typically due
	// to an error in communicating with the host of the pod.
	// Deprecated: It isn't being set since 2015 (74da3b14b0c0f658b3bb8d2def5094686d0e9095)
	PodUnknown PodPhase = "Unknown"
*/

func NewTestMonitor() (*TestMonitor, error) {
	var testMonitor TestMonitor
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config")
	}
	if restConfig == nil {
		return nil, fmt.Errorf("rest config is nil")
	}
	testMonitor.clientset = *kubernetes.NewForConfigOrDie(restConfig)

	// find the test_director
	testDirectorPod, err := util.WaitForPodReady("test-director", common.EtfwNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get test-director, error: %v", err)
	}

	logf.Log.Info("test-director", "pod IP", testDirectorPod.Status.PodIP)
	testDirectorLoc := testDirectorPod.Status.PodIP + ":8080"

	transportConfig := client.DefaultTransportConfig().WithHost(testDirectorLoc)
	testMonitor.pTestDirectorClient = client.NewHTTPClientWithConfig(nil, transportConfig)

	return &testMonitor, nil
}

func startMonitor(testMonitor *TestMonitor) {
	for {
		time.Sleep(10 * time.Second)
		list := util.GetWorkloadList()

		for _, wl := range list {
			for _, spec := range wl.WorkloadSpec.Violations {
				switch spec {
				case models.WorkloadViolationEnumRESTARTED:
					pod, present, err := util.GetPodByUuid(string(wl.ID))
					if err != nil {
						fmt.Printf("failed to get pod %s\n", wl.Name)
					}
					if present {
						containerStatuses := pod.Status.ContainerStatuses
						restartcount := int32(0)
						for _, containerStatus := range containerStatuses {
							if containerStatus.RestartCount != 0 {
								if containerStatus.RestartCount > restartcount {
									restartcount = containerStatus.RestartCount
								}
								logf.Log.Info(pod.Name, "restarts", containerStatus.RestartCount)
								break
							}
						}
						if restartcount != 0 {
							fmt.Printf("pod %s restarted\n", wl.Name)
							if err := sendEvent(testMonitor.pTestDirectorClient, "pod restarted", string(wl.Name)); err != nil {
								logf.Log.Info("failed to send", "error", err)
							} else {
								util.DeleteWorkloadById(wl.ID)
							}
						}
					}
				case models.WorkloadViolationEnumTERMINATED:
					podstatus, present, err := util.GetPodStatus(string(wl.ID))
					if err != nil {
						fmt.Printf("failed to get pod status %s\n", wl.Name)
					}
					if present {
						fmt.Printf("pod status %v\n", podstatus)
						fmt.Printf(" checking pod %s for terminated\n", wl.Name)
						if podstatus == v1.PodFailed {
							fmt.Printf("pod %s failed\n", wl.Name)
							if err := sendEvent(testMonitor.pTestDirectorClient, "pod terminated", string(wl.Name)); err != nil {
								logf.Log.Info("failed to send", "error", err)
							} else {
								util.DeleteWorkloadById(wl.ID)
							}
						}
					}
				case models.WorkloadViolationEnumNOTPRESENT:
					present, err := util.GetPodExists(string(wl.ID))
					if err != nil {
						fmt.Printf("failed to get pod status %s\n", wl.Name)
					}
					fmt.Printf(" checking pod %s for not present\n", wl.Name)
					if !present {
						fmt.Printf("pod %s does not exist\n", wl.Name)
						if err := sendEvent(testMonitor.pTestDirectorClient, "pod absent", string(wl.Name)); err != nil {
							logf.Log.Info("failed to send", "error", err)
						} else {
							util.DeleteWorkloadById(wl.ID)
						}
					}
				}
			}
		}
	}
}

func startServer() {
	logf.Log.Info("tm server started")

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewEtfwAPI(swaggerSpec)
	server := restapi.NewServer(api)
	//defer server.Shutdown()

	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = "Test Framework API"
	parser.LongDescription = "MayaData System Test Framework API"

	server.ConfigureFlags()
	for _, optsGroup := range api.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}
	logf.Log.Info("workload_monitor about to configure")
	server.ConfigureAPI()

	logf.Log.Info("workload_monitor about to serve")
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}

func main() {
	banner()

	//util.InitGoClient()
	//util.InitWorkloadList()

	testMonitor, err := NewTestMonitor()
	if err != nil {
		logf.Log.Info("failed to create test monitor", "error", err)
		return
	}
	_ = testMonitor
	logger := zap.New(zap.UseDevMode(true))
	logf.SetLogger(logger)

	go startServer()
	go startMonitor(testMonitor)

	logf.Log.Info("waiting")
	time.Sleep(6000 * time.Second)
	logf.Log.Info("finishing")
}
