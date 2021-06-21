// JIRA: CAS-505
// JIRA: CAS-506
package node_failure

import (
	"fmt"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/platform"
	"mayastor-e2e/common/platform/types"

	"mayastor-e2e/common/custom_resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defTimeoutSecs   = 240  // in seconds
	durationSecs     = 600  // in seconds
	volumeFileSizeMb = 250  // in Mb
	thinkTime        = 1000 // in milliseconds
)

func TestNodeFailures(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Node Failure Tests", "node_failure")
}

func (c *failureConfig) createSC() {

	err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithFileSystemType(c.fsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.scName)
}

func (c *failureConfig) createPVC() string {
	// Create the volume with 1 replica
	return k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, common.VolFileSystem, common.NSDefault)
}

func (c *failureConfig) deletePVC() {
	k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
}

func (c *failureConfig) createDeployment() {

	labelselector := map[string]string{
		"e2e-test": "reboot",
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
						WithImagePullPolicy(corev1.PullAlways).
						WithVolumeMountsNew(volMounts).
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

func (c *failureConfig) deleteDeployment() {
	err := k8stest.DeleteDeployment(c.deployName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Deleting deployment %s", c.deployName)
	c.verifyApplicationPodRunning(false)
}

func (c *failureConfig) getNexusAndNonNexusNodes(uuid string) (string, []string) {
	nexusNode, replicaNodes := k8stest.GetMsvNodes(uuid)
	fmt.Printf("NexusNode: %v replicaNodes: %v\n", nexusNode, replicaNodes)
	Expect(nexusNode).NotTo(Equal(""))
	Expect(replicaNodes).NotTo(Equal(nil))
	nonNexusNodes := []string{}
	for _, node := range replicaNodes {
		if node != nexusNode {
			nonNexusNodes = append(nonNexusNodes, node)
		}
	}
	return nexusNode, nonNexusNodes
}

func (c *failureConfig) RebootDesiredNodes(uuid string) {
	nexusNode, nonNexusNodes := c.getNexusAndNonNexusNodes(uuid)

	switch c.testType {
	case RebootAllNodes:
		Expect(c.platform.PowerOffNode(nonNexusNodes[0])).ToNot(HaveOccurred(), "PowerOffNode")
		Expect(c.platform.PowerOffNode(nonNexusNodes[1])).ToNot(HaveOccurred(), "PowerOffNode")
		Expect(c.platform.PowerOffNode(nexusNode)).ToNot(HaveOccurred(), "PowerOffNode")

		time.Sleep(c.DownTime)
		c.verifyNodeNotReady(nonNexusNodes[0])
		c.verifyNodeNotReady(nonNexusNodes[1])
		c.verifyNodeNotReady(nexusNode)

		Expect(c.platform.PowerOnNode(nonNexusNodes[0])).ToNot(HaveOccurred(), "PowerOffNode")
		Expect(c.platform.PowerOnNode(nonNexusNodes[1])).ToNot(HaveOccurred(), "PowerOffNode")
		Expect(c.platform.PowerOnNode(nexusNode)).ToNot(HaveOccurred(), "PowerOffNode")

	case RebootOneNonNexusNode:

		Expect(c.platform.PowerOffNode(nonNexusNodes[0])).ToNot(HaveOccurred(), "PowerOffNode")

		time.Sleep(c.DownTime)
		c.verifyNodeNotReady(nonNexusNodes[0])

		Expect(c.platform.PowerOnNode(nonNexusNodes[0])).ToNot(HaveOccurred(), "PowerOffNode")

	case RebootTwoNonNexusNodes:
		Expect(c.platform.PowerOffNode(nonNexusNodes[0])).ToNot(HaveOccurred(), "PowerOffNode")
		Expect(c.platform.PowerOffNode(nonNexusNodes[1])).ToNot(HaveOccurred(), "PowerOffNode")

		time.Sleep(c.DownTime)
		c.verifyNodeNotReady(nonNexusNodes[0])
		c.verifyNodeNotReady(nonNexusNodes[1])

		Expect(c.platform.PowerOnNode(nonNexusNodes[0])).ToNot(HaveOccurred(), "PowerOffNode")
		Expect(c.platform.PowerOnNode(nonNexusNodes[1])).ToNot(HaveOccurred(), "PowerOffNode")

	case RebootNexusNode:
		Expect(c.platform.PowerOffNode(nexusNode)).ToNot(HaveOccurred(), "PowerOffNode")

		time.Sleep(c.DownTime)
		c.verifyNodeNotReady(nonNexusNodes[0])
		c.verifyNodeNotReady(nonNexusNodes[1])
		c.verifyNodeNotReady(nexusNode)

		Expect(c.platform.PowerOnNode(nexusNode)).ToNot(HaveOccurred(), "PowerOffNode")

	}
}

func (c *failureConfig) verifyMayastorComponentStates() {
	Eventually(func() bool {
		nodeList, err := custom_resources.ListMsNodes()
		Expect(err).ToNot(HaveOccurred(), "ListMsNodes")
		for _, node := range nodeList {
			if node.Status != "online" {
				return false
			}
		}
		ready, err := k8stest.MayastorReady(3, 540)
		Expect(err).ToNot(HaveOccurred())
		return ready
	}, defTimeoutSecs, 5,
	).Should(Equal(true))
}

func (c *failureConfig) verifyApplicationPodRunning(state bool) {
	labels := "e2e-test=reboot"
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

func (c *failureConfig) verifyNodesReady() {
	Eventually(func() bool {
		readyStatus, err := k8stest.AreNodesReady()
		Expect(err).ToNot(HaveOccurred())
		return readyStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(true))
}

func (c *failureConfig) verifyNodeNotReady(nodeName string) {

	Eventually(func() bool {
		readyStatus, err := k8stest.IsNodeReady(nodeName, nil)
		Expect(err).ToNot(HaveOccurred())
		return readyStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(false))

	/*
		// This check is not always possible as MOAC might be
		// running on the node which has been turned off
		Eventually(func() string {
			msn, err := custom_resources.GetMsNode(nodeName)
			Expect(err).ToNot(HaveOccurred(), "GetMsNode")
			return msn.Status
		},
			defTimeoutSecs, // timeout
			"5s",           // polling interval
		).Should(Equal("offline"))
	*/
}

func (c *failureConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

type failureConfig struct {
	protocol       common.ShareProto
	fsType         common.FileSystemType
	volType        common.VolumeType
	volBindingMode storageV1.VolumeBindingMode
	replicas       int
	testType       TestType
	DownTime       time.Duration
	scName         string
	pvcName        string
	pvcSize        int
	deployName     string
	podName        string
	nodeList       map[string]string
	platform       types.Platform
}

type TestType int

var nodes []string

const (
	RebootAllNodes TestType = iota
	RebootOneNonNexusNode
	RebootTwoNonNexusNodes
	RebootNexusNode
)

func generateFailureConfig(testType TestType, downTime time.Duration, testName string) *failureConfig {
	c := &failureConfig{
		protocol:       common.ShareProtoNvmf,
		volType:        common.VolFileSystem,
		fsType:         common.Ext4FsType,
		volBindingMode: storageV1.VolumeBindingImmediate,
		replicas:       3,
		DownTime:       downTime,
		testType:       testType,
		pvcSize:        5120, // In Mb
		scName:         testName + "-sc",
		pvcName:        testName + "-pvc",
		deployName:     testName + "-deploy",
		podName:        testName + "-pod",
		nodeList:       make(map[string]string),
	}

	nodeLocs, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), err)
	c.platform = platform.Create()
	Expect(c.platform).ToNot(BeNil())
	for _, node := range nodeLocs {
		c.nodeList[node.NodeName] = node.IPAddress
		nodes = append(nodes, node.NodeName)
	}
	return c
}

func (c *failureConfig) nodeRebootTests() {
	c.createSC()
	uuid := c.createPVC()
	c.createDeployment()
	c.RebootDesiredNodes(uuid)
	c.verifyNodesReady()

	c.verifyMayastorComponentStates()
	c.verifyApplicationPodRunning(true)

	c.deleteDeployment()
	c.deletePVC()
	c.deleteSC()
	err := k8stest.RestartMayastor(240, 240, 240)
	Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")
	c.verifyMayastorComponentStates()
}

var _ = Describe("Mayastor node failure tests", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		platform := platform.Create()
		for _, node := range nodes {
			_ = platform.PowerOnNode(node)
		}
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})
	It("should verify data integrity after one non nexus node is rebooted with 1 min downtime", func() {
		c := generateFailureConfig(RebootOneNonNexusNode, 1*time.Minute, "reboot-one-non-nexus-node-test-1min-downtime")
		c.nodeRebootTests()
	})
	It("should verify data integrity after one non nexus node is rebooted with 6 mins downtime", func() {
		c := generateFailureConfig(RebootOneNonNexusNode, 6*time.Minute, "reboot-one-non-nexus-node-test-6min-downtime")
		c.nodeRebootTests()
	})

	It("should verify data integrity after two non nexus nodes are rebooted with 1 min downtime", func() {
		c := generateFailureConfig(RebootTwoNonNexusNodes, 1*time.Minute, "reboot-two-non-nexus-nodes-test-1min-downtime")
		c.nodeRebootTests()
	})

	It("should verify data integrity after two non nexus nodes are rebooted with 6 mins downtime", func() {
		c := generateFailureConfig(RebootTwoNonNexusNodes, 6*time.Minute, "reboot-two-non-nexus-nodes-test-6min-downtime")
		c.nodeRebootTests()
	})

	It("should verify data integrity after all nodes rebooted with 1 min downtime", func() {
		c := generateFailureConfig(RebootAllNodes, 1*time.Minute, "reboot-all-nodes-test-1min-downtime")
		c.nodeRebootTests()
	})
	It("should verify data integrity after all nodes rebooted with 6 min downtime", func() {
		c := generateFailureConfig(RebootAllNodes, 6*time.Minute, "reboot-all-nodes-test-6min-downtime")
		c.nodeRebootTests()
	})

	It("should verify data integrity after nexus node is rebooted with 1 min downtime", func() {
		c := generateFailureConfig(RebootNexusNode, 1*time.Minute, "reboot-nexus-node-test-1min-downtime")
		c.nodeRebootTests()
	})
	It("should verify data integrity after nexus node is rebooted with 6 mins downtime", func() {
		c := generateFailureConfig(RebootNexusNode, 6*time.Minute, "reboot-nexus-node-test-6min-downtime")
		c.nodeRebootTests()
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()
	// err := k8stest.RestartMayastor(120, 120, 120)
	// Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")
	close(done)
}, 60)

var _ = AfterSuite(func() {

	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
