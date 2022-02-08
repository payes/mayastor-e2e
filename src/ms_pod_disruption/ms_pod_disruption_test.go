package ms_pod_disruption

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"mayastor-e2e/common"
	e2e_agent "mayastor-e2e/common/e2e-agent"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defTimeoutSecs = "180s"
	mayastorRegexp = "^mayastor-.....$"
	engineLabel    = "openebs.io/engine"
	mayastorLabel  = "mayastor"
)

type DisruptionEnv struct {
	replicaToRemove          string
	unusedNodes              []string
	uuid                     string
	volToDelete              string
	storageClass             string
	fioPodName               string
	nexusIP                  string
	nexusLocalRep            string
	podUnscheduleTimeoutSecs int
	podRescheduleTimeoutSecs int
	removeThinkTime          int
	repairThinkTime          int
	thinkTimeBlocks          int
	rebuildTimeoutSecs       int
	fioTimeoutSecs           int
	nexusUuid                string
}

var env DisruptionEnv

func getMsvState(uuid string) string {
	volState, err := k8stest.GetMsvState(uuid)
	Expect(err).To(BeNil(), "failed to access volume state %s, error=%v", uuid, err)
	return volState
}

// Identify the nexus IP address,
// the uri of the replica local to the nexus,
// and the non-nexus node hosting a replica.
func (env *DisruptionEnv) getNodes() {
	nodeList, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	nexus, replicaNodes := k8stest.GetMsvNodes(env.uuid)
	Expect(nexus).NotTo(Equal(""), "Nexus not found")

	// identify the nexus IP address
	nexusIP := ""
	for _, node := range nodeList {
		if node.NodeName == nexus {
			nexusIP = node.IPAddress
			break
		}
	}
	Expect(nexusIP).NotTo(Equal(""), "Nexus IP not found")
	env.nexusIP = nexusIP

	var nxlist []string
	nxlist = append(nxlist, nexusIP)

	nexusList, _ := mayastorclient.ListNexuses(nxlist)
	Expect(len(nexusList)).To(Equal(1), "Expected to find 1 nexus")
	nx := nexusList[0]

	// identify the replica local to the nexus
	nxChild := ""
	for _, ch := range nx.Children {
		if strings.HasPrefix(ch.Uri, "bdev:///") {
			Expect(nxChild).To(Equal(""), "More than 1 nexus local replica found")
			nxChild = ch.Uri
			break
		}
	}
	Expect(nxChild).NotTo(Equal(""), "Could not find nexus local replica")
	env.nexusLocalRep = nxChild

	// find a node which is not the nexus and is a replica
	toRemove := ""
	for _, node := range replicaNodes {
		if node != nexus {
			toRemove = node
			break
		}
	}
	Expect(toRemove).NotTo(Equal(""), "Could not find a replica to remove")
	env.replicaToRemove = toRemove

	// get nexus uuid
	msv, err := k8stest.GetMSV(env.uuid)
	Expect(err).ToNot(HaveOccurred(), "failed to retrieve MSV for volume %s", env.uuid)
	env.nexusUuid = msv.Status.Nexus.Uuid

	logf.Log.Info("identified", "nexus IP", env.nexusIP, "local replica", env.nexusLocalRep, "node of replica to remove", env.replicaToRemove)
}

// select nodes to suppress so that the number of available mayastor
// instances equals the number of replicas we will use
func (env *DisruptionEnv) suppressSpareNodes() {
	nodeList, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	maxReplicas := 2
	mayastorInstances := 0
	var unusedNodes []string

	for _, node := range nodeList {
		if node.MayastorNode {
			mayastorInstances++
			if mayastorInstances > maxReplicas {
				env.suppressMayastorPodOn(node.NodeName, 0)
				unusedNodes = append(unusedNodes, node.NodeName)
			}
		}
	}
	// we need at least 1 spare node to re-enable to allow the volume to become healthy again
	Expect(len(unusedNodes)).NotTo(Equal(0), "Need at least 1 unused mayastor node")
	env.unusedNodes = unusedNodes
}

