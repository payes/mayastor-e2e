package tests

import (
	"fmt"
	"hash/fnv"
	"mayastor-e2e/tools/extended-test-framework/common/custom_resources"
	"mayastor-e2e/tools/extended-test-framework/common/mini_mcp_client"
	"strconv"

	"mayastor-e2e/tools/extended-test-framework/common"

	"mayastor-e2e/tools/extended-test-framework/common/k8sclient"

	td "mayastor-e2e/tools/extended-test-framework/common/td/models"
	"mayastor-e2e/tools/extended-test-framework/common/wm/models"
	"mayastor-e2e/tools/extended-test-framework/test_conductor/tc"
	"regexp"
	"time"

	v1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type VolSpec struct {
	sc_names    []string
	vol_types   []k8sclient.VolumeType
	vol_size_mb int
}

const MCP_NEXUS_ONLINE = "NEXUS_ONLINE"
const MCP_NEXUS_DEGRADED = "NEXUS_DEGRADED"
const MCP_NEXUS_FAULTED = "NEXUS_FAULTED"
const MCP_NEXUS_UNKNOWN = "NEXUS_UNKNOWN"

const MCP_MSV_DEGRADED = "Degraded"
const MCP_MSV_FAULTED = "Faulted"
const MCP_MSV_ONLINE = "Online"
const MCP_MSV_UNKNOWN = "Unknown"

const PV_PREFIX = "pvc-"

var violations = []models.WorkloadViolationEnum{
	models.WorkloadViolationEnumRESTARTED,
	models.WorkloadViolationEnumTERMINATED,
	models.WorkloadViolationEnumNOTPRESENT,
}

func checkPVCexists(vol_id string) (bool, error) {
	pvc_list, err := k8sclient.ListPVCs(k8sclient.NSDefault)
	if err != nil {
		return false, err
	}
	for _, pvc := range pvc_list.Items {
		if pvc.Spec.VolumeName == PV_PREFIX+vol_id {
			return true, err
		}
	}
	return false, err
}

func CheckMSVwithCP(ms_ip string) error {
	vols, err := mini_mcp_client.GetVolumes(ms_ip)
	if err != nil {
		return fmt.Errorf("failed to get volume uuids, error: %v", err)
	}
	for _, vol := range vols {
		switch vol.State.Status {

		case MCP_MSV_FAULTED:
			exists, err := checkPVCexists(vol.State.Uuid)
			if err != nil {
				return fmt.Errorf("failed to check PVC exists, error: %v", err)
			} else if exists {
				return fmt.Errorf("found faulted volume, uuid %s", vol.State.Uuid)
			} else {
				logf.Log.Info("Spurious faulted volume", "uuid", vol.State.Uuid)
			}
		case MCP_MSV_ONLINE:
		case MCP_MSV_DEGRADED:
		case MCP_MSV_UNKNOWN:

		default:
			return fmt.Errorf("MSV unexpected state %s in volume %s", vol.State.Status, vol.State.Uuid)
		}
	}
	return nil
}

func CheckMSVwithGrpc(ms_ips []string) error {
	states, err := k8sclient.GetNexusStates(ms_ips)
	if err != nil {
		return fmt.Errorf("Nexus grpc check failed, err: %v", err)
	}
	for _, state := range states {
		switch state {

		case MCP_NEXUS_FAULTED:
			return fmt.Errorf("Nexus is not healthy, state is %s", state)

		case MCP_NEXUS_UNKNOWN:
		case MCP_NEXUS_DEGRADED:
		case MCP_NEXUS_ONLINE:

		default:
			return fmt.Errorf("Nexus unexpected state %s", state)
		}
	}
	return nil
}

func CheckPools(poolcount int) error {
	if err := custom_resources.CheckAllMsPoolsAreOnline(); err != nil {
		return fmt.Errorf("Not all pools are healthy, error: %v", err)
	}
	pools, err := custom_resources.ListMsPools()
	if err != nil {
		return fmt.Errorf("failed to list pools, err: %v", err)
	}
	if len(pools) != poolcount {
		return fmt.Errorf("Wrong number of pools, expected %d, have %d", poolcount, len(pools))
	}
	return nil
}

func CheckNodes(nodecount int) error {
	nodeList, err := custom_resources.ListMsNodes()
	if err != nil {
		return fmt.Errorf("Failed to list MS Nodes, error: %v", err)
	}
	if len(nodeList) != nodecount {
		return fmt.Errorf("Wrong number of nodes, expected %d, have %d", nodecount, len(nodeList))
	}
	for _, node := range nodeList {
		if node.Status != "online" {
			return fmt.Errorf("Found offline node %s", node.Name)
		}
	}
	return nil
}

func CombineErrors(first error, second error) error {
	if first == nil {
		return second
	} else {
		return fmt.Errorf("%s: %s", first.Error(), second.Error())
	}
}

func ConvertTime(timestr string) (string, error) {
	// convert days part of time string to hours
	// e.g. converts "2d8h20m30s" to "56h20m30s"
	//golang duration strings do not parse days
	result := ""

	r, _ := regexp.Compile("^[0-9]{1,3}d") // are days in the string ?
	daystrarr := r.FindStringSubmatch(timestr)
	hours := 0
	switch len(daystrarr) {
	case 0: // no day field
	case 1:
		daystr := daystrarr[0]
		timestr = timestr[len(daystr):] // trim the days from the string
		daystr = daystr[:len(daystr)-1] // lose the 'd'
		days, err := strconv.Atoi(daystr)
		if err != nil {
			return "", fmt.Errorf("Internal error, failed to convert time, error %v", err)
		}
		rh, _ := regexp.Compile("^[0-9]{1,2}h") // are hours in the string?
		hourstrarr := rh.FindStringSubmatch(timestr)

		switch len(hourstrarr) {
		case 0: // no hour field
		case 1:
			hourstr := hourstrarr[0]
			timestr = timestr[len(hourstr):]   // trim the hours from the string
			hourstr = hourstr[:len(hourstr)-1] // lose the 'h'
			hours, err = strconv.Atoi(hourstr)
			if err != nil {
				return "", fmt.Errorf("Internal error, failed to convert time, error %v", err)
			}
		default: // more than 1 hour field
			return "", fmt.Errorf("Internal error, failed to convert time")
		}
		hours += days * 24
		if hours != 0 {
			result = fmt.Sprintf("%dh", hours)
		}
	default:
		return "", fmt.Errorf("Internal error, failed to convert time")
	}
	return result + timestr, nil
}

func GetDuration(durationstr string) (time.Duration, error) {
	durationstr, err := ConvertTime(durationstr)
	if err != nil {
		return time.Duration(0), fmt.Errorf("failed to convert duration %v", err)
	}
	logf.Log.Info("Converted duration", "in hours", durationstr)
	duration, err := time.ParseDuration(durationstr)
	if err != nil {
		return time.Duration(0), fmt.Errorf("failed to parse duration %v", err)
	}
	return duration, err
}

// if pod_to_check is given, the test wait for the pod to complete and duration is a timeout
// otherwise run the check for the full length of duration.
func MonitorCRs(
	testConductor *tc.TestConductor,
	duration time.Duration,
	pod_to_check string) error {

	var endTime = time.Now().Add(duration)
	var elapsedSecs = 0
	const waitSecs = 5
	const progressSecs = 240
	for {
		if elapsedSecs%progressSecs == 0 {
			logf.Log.Info("Monitoring CRs", "hours", elapsedSecs/3600, "minutes", elapsedSecs/60)
		}
		ms_ips, err := k8sclient.GetMayastorNodeIPs()
		if err != nil {
			return fmt.Errorf("MSV grpc check failed to get nodes, err: %s", err.Error())
		}
		if len(ms_ips) == 0 {
			return fmt.Errorf("No MS nodes found")
		}
		if err := CheckMSVwithGrpc(ms_ips); err != nil {
			return fmt.Errorf("MSV grpc check failed, err: %s", err.Error())
		}
		if err := CheckMSVwithCP(ms_ips[0]); err != nil {
			return fmt.Errorf("MSV CP check failed, err: %s", err.Error())
		}
		if err := CheckPools(testConductor.Config.Msnodes); err != nil {
			return fmt.Errorf("MSP check failed, err: %s", err.Error())
		}
		if pod_to_check != "" {
			pod, err := k8sclient.GetPod(pod_to_check, k8sclient.NSDefault)
			if err != nil {
				return fmt.Errorf("Failed to get application pod %s, err: %s", pod_to_check, err.Error())
			}
			if pod.Status.Phase != v1.PodRunning {
				break
			}
		}
		if time.Now().After(endTime) {
			if pod_to_check == "" {
				break // the run is time-limited, force stop
			} else {
				// we expected the pod to complete by now
				return fmt.Errorf("Pod is still running, timeout triggered")
			}
		}
		time.Sleep(time.Duration(waitSecs) * time.Second)
		elapsedSecs += waitSecs
	}
	return nil
}

// EtfwRandom - effective pseudo -random integer generator
// range 0 -> valrange-1 inclusive
// Doesn't need seeding
// The rand library doesn't seem to work very well
func EtfwRandom(valrange uint32) (int, error) {
	tmstr := fmt.Sprintf("%x", time.Now().UTC().UnixNano())
	h := fnv.New32a()
	_, err := h.Write([]byte(tmstr))
	if err != nil { // very unlikely
		return 0, fmt.Errorf("Internal checksum error, error: %v", err)
	}
	return int(h.Sum32() % valrange), err
}

func WaitPodNotRunning(pod_to_check string, timeout time.Duration) error {

	var endTime = time.Now().Add(timeout)
	var waitSecs = 5

	for {
		pod, err := k8sclient.GetPod(pod_to_check, k8sclient.NSDefault)
		if err != nil {
			return fmt.Errorf("Failed to get application pod %s, err: %s", pod_to_check, err.Error())
		}
		if pod.Status.Phase != v1.PodRunning {
			break
		}
		if time.Now().After(endTime) {
			return fmt.Errorf("Pod %s is still running, timeout triggered", pod_to_check)
		}
		time.Sleep(time.Duration(waitSecs) * time.Second)
	}
	return nil
}

func SendTestRunToDo(testConductor *tc.TestConductor) error {
	if testConductor.Config.SendXrayTest == 1 {
		text := testDescription(testConductor) + " Test run: preparing"
		if err := common.SendTestRunToDo(
			testConductor.TestDirectorClient,
			testConductor.TestRunId,
			text,
			testConductor.Config.XrayTestID); err != nil {

			return fmt.Errorf("failed to create test run, error: %v", err)
		}
	}
	if err := SendEventTestPreparing(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of preparation event, error: %v", err)
	}
	return nil
}

func SendTestRunStarted(testConductor *tc.TestConductor) error {
	if testConductor.Config.SendXrayTest == 1 {
		text := testDescription(testConductor) + " Test run: started"
		if err := common.SendTestRunStarted(
			testConductor.TestDirectorClient,
			testConductor.TestRunId,
			text,
			testConductor.Config.XrayTestID); err != nil {

			return fmt.Errorf("failed to set test run to executing, error: %v", err)
		}
	}
	if err := SendEventTestStarted(testConductor); err != nil {
		return fmt.Errorf("failed to inform test director of start event, error: %v", err)
	}
	return nil
}

func SendTestRunFinished(testConductor *tc.TestConductor, err error) error {
	if testConductor.Config.SendXrayTest == 1 {
		text := testDescription(testConductor) + " Test run: finished"
		if err == nil {
			if err := common.SendTestRunCompletedOk(
				testConductor.TestDirectorClient,
				testConductor.TestRunId,
				text,
				testConductor.Config.XrayTestID); err != nil {
				return fmt.Errorf("failed to set test run to completed, error: %v", err)
			}
		} else {
			if err := common.SendTestRunCompletedFail(
				testConductor.TestDirectorClient,
				testConductor.TestRunId,
				err.Error(),
				testConductor.Config.XrayTestID); err != nil {
				return fmt.Errorf("failed to set test run to failed, error: %v", err)
			}
		}
	}
	if err == nil {
		if err := SendEventTestCompletedOk(testConductor); err != nil {
			return fmt.Errorf("failed to inform test director of completion event, error: %v", err)
		}
	} else {
		if err := SendEventTestCompletedFail(testConductor, err.Error()); err != nil {
			return fmt.Errorf("failed to inform test director of completion event, error: %v", err)
		}
	}
	return nil
}

func testDescription(testConductor *tc.TestConductor) string {
	return testConductor.Config.XrayTestID + ", " + testConductor.Config.TestName + ", " + testConductor.Config.RunName
}

func SendEventTestPreparing(testConductor *tc.TestConductor) error {
	return SendEvent(
		testConductor,
		td.EventClassEnumINFO,
		"Event: test preparing on node: "+testConductor.NodeName)
}

func SendEventTestStarted(testConductor *tc.TestConductor) error {
	return SendEvent(testConductor, td.EventClassEnumINFO, "Event: test started")
}

func SendEventTestCompletedOk(testConductor *tc.TestConductor) error {
	return SendEvent(testConductor, td.EventClassEnumINFO, "Event: test completed")
}

func SendEventTestCompletedFail(testConductor *tc.TestConductor, text string) error {
	return SendEvent(testConductor, td.EventClassEnumFAIL, "Event: test failure: "+text)
}

func SendEvent(testConductor *tc.TestConductor, eventClass td.EventClassEnum, message string) error {
	if testConductor.Config.SendEvent == 1 {
		if err := common.SendEvent(
			testConductor.TestDirectorClient,
			testConductor.TestRunId,
			testDescription(testConductor)+" "+message,
			eventClass,
			td.EventSourceClassEnumTestDashConductor); err != nil {
			return fmt.Errorf("failed to inform test director of event, error: %v", err)
		}
	}
	return nil
}
