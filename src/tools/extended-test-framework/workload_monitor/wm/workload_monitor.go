package wm

import (
	"fmt"
	"log"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/strfmt"
	flags "github.com/jessevdk/go-flags"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"
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

var current_registrant *strfmt.UUID = nil

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

// Who guards the guard ? This does.
func (workloadMonitor *WorkloadMonitor) WaitSignal() {
	exitCode := <-workloadMonitor.channel
	if exitCode != 0 { // abnormal termination
		if err := SendEvent(workloadMonitor.pTestDirectorClient, "workload monitor terminated", "workload-monitor", current_registrant); err != nil {
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
	logf.Log.Info("workload monitor polling .")

	for {
		time.Sleep(10 * time.Second)

		// Lock the list while iterating.
		// This ensures that the tests reflect exactly what was most
		// recently specified by the test conductor. This avoids a
		// potential race issue if the test conductor removes a pod from
		// the list, removes the corresponding pod and the workload monitor
		// sees it as missing because it has old data.
		// The alternative - locking the list, copying the data and unlocking -
		// would suffer from this race issue.
		// This will obviously add latency to REST calls from the test_conductor,
		// but hopefully not problematically.
		wlist.Lock()
		if current_registrant == nil {
			current_registrant = wlist.GetRegistrant()
			if current_registrant != nil {
				logf.Log.Info("Using registrant", "rid", *current_registrant)
			}
		}
		if current_registrant != nil {

			list := wlist.GetWorkloadListByRegistrant(*current_registrant)

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
								if err := SendEvent(workloadMonitor.pTestDirectorClient, message, string(wl.Name), current_registrant); err != nil {
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
								if err := SendEvent(workloadMonitor.pTestDirectorClient, message, string(wl.Name), current_registrant); err != nil {
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

							if err := SendEvent(workloadMonitor.pTestDirectorClient, message, string(wl.Name), current_registrant); err != nil {
								logf.Log.Info("failed to send", "error", err)
							} else {
								wlist.DeleteWorkloadById(wl.ID)
							}
						}
					}
				}
			}
		}
		wlist.Unlock()
	}
}

// this code copied from generated server code
func (workloadMonitor *WorkloadMonitor) StartServer() {
	logf.Log.Info("API server started")

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		log.Fatalln(err)
	}

	api := operations.NewEtfwAPI(swaggerSpec)
	server := restapi.NewServer(api)

	// following call commented-out as it violates "unchecked return value" lint error
	// TODO fix this
	// defer server.Shutdown()

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
