package node_shutdown

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *shutdownConfig) createSC() {

	scObj, err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithLocal(true).
		WithFileSystemType(c.fsType).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating storage class definition %s", c.scName)
	err = k8stest.CreateSc(scObj)
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
	command := []string{}
	command = append(command, "sleep", "3600")
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
						WithName("busybox").
						WithImage("busybox").
						WithCommandNew(command).
						WithVolumeMountsNew(
							[]corev1.VolumeMount{
								corev1.VolumeMount{
									Name:      "datavol1",
									MountPath: "/mnt/e2e-test",
								},
							},
						),
				).WithVolumeBuilders(
				k8stest.NewVolumeBuilder().
					WithName("datavol1").
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
	nodeList, err := custom_resources.ListNodes()
	Expect(err).ToNot(HaveOccurred(), "ListNodes")
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
		runningStatus, err := k8stest.IsPodWithLabelsRunning(labels, common.NSDefault)
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
		msn, err := custom_resources.GetNode(nodeName)
		Expect(err).ToNot(HaveOccurred(), "GetNode")
		return msn.Status
	},
		defTimeoutSecs, // timeout
		"5s",           // polling interval
	).Should(Equal("offline"))
}