// Common steps required when setting up the test.
// Removes excess mayastor instances, creates the PVC,
// deploys fio, and records variables needed for the
// test in the DisruptionEnv structure
func setup(pvcName string, storageClassName string, fioPodName string) DisruptionEnv {
	var err error
	e2eCfg := e2e_config.GetConfig()
	volMb := e2eCfg.MsPodDisruption.VolMb
	env := DisruptionEnv{}

	env.podUnscheduleTimeoutSecs = e2eCfg.MsPodDisruption.PodUnscheduleTimeoutSecs
	env.podRescheduleTimeoutSecs = e2eCfg.MsPodDisruption.PodRescheduleTimeoutSecs
	env.removeThinkTime = e2eCfg.MsPodDisruption.RemoveThinkTime
	env.repairThinkTime = e2eCfg.MsPodDisruption.RepairThinkTime
	env.thinkTimeBlocks = e2eCfg.MsPodDisruption.ThinkTimeBlocks
	env.rebuildTimeoutSecs = volMb / 5 // rebuild timeout depends on volume size, e.g. 100s for 512Mb
	env.fioTimeoutSecs = volMb / 2     // fio run should take longer than a re-build, use thinkTime to ensure this

	env.suppressSpareNodes()

	env.volToDelete = pvcName
	env.storageClass = storageClassName
	env.uuid, err = k8stest.MkPVC(volMb, pvcName, storageClassName, common.VolRawBlock, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", pvcName)
	podObj := k8stest.CreateFioPodDef(fioPodName, pvcName, common.VolRawBlock, common.NSDefault)
	// add node selector to fio pod
	podObj.Spec.NodeSelector = map[string]string{
		common.MayastorEngineLabel: common.MayastorEngineLabelValue,
	}
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	env.fioPodName = fioPodName
	logf.Log.Info("waiting for pod", "name", env.fioPodName)
	Eventually(func() bool {
		return k8stest.IsPodRunning(env.fioPodName, common.NSDefault)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(true))

	env.getNodes()
	return env
}

// Common steps required when tearing down the test
func (env *DisruptionEnv) teardown() {
	var err error

	if env.replicaToRemove != "" {
		env.unsuppressMayastorPodOn(env.replicaToRemove, 0)
		env.replicaToRemove = ""
	}
	for _, node := range env.unusedNodes {
		env.unsuppressMayastorPodOn(node, 0)
	}
	if env.fioPodName != "" {
		err := k8stest.DeletePod(env.fioPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env.fioPodName = ""
	}
	if env.volToDelete != "" {
		err := k8stest.RmPVC(env.volToDelete, env.storageClass, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", env.volToDelete)
		env.volToDelete = ""
	}
	if env.storageClass != "" {
		err = k8stest.RmStorageClass(env.storageClass)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env.storageClass = ""
	}
}

// Prevent mayastor pod from running on the given node.
// Action can be delayed to ensure overlap with IO in main thread.
func (env *DisruptionEnv) suppressMayastorPodOn(nodeName string, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	logf.Log.Info("suppressing mayastor pod", "node", nodeName)
	err := k8stest.UnlabelNode(nodeName, engineLabel)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	err = k8stest.WaitForPodNotRunningOnNode(mayastorRegexp, common.NSMayastor(), nodeName, env.podUnscheduleTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func (env *DisruptionEnv) disablePoolDeviceAtIp(ipAddress string, device string, delaySecs int) {
	time.Sleep(time.Duration(delaySecs) * time.Second)
	logf.Log.Info("disabling device", "address", ipAddress, "device", device)
	_, err := e2e_agent.ControlDevice(ipAddress, device, "offline")
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func (env *DisruptionEnv) enablePoolDeviceAtIp(ipAddress string, device string, delaySecs int) {
	time.Sleep(time.Duration(delaySecs) * time.Second)
	logf.Log.Info("enabling device", "address", ipAddress, "device", device)
	_, err := e2e_agent.ControlDevice(ipAddress, device, "running")
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func (env *DisruptionEnv) disablePoolDeviceAtNode(nodeName string, device string, delaySecs int) {
	logf.Log.Info("disabling device", "nodeName", nodeName, "device", device)
	pIpAddress, err := k8stest.GetNodeIPAddress(nodeName)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	env.disablePoolDeviceAtIp(*pIpAddress, device, delaySecs)
}

func (env *DisruptionEnv) enablePoolDeviceAtNode(nodeName string, device string, delaySecs int) {
	logf.Log.Info("enabling device", "NodeName", nodeName, "device", device)
	pIpAddress, err := k8stest.GetNodeIPAddress(nodeName)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	env.enablePoolDeviceAtIp(*pIpAddress, device, delaySecs)
}

// Allow mayastor pod to run on the given node.
// Action can be delayed to ensure overlap with IO in main thread.
func (env *DisruptionEnv) unsuppressMayastorPodOn(nodeName string, delay int) {
	// add the mayastor label to the node
	time.Sleep(time.Duration(delay) * time.Second)
	logf.Log.Info("restoring mayastor pod", "node", nodeName)
	err := k8stest.LabelNode(nodeName, engineLabel, mayastorLabel)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	err = k8stest.WaitForPodRunningOnNode(mayastorRegexp, common.NSMayastor(), nodeName, env.podRescheduleTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

// Fault the replica hosted on the nexus node
// Action can be delayed to ensure overlap with IO in main thread.
func (env *DisruptionEnv) faultNexusChild(delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	logf.Log.Info("faulting the nexus replica")
	err := mayastorclient.FaultNexusChild(env.nexusIP, env.nexusUuid, env.nexusLocalRep)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

// Run fio against the device, finish when all blocks are accessed
func runFio(podName string, filename string, args ...string) ([]byte, error) {
	argFilename := fmt.Sprintf("--filename=%s", filename)

	logf.Log.Info("RunFio",
		"podName", podName,
		"filename", filename,
		"args", args)

	cmdArgs := []string{
		"exec",
		"-it",
		podName,
		"--",
		"fio",
		"--name=benchtest",
		"--verify_fatal=1",
		"--verify_async=2",
		argFilename,
		"--direct=1",
		"--ioengine=libaio",
		"--bs=4k",
		"--iodepth=16",
		"--numjobs=1",
	}

	if args != nil {
		cmdArgs = append(cmdArgs, args...)
	}

	cmd := exec.Command(
		"kubectl",
		cmdArgs...,
	)
	cmd.Dir = ""
	output, err := cmd.CombinedOutput()
	if err != nil {
		logf.Log.Info("Running fio failed", "error", err, "output", string(output))
	}
	return output, err
}

// write to all blocks with a block-specific pattern and its checksum
func (env *DisruptionEnv) fioWriteOnly(fioPodName string, hash string, thinkTime int) error {
	verifyParam := fmt.Sprintf("--verify=%s", hash)
	thinkTimeParam := fmt.Sprintf("--thinktime=%d", thinkTime)
	thinkTimeBlocksParam := fmt.Sprintf("--thinktime_blocks=%d", env.thinkTimeBlocks)

	var err error
	ch := make(chan bool, 1)

	go func() {
		_, err = runFio(
			fioPodName,
			common.FioBlockFilename,
			"--rw=randwrite",
			"--do_verify=0",
			verifyParam,
			"--verify_pattern=%o",
			thinkTimeParam,
			thinkTimeBlocksParam)
		ch <- true
	}()
	select {
	case <-ch:
		return err
	case <-time.After(time.Duration(env.fioTimeoutSecs) * time.Second):
		return fmt.Errorf("Fio timed out")
	}
}

// verify the content of all the blocks
func fioVerify(fioPodName string, hash string) error {
	verifyParam := fmt.Sprintf("--verify=%s", hash)

	ch := make(chan bool, 1)
	var err error

	go func() {
		_, err = runFio(
			fioPodName,
			common.FioBlockFilename,
			"--rw=randread",
			verifyParam)
		ch <- true
	}()
	select {
	case <-ch:
		return err
	case <-time.After(time.Duration(env.fioTimeoutSecs) * time.Second):
		return fmt.Errorf("Fio timed out")
	}
}

var _ = Describe("Mayastor replica pod removal test", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred(), "%v", err)
	})

	AfterEach(func() {
		env.teardown() // removes fio pod and volume
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred(), "%v", err)
	})

	It("should verify nvmf nexus behaviour when a mayastor pod is removed", func() {
		if e2e_config.GetConfig().MsPodDisruption.PodRemovalTest == 0 {
			Skip("Not executing pod removal / nexus child-fault test")
		}
		sc := "sc-ms-pod-remove-test-write-continuously"
		err := k8stest.MkStorageClass(sc, 2, common.ShareProtoNvmf, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env = setup("pvc-ms-pod-remove-test-write-continuously", sc, "fio-ms-pod-remove-test-write-continuously")
		env.PodLossTestWriteContinuously()

	})
	It("should verify nvmf nexus behaviour when a backing device is offlined", func() {
		if e2e_config.GetConfig().MsPodDisruption.DeviceRemovalTest == 0 {
			Skip("Not executing device removal test")
		}
		sc := "sc-offline-device-test"
		err := k8stest.MkStorageClass(sc, 2, common.ShareProtoNvmf, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env = setup("pvc-offline-device-test", sc, "fio-offline-device-test")
		env.DeviceLossTest()
	})
})

var _ = BeforeSuite(func(done Done) {
	err := k8stest.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnv %v", err)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := k8stest.TeardownTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)
})

func TestMayastorPodLoss(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Replica pod removal tests", "ms_pod_disruption")
}
