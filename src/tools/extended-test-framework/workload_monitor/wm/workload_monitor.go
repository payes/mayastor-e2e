package wm

import (
	"fmt"
	"log"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-openapi/loads"
	flags "github.com/jessevdk/go-flags"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/k8sclient"
	wlist "mayastor-e2e/tools/extended-test-framework/workload_monitor/list"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/models"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/restapi"
	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/restapi/operations"

	"mayastor-e2e/tools/extended-test-framework/workload_monitor/swagger/client"

	"mayastor-e2e/tools/extended-test-framework/common"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

type WorkloadMonitor struct {
	pTestDirectorClient *client.Etfw
	channel             chan int
}

const SourceInstance = "workload-monitor"

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

func (workloadMonitor *WorkloadMonitor) InstallSignalHandler() {
	signal_channel := make(chan os.Signal, 1)
	signal.Notify(signal_channel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	workloadMonitor.channel = make(chan int)
	go func() {
		for {
			s := <-signal_channel
			switch s {
			case syscall.SIGTERM:
				workloadMonitor.channel <- 0

			default:
				workloadMonitor.channel <- 1
			}
		}
	}()
}

func (workloadMonitor *WorkloadMonitor) WaitSignal() {
	exitCode := <-workloadMonitor.channel
	if exitCode != 0 { // abnormal termination
		if err := SendEvent(workloadMonitor.pTestDirectorClient, "workload monitor terminated", SourceInstance, SourceInstance); err != nil {
			logf.Log.Info("failed to send", "error", err)
		}
	}
}

func NewWorkloadMonitor() (*WorkloadMonitor, error) {
	var workloadMonitor WorkloadMonitor
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster config")
	}
	if restConfig == nil {
		return nil, fmt.Errorf("rest config is nil")
	}

	// find the test_director
	testDirectorPod, err := k8sclient.WaitForPodReady("test-director", common.EtfwNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get test-director, error: %v", err)
	}

	logf.Log.Info("test-director", "pod IP", testDirectorPod.Status.PodIP)
	testDirectorLoc := testDirectorPod.Status.PodIP + ":8080"

	transportConfig := client.DefaultTransportConfig().WithHost(testDirectorLoc)
	workloadMonitor.pTestDirectorClient = client.NewHTTPClientWithConfig(nil, transportConfig)

	return &workloadMonitor, nil
}

func (workloadMonitor *WorkloadMonitor) StartMonitor() {
	logf.Log.Info("workload monitor polling")
	for {
		time.Sleep(10 * time.Second)

		// Lock the list while iterating.
		// This ensures that the tests reflect exactly what was most
		// recently specified by the test conductor. This avoids a
		// potential race issue if the test conductor removes a pod from
		// the list, removes the corresponding pod and the workload monitor
		// sees it as missing because it has old data.
		wlist.Lock()
		list := wlist.GetWorkloadList()

		for _, wl := range list {
			for _, spec := range wl.WorkloadSpec.Violations {
				switch spec {
				case models.WorkloadViolationEnumRESTARTED:
					pod, present, err := k8sclient.GetPodByUuid(string(wl.ID))
					if err != nil {
						logf.Log.Info("failed to get pod by UUID", "pod", wl.ID)
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
							message := fmt.Sprintf("pod %s restarted %d times", wl.Name, restartcount)
							logf.Log.Info("restart", "message", message)
							if err := SendEvent(workloadMonitor.pTestDirectorClient, message, string(wl.Name), SourceInstance); err != nil {
								logf.Log.Info("failed to send", "error", err)
							} else {
								wlist.DeleteWorkloadById(wl.ID)
							}
						}
					}
				case models.WorkloadViolationEnumTERMINATED:
					podstatus, present, err := k8sclient.GetPodStatus(string(wl.ID))
					if err != nil {
						logf.Log.Info("failed to get pod status", "error", err)
					}
					if present {
						if podstatus == v1.PodFailed {
							message := fmt.Sprintf("pod %s terminated", wl.Name)
							logf.Log.Info("termination", "message", message)
							if err := SendEvent(workloadMonitor.pTestDirectorClient, message, string(wl.Name), SourceInstance); err != nil {
								logf.Log.Info("failed to send", "error", err)
							} else {
								wlist.DeleteWorkloadById(wl.ID)
							}
						}
					}
				case models.WorkloadViolationEnumNOTPRESENT:
					present, err := k8sclient.GetPodExists(string(wl.ID))
					if err != nil {
						fmt.Printf("failed to get pod status %s\n", wl.Name)
					}
					if !present {
						message := fmt.Sprintf("pod %s absent", wl.Name)
						logf.Log.Info("absent", "message", message)

						if err := SendEvent(workloadMonitor.pTestDirectorClient, message, string(wl.Name), SourceInstance); err != nil {
							logf.Log.Info("failed to send", "error", err)
						} else {
							wlist.DeleteWorkloadById(wl.ID)
						}
					}
				}
			}
		}
		wlist.Unlock()
	}
}

func (workloadMonitor *WorkloadMonitor) StartServer() {
	logf.Log.Info("API server started")

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
	server.ConfigureAPI()

	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}

}
