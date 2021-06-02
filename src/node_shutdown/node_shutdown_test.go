package node_shutdown

import (
	"mayastor-e2e/common/platform"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var poweredOffNode string

func TestNodeFailures(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Node Shutdown Tests", "node_shutdown")
}

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})

var _ = Describe("Mayastor node failure tests", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if len(poweredOffNode) != 0 {
			platform := platform.Create()
			_ = platform.PowerOnNode(poweredOffNode)
		}

		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})
	It("should verify node shutdown test", func() {
		c := generateShutdownConfig("node-shutdown")
		c.nodeShutdownTest()
	})
})

func (c *shutdownConfig) nodeShutdownTest() {
	// Create SC, PVC and Application Deployment
	c.createSC()
	uuid := c.createPVC()
	c.createDeployment()

	// Get the nexus node
	oldNexusNode, _ := k8stest.GetMsvNodes(uuid)
	Expect(oldNexusNode).NotTo(Equal(""))

	// Get the node on which the application pod is running
	labels := "e2e-test=shutdown"
	nodes, err := k8stest.GetNodeListForPods(labels, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "GetNodeListForPods")

	// Verify if the application pod is running on the nexus node
	match := false
	for node, status := range nodes {
		if node == oldNexusNode && status == v1.PodRunning {
			match = true
		}
	}
	Expect(match).To(Equal(true))

	// Power off nexus node on which application is running
	poweredOffNode = oldNexusNode
	Expect(c.platform.PowerOffNode(oldNexusNode)).ToNot(HaveOccurred(), "PowerOffNode")
	time.Sleep(6 * time.Minute)
	c.verifyNodeNotReady(oldNexusNode)

	// Verify mayastor pods at the other nodes are still running
	c.verifyMayastorComponentStates(c.numMayastorInstances - 1)

	// Verify the application comes back in running state on a different node
	c.verifyApplicationPodRunning(true)

	Eventually(func() bool {
		// Get the node on which the nexus has shifted
		newNexusNode, newReplicaNodes := k8stest.GetMsvNodes(uuid)
		if newNexusNode == "" {
			logf.Log.Info("newNexusNode", newNexusNode)
			return false
		}

		// Verify that msv has removed the powered off replica node from the list of replica nodes
		for _, node := range newReplicaNodes {
			if node == oldNexusNode {
				logf.Log.Info("msv still contains oldNexusNode", oldNexusNode)
				return false
			}
		}

		// Get the node on which application pod has moved
		labels := "e2e-test=shutdown"
		nodes, err := k8stest.GetNodeListForPods(labels, common.NSDefault)
		if err != nil {
			logf.Log.Info("Error in GetNodeListForPods", err)
			return false
		}

		// Verify if the application pod is running on new nexus node
		match := false
		for node, status := range nodes {
			if node == newNexusNode && status == v1.PodRunning {
				match = true
			}
		}
		Expect(match).To(Equal(true))
		return true
	},
		defTimeoutSecs,
		5,
	).Should(Equal(true))

	// Delete deployment, PVC and SC
	c.deleteDeployment()
	c.deletePVC()
	c.deleteSC()

	// Poweron the node for other tests to proceed
	Expect(c.platform.PowerOnNode(oldNexusNode)).ToNot(HaveOccurred(), "PowerOnNode")
	poweredOffNode = ""
	c.verifyMayastorComponentStates(c.numMayastorInstances)
	err = k8stest.RestartMayastor(240, 240, 240)
	Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")
}
