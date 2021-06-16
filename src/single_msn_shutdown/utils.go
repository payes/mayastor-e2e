package single_msn_shutdown

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

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
	args := []string{
		"--",
		"--time_based",
		fmt.Sprintf("--runtime=%d", durationSecs),
		fmt.Sprintf("--filename=%s", common.FioFsFilename),
		fmt.Sprintf("--size=%dm", volumeFileSizeMb),
		fmt.Sprintf("--thinktime=%d", thinkTime),
	}
	cmds := []string{
		"touch",
		"livenessProbe",
	}
	probe := &corev1.Probe{
		PeriodSeconds:       periodSeconds,
		InitialDelaySeconds: initialDelaySeconds,
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: cmds,
			},
		},
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
				WithNodeSelector(nodeSelector).
				WithLabels(labelselector).
				WithContainerBuildersNew(
					k8stest.NewContainerBuilder().
						WithName(c.deployName).
						WithImage(common.GetFioImage()).
						WithImagePullPolicy(corev1.PullAlways).
						WithLivenessProbe(probe).
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
}

func (c *appConfig) verifyApplicationPodRunning(state bool) {
	labels := "e2e-test=" + c.deployName
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
