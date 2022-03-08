package rc_reconciliation

import (
	"mayastor-e2e/common/controlplane"
	"testing"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestPrimitiveDeviceRetirement(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Replica Count Reconciliation", "rc_reconciliation")
}

var _ = BeforeSuite(func(done Done) {
	err := k8stest.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnv %v", err)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	err := k8stest.TeardownTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)
})

var _ = Describe("Mayastor replica count reconciliation tests", func() {

	BeforeEach(func() {
		// Check ready to run
		Expect(controlplane.MajorVersion()).To(Equal(1))
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})
	It("should verify replica count reconciliation test", func() {
		c := generateConfig("rc-reconciliation")
		c.RcReconciliationTest()
	})
})

func (c *Config) RcReconciliationTest() {
	var (
		err       error
		spareNode string
	)

	// Prepare list of nodes
	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())
	nodeList := []string{}
	for _, node := range nodes {
		if node.MayastorNode {
			nodeList = append(nodeList, node.NodeName)
		}
	}
	Expect(len(nodeList)).To(Equal(3))

	// Select the last node for spare pool
	spareNode = nodeList[2]

	// Delete pool from the spare node
	c.DeletePoolOnNode(spareNode)

	// Create PVC with replicas on node 1 and 2
	c.createSC()
	uuid := c.createPVC()

	// Verify replicas are created on node 1 and 2
	Expect(verifyReplicaOnNodes(uuid, []string{nodeList[0], nodeList[1]})).To(BeTrue())

	// Create spare pool to be faulted on Node 3
	CreateFaultyPoolOnNode(spareNode)

	// Create fio pod on node 1
	c.podName = "fio-write-" + c.pvcName
	c.createFioPod(nodeList[0])

	// Remove mayastor label from node 2
	suppressMayastorPodOn(nodeList[1], 0)

	// Verify that the volume has become degraded
	Eventually(func() string {
		volState, err := k8stest.GetMsvState(uuid)
		Expect(err).ToNot(HaveOccurred())
		logf.Log.Info("MSV", "state", volState, "expectedState", "Degraded")
		return volState
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal("Degraded"))

	// Verify a new replica is created on node 3, and msv now contains replicas of node 1 and node 3
	Expect(verifyReplicaOnNodes(uuid, []string{nodeList[0], nodeList[2]})).To(BeTrue())

	// Add mayastor label back on node 2
	unSuppressMayastorPodOn(nodeList[1], 0)

	// FIXME We should not suppress the mayastor pod on node 3 to bring the replica
	// up on node 2. Remove below line once CAS-1171 is fixed
	// Remove mayastor label from node 3 so that replicas start coming up on node 2
	suppressMayastorPodOn(nodeList[2], 0)

	// Verify that replica on node 3 is removed from msv and is reinstated on node 2
	Expect(verifyReplicaOnNodes(uuid, []string{nodeList[0], nodeList[1]})).To(BeTrue())

	// Add mayastor label back on node 3
	unSuppressMayastorPodOn(nodeList[2], 0)

	Eventually(func() string {
		volState, err := k8stest.GetMsvState(uuid)
		Expect(err).ToNot(HaveOccurred())
		logf.Log.Info("MSV", "state", volState, "expectedState", "Online")
		return volState
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal("Online"))

	// Delete the fio pod
	err = k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	c.deletePVC()
	c.deleteSC()

	// Delete faulty pool on Node 3
	DeleteFaultyPoolOnNode(spareNode)
	err = k8stest.RestoreConfiguredPools()
	Expect(err).ToNot(HaveOccurred())
}
