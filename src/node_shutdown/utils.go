package node_shutdown

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *shutdownConfig) createSC() {

	err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithLocal(true).
		WithFileSystemType(c.fsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.scName)
}

func (c *shutdownConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

func (c *shutdownConfig) createPVC() string {
	// Create the volume with 1 replica
	return k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, common.VolFileSystem, common.NSDefault)
}

func (c *shutdownConfig) deletePVC() {
	k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
}

func (c *shutdownConfig) createDeployment() {

	labelselector := map[string]string{
		"e2e-test": "shutdown",
	}
	mount := corev1.VolumeMount{
		Name:      "ms-volume",
		MountPath: common.FioFsMountPoint,
	}
	var volMounts []corev1.VolumeMount
	volMounts = append(volMounts, mount)

	args := []string{
		"--",
		"--time_based",
		fmt.Sprintf("--runtime=%d", durationSecs),
		fmt.Sprintf("--filename=%s", common.FioFsFilename),
		fmt.Sprintf("--size=%dm", volumeFileSizeMb),
		fmt.Sprintf("--thinktime=%d", thinkTime),
	}

	fioArgs := append(args, common.GetFioArgs()...)
	logf.Log.Info("fio", "arguments", fioArgs)
	deployObj, err := k8stest.NewDeploymentBuilder().
		WithName(c.deployName).
		WithNamespace(common.NSDefault).
		WithLabelsNew(labelselector).
		WithSelectorMatchLabelsNew(labelselector).
		WithPodTemplateSpecBuilder(
			k8stest.NewPodtemplatespecBuilder().
				WithLabels(labelselector).
				WithContainerBuildersNew(
					k8stest.NewContainerBuilder().
						WithName(c.podName).
						WithImage(common.GetFioImage()).
						WithVolumeMountsNew(volMounts).
						WithImagePullPolicy(corev1.PullAlways).
						WithArgumentsNew(fioArgs)).
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
}

func (c *shutdownConfig) deleteDeployment() {
	err := k8stest.DeleteDeployment(c.deployName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Deleting deployment %s", c.deployName)
	c.verifyApplicationPodRunning(false)
}

func (c *shutdownConfig) verifyMayastorComponentStates(numMayastorInstances int) {
	nodeList, err := custom_resources.ListMsNodes()
	Expect(err).ToNot(HaveOccurred(), "ListMsNodes")
	count := 0
	for _, node := range nodeList {
		if node.Status == "online" {
			count++
		}
	}
	Expect(count).To(Equal(numMayastorInstances))
	ready, err := k8stest.MayastorInstancesReady(numMayastorInstances, 3, 540)
	Expect(err).ToNot(HaveOccurred())
	Expect(ready).To(Equal(true))
	// FIXME: MCP is this correct?
	ready = k8stest.ControlPlaneReady(3, 60)
	Expect(ready).To(Equal(true), "control is not ready")
}

func (c *shutdownConfig) verifyApplicationPodRunning(state bool) {
	labels := "e2e-test=shutdown"
	logf.Log.Info("Verify application deployment ready", "state", state)
	Eventually(func() bool {
		runningStatus := k8stest.DeploymentReady(c.deployName, common.NSDefault)
		return runningStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(state))

	logf.Log.Info("Verify application pod running", "state", state)
	Eventually(func() bool {
		_, runningStatus, err := k8stest.IsPodWithLabelsRunning(labels, common.NSDefault)
		Expect(err).ToNot(HaveOccurred())
		return runningStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(state))
}

func (c *shutdownConfig) verifyNodeNotReady(nodeName string) {
	Eventually(func() bool {
		readyStatus, err := k8stest.IsNodeReady(nodeName, nil)
		Expect(err).ToNot(HaveOccurred())
		return readyStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(false))

	Eventually(func() string {
		status, err := k8stest.GetMsNodeStatus(nodeName)
		Expect(err).ToNot(HaveOccurred(), "GetMsNodeStatus")
		return status
	},
		defTimeoutSecs, // timeout
		"5s",           // polling interval
	).Should(Equal(controlplane.NodeStateOffline()))
}
