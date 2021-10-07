package single_msn_shutdown

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"os/exec"
	"time"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *appConfig) createSC() {

	err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithLocal(true).
		WithIOTimeout(common.DefaultIOTimeout).
		WithFileSystemType(c.fsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.scName)
}

func (c *appConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

func (c *appConfig) createPVC() string {
	// Create the volume with 1 replica
	return k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, common.VolFileSystem, common.NSDefault)
}

func (c *appConfig) deletePVC() {
	k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
}

func (c *appConfig) createDeployment() {

	labelselector := map[string]string{
		"e2e-test": c.deployName,
	}
	nodeSelector := map[string]string{
		"kubernetes.io/hostname": c.nodeName,
	}

	mount := corev1.VolumeMount{
		Name:      "ms-volume",
		MountPath: common.FioFsMountPoint,
	}
	var volMounts []corev1.VolumeMount
	volMounts = append(volMounts, mount)

	args := []string{"sleep", "1000000"}

	deployObj, err := k8stest.NewDeploymentBuilder().
		WithName(c.deployName).
		WithNamespace(common.NSDefault).
		WithLabelsNew(labelselector).
		WithSelectorMatchLabelsNew(labelselector).
		WithPodTemplateSpecBuilder(
			k8stest.NewPodtemplatespecBuilder().
				WithNodeSelector(nodeSelector).
				WithLabels(labelselector).
				WithContainerBuildersNew(
					k8stest.NewContainerBuilder().
						WithName(c.deployName).
						WithImage(common.GetFioImage()).
						WithImagePullPolicy(corev1.PullAlways).
						WithVolumeMountsNew(volMounts).
						WithArgumentsNew(args)).
				WithVolumeBuilders(
					k8stest.NewVolumeBuilder().
						WithName("ms-volume").
						WithPVCSource(c.pvcName),
				),
		).
		Build()
	Expect(err).ShouldNot(
		HaveOccurred(),
		"while building delpoyment {%s} in namespace {%s}",
		c.deployName,
		common.NSDefault,
	)

	Expect(err).ToNot(HaveOccurred(), "Generating deployment definition %s", c.deployName)
	err = k8stest.CreateDeployment(deployObj)
	Expect(err).ToNot(HaveOccurred(), "Creating deployment %s", c.deployName)

	c.verifyApplicationPodRunning(true)
	go c.fioWriteOnly(c.podName, "sha1", thinkTime)
}

func (c *appConfig) deleteDeployment() {
	err := k8stest.DeleteDeployment(c.deployName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Deleting deployment %s", c.deployName)
	c.verifyApplicationPodRunning(false)
}

func verifyMayastorComponentStates(numMayastorInstances int) {
	// TODO Enable this once issue is fixed in Mayastor
	/*
		nodeList, err := crds.ListNodes()
		Expect(err).ToNot(HaveOccurred(), "ListNodes")
		count := 0
		for _, node := range nodeList {
			if node.Status == "online" {
				count++
			}
		}
		Expect(count).To(Equal(numMayastorInstances))
	*/
	ready, err := k8stest.MayastorInstancesReady(numMayastorInstances, 3, 540)
	Expect(err).ToNot(HaveOccurred())
	Expect(ready).To(Equal(true))
	// FIXME: MCP is this correct?
	ready = k8stest.ControlPlaneReady(3, 60)
	Expect(ready).To(Equal(true), "control plane is not ready")
}

func (c *appConfig) verifyApplicationPodRunning(state bool) {
	var (
		podName       string
		runningStatus bool
		err           error
	)
	labels := "e2e-test=" + c.deployName
	logf.Log.Info("Verify application deployment ready", "state", state)
	Eventually(func() bool {
		runningStatus = k8stest.DeploymentReady(c.deployName, common.NSDefault)
		return runningStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(state))

	logf.Log.Info("Verify application pod running", "state", state)
	Eventually(func() bool {
		podName, runningStatus, err = k8stest.IsPodWithLabelsRunning(labels, common.NSDefault)
		Expect(err).ToNot(HaveOccurred())
		return runningStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(state))
	c.podName = podName
}

func verifyNodeNotReady(nodeName string) {
	Eventually(func() bool {
		readyStatus, err := k8stest.IsNodeReady(nodeName, nil)
		Expect(err).ToNot(HaveOccurred())
		return readyStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(false))

	// TODO Enable it once fixed in Mayastor
	/*
		Eventually(func() string {
			msn, err := crds.GetNode(nodeName)
			Expect(err).ToNot(HaveOccurred(), "GetNode")
			return msn.Status
		},
			defTimeoutSecs, // timeout
			"5s",           // polling interval
		).Should(Equal("offline"))
	*/
}

func verifyNodesReady() {
	Eventually(func() bool {
		readyStatus, err := k8stest.AreNodesReady()
		Expect(err).ToNot(HaveOccurred())
		return readyStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(true))
}

// write to all blocks with a block-specific pattern and its checksum
func (c *appConfig) fioWriteOnly(fioPodName string, hash string, thinkTime int) {
	verifyParam := fmt.Sprintf("--verify=%s", hash)
	// thinkTime is being added to control the time of execution of fio
	thinkTimeParam := fmt.Sprintf("--thinktime=%d", thinkTime)
	thinkTimeBlocksParam := fmt.Sprintf("--thinktime_blocks=%d", thinkTimeBlocks)

	var err error
	ch := make(chan bool, 1)

	go func() {
		_, err = runFio(
			fioPodName,
			common.FioFsFilename,
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
		if err != nil {
			logf.Log.Info("FIO failed", "podName", c.podName, "err", err)
			c.taskCompletionStatus = "failed"
		}
		c.taskCompletionStatus = "success"
	case <-time.After(time.Duration(fioTimeoutSecs) * time.Second):
		logf.Log.Info("FIO timedout", "podName", c.podName)
		c.taskCompletionStatus = "failed"
	}
}

// Run fio against the device, finish when all blocks are accessed
func runFio(podName string, filename string, args ...string) ([]byte, error) {
	argFilename := fmt.Sprintf("--filename=%s", filename)
	volumeFileSize := fmt.Sprintf("%dM", volumeFileSizeMb)
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
		"--size=" + volumeFileSize,
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

func (c *appConfig) verifyTaskCompletionStatus(status string) {
	Eventually(func() string {
		Expect(c.taskCompletionStatus).NotTo(Equal("failed"))
		logf.Log.Info("Verify task completion", "pod", c.podName, "status", c.taskCompletionStatus)
		return c.taskCompletionStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal("success"))

}

func getMsvState(uuid string) string {
	volState, err := k8stest.GetMsvState(uuid)
	Expect(err).To(BeNil(), "failed to access volume state %s, error=%v", uuid, err)
	return volState
}
