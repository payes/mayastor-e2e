package ms_pod_disruption

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defTimeoutSecs = "90s"
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
}

var env DisruptionEnv

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
	env.uuid = k8stest.MkPVC(volMb, pvcName, storageClassName, common.VolRawBlock, common.NSDefault)

	podObj := k8stest.CreateFioPodDef(fioPodName, pvcName, common.VolRawBlock, common.NSDefault)
	_, err := k8stest.CreatePod(podObj, common.NSDefault)
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

	env.unsuppressMayastorPodOn(env.replicaToRemove, 0)

	for _, node := range env.unusedNodes {
		env.unsuppressMayastorPodOn(node, 0)
	}
	if env.fioPodName != "" {
		err := k8stest.DeletePod(env.fioPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env.fioPodName = ""
	}
	if env.volToDelete != "" {
		k8stest.RmPVC(env.volToDelete, env.storageClass, common.NSDefault)
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
	k8stest.UnlabelNode(nodeName, engineLabel)
	err := k8stest.WaitForPodNotRunningOnNode(mayastorRegexp, common.NSMayastor, nodeName, env.podUnscheduleTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

// Allow mayastor pod to run on the given node.
// Action can be delayed to ensure overlap with IO in main thread.
func (env *DisruptionEnv) unsuppressMayastorPodOn(nodeName string, delay int) {
	// add the mayastor label to the node
	time.Sleep(time.Duration(delay) * time.Second)
	logf.Log.Info("restoring mayastor pod", "node", nodeName)
	k8stest.LabelNode(nodeName, engineLabel, mayastorLabel)
	err := k8stest.WaitForPodRunningOnNode(mayastorRegexp, common.NSMayastor, nodeName, env.podRescheduleTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

// Fault the replica hosted on the nexus node
// Action can be delayed to ensure overlap with IO in main thread.
func (env *DisruptionEnv) faultNexusChild(delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	logf.Log.Info("faulting the nexus replica")
	err := mayastorclient.FaultNexusChild(env.nexusIP, env.uuid, env.nexusLocalRep)
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

// write to all blocks with a block-specific pattern and its checksum
// verify the contents afterward
func fioWriteAndVerify(fioPodName string, hash string) error {
	verifyParam := fmt.Sprintf("--verify=%s", hash)

	var err error
	ch := make(chan bool, 1)

	go func() {
		_, err = runFio(
			fioPodName,
			common.FioBlockFilename,
			"--rw=randwrite",
			"--do_verify=1",
			verifyParam,
			"--verify_pattern=%o")
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

// PodLossTestDataCopy
// Run fio against the cluster while a replica mayastor pod is unscheduled and then rescheduled
// This is to verify that data written to a volume is completely copied to a new replica when
// all of the initial copies are removed.
// The sequence is:
// 1) write pattern to 2-replica volume, then verify the pattern
// 2) remove one non-nexus replica by unscheduling the mayastor pod
// 3) verify that the volume becomes degraded, then verify the data
// 4) enable a new replica,
// 5) verify that the volume becomes healthy, then verify the data
// 6) disable the nexus local replica
// 7) verify that the volume becomes degraded, then verify the data
//    This checks data that was never originally written
// 8) Unsuppress the first replica and wait for the volume to become healthy
func (env *DisruptionEnv) PodLossTestDataCopy() {

	// 1) Running fio with --do_verify=0, --verify=crc32 and --rw=randwrite means that only writes will occur
	// and no verification reads happen, verification can be done in the next step "off-line"
	// This step writes exactly once to each block
	logf.Log.Info("writing and verifying the volume")
	err := fioWriteAndVerify(env.fioPodName, "crc32")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 2) remove one non-nexus replica by unscheduling the mayastor pod
	logf.Log.Info("about to suppress mayastor on one replica")
	env.suppressMayastorPodOn(env.replicaToRemove, 0)

	// 3) wait for the volume to become degraded then run fio
	// Running fio with --verify=crc32 and --rw=randread means that only reads will occur
	// and verification is done
	Eventually(func() string {
		return k8stest.GetMsvState(env.uuid)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal("degraded"))
	logf.Log.Info("volume condition", "state", k8stest.GetMsvState(env.uuid))

	logf.Log.Info("verifying the degraded volume")
	err = fioVerify(env.fioPodName, "crc32")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 4) re-enable mayastor on one unused node
	logf.Log.Info("replacing the original replica")
	env.unsuppressMayastorPodOn(env.unusedNodes[0], 0)

	// 5) verify that the volume becomes healthy, then verify the data
	Eventually(func() string {
		return k8stest.GetMsvState(env.uuid)
	},
		env.rebuildTimeoutSecs, // timeout
		"1s",                   // polling interval
	).Should(Equal("healthy"))
	logf.Log.Info("volume condition", "state", k8stest.GetMsvState(env.uuid))

	logf.Log.Info("verifying the repaired volume")
	err = fioVerify(env.fioPodName, "crc32")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 6) disable the nexus local replica
	env.faultNexusChild(0)

	// 7) verify that the volume becomes degraded, then verify the data (only re-built data)
	Eventually(func() string {
		return k8stest.GetMsvState(env.uuid)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal("degraded"))
	logf.Log.Info("volume condition", "state", k8stest.GetMsvState(env.uuid))

	logf.Log.Info("verifying the degraded volume")
	err = fioVerify(env.fioPodName, "crc32")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	logf.Log.Info("restoring the original replica")
	env.unsuppressMayastorPodOn(env.replicaToRemove, 0)

	// 8) verify that the volume becomes healthy
	Eventually(func() string {
		return k8stest.GetMsvState(env.uuid)
	},
		env.rebuildTimeoutSecs, // timeout
		"1s",                   // polling interval
	).Should(Equal("healthy"))
	logf.Log.Info("volume condition", "state", k8stest.GetMsvState(env.uuid))
}

// Write data to a volume while replicas are being removed and added.
// After each transition, verify every block in the volume is correct
// 1) Write pattern-1 to every block in the volume once and simultaneously remove one replica
// 2) Verify that the volume becomes degraded and the data is correct
// 3) Write pattern-2 while adding a new replica to the volume while
// 4) Verify that the volume becomes healthy and the data is correct
// 5) Write pattern-3 while faulting the replica on the nexus node
// 6) Verify that the volume becomes degraded and the data is correct
// 7) Re enable the first replica, wait for the volume to become healthy
//    Verify that the data is still correct
func (env *DisruptionEnv) PodLossTestWriteContinuously() {
	e2eCfg := e2e_config.GetConfig()

	// 1) Write pattern-1 to every block in the volume once and simultaneously remove one replica
	// Running fio with --do_verify=0, --verify=crc32 and --rw=randwrite means that only writes will occur
	// and no verification reads happen, verification can be done in the next step "off-line"
	logf.Log.Info("about to suppress mayastor on one replica")
	go env.suppressMayastorPodOn(env.replicaToRemove, e2eCfg.MsPodDisruption.UnscheduleDelay)

	logf.Log.Info("writing to the volume")
	err := env.fioWriteOnly(env.fioPodName, "crc32", env.removeThinkTime)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 2) Verify that the volume has become degraded and the data is correct
	// We make the assumption that the volume has had enough time to become faulted
	Expect(k8stest.GetMsvState(env.uuid)).To(Equal("degraded"), "Unexpected MSV state")
	logf.Log.Info("volume condition", "state", k8stest.GetMsvState(env.uuid))

	// Running fio with --verify=crc32 and --rw=randread means that only reads will occur
	// and verification is performed
	logf.Log.Info("verifying the degraded volume")
	err = fioVerify(env.fioPodName, "crc32")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 3) Write pattern-2 while adding a new replica to the volume while
	// re-enable mayastor on one unused node
	logf.Log.Info("replacing the original replica")
	go env.unsuppressMayastorPodOn(env.unusedNodes[0], e2eCfg.MsPodDisruption.RescheduleDelay)

	// Random writes only. Note the checksum is now md5 and is stored in the block
	// Any blocks that do not get modified will fail the verification run (below)
	// because the stored checksum will still be crc32
	logf.Log.Info("writing to the volume")
	err = env.fioWriteOnly(env.fioPodName, "md5", env.repairThinkTime)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 4) Verify that the volume becomes healthy and the data is correct
	// We make the assumption that the volume has had enough time to be repaired
	Expect(k8stest.GetMsvState(env.uuid)).To(Equal("healthy"), "Unexpected MSV state")
	logf.Log.Info("volume condition", "state", k8stest.GetMsvState(env.uuid))

	// Verify the data just written.
	logf.Log.Info("verifying the repaired volume")
	err = fioVerify(env.fioPodName, "md5")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 5) Write pattern-3 while faulting the replica on the nexus node
	// remove the replica from the nexus
	go env.faultNexusChild(e2eCfg.MsPodDisruption.UnscheduleDelay)

	// Running fio with --do_verify=0, --verify=sha1 and --rw=randwrite means that only writes will occur
	// and no verification is performed at this point
	logf.Log.Info("writing to the volume")
	err = env.fioWriteOnly(env.fioPodName, "sha1", env.removeThinkTime)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 6) Verify that the volume becomes degraded and the data is correct
	// We make the assumption that the volume has had enough time to become faulted
	Expect(k8stest.GetMsvState(env.uuid)).To(Equal("degraded"), "Unexpected MSV state")

	// Running fio with --verify=sha1 and --rw=randread means that only reads will occur
	// and verification happens
	// This step reads each block once
	logf.Log.Info("verifying the degraded volume")
	err = fioVerify(env.fioPodName, "sha1")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	logf.Log.Info("restoring the original replica")
	env.unsuppressMayastorPodOn(env.replicaToRemove, 0)

	// 7) verify that the volume becomes healthy again
	Eventually(func() string {
		return k8stest.GetMsvState(env.uuid)
	},
		env.rebuildTimeoutSecs, // timeout
		"1s",                   // polling interval
	).Should(Equal("healthy"))
	logf.Log.Info("volume condition", "state", k8stest.GetMsvState(env.uuid))

	// Re-verify with the original replica on-line, It it gets any IO the
	// verification will fail because it contains the wrong data.
	logf.Log.Info("verifying the repaired volume")
	err = fioVerify(env.fioPodName, "sha1")
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func TestMayastorPodLoss(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Replica pod removal tests", "ms_pod_disruption")
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

	It("should verify nexus data is copied when a mayastor pod is removed", func() {
		sc := "mayastor-nvmf-pod-remove-test-sc-1"
		err := k8stest.MkStorageClass(sc, 2, common.ShareProtoNvmf, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env = setup("loss-test-pvc-1", sc, "fio-pod-remove-test-1")
		env.PodLossTestDataCopy()
	})

	It("should verify nvmf nexus behaviour when a mayastor pod is removed", func() {
		sc := "mayastor-nvmf-pod-remove-test-sc-2"
		err := k8stest.MkStorageClass(sc, 2, common.ShareProtoNvmf, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env = setup("loss-test-pvc-2", sc, "fio-pod-remove-test-2")
		env.PodLossTestWriteContinuously()
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
