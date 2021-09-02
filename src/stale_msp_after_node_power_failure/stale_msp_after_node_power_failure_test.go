package stale_msp_after_node_power_failure

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/platform"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var poweredOffNode string

func TestStaleMspAfterNodePowerFailure(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-1981", "MQ-1981")
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

var _ = Describe("Stale MSP after node power failure test", func() {

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

		//Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Stale msp after node power failure test", func() {
		c := generateNodePowerFailureConfig("Stale MSP after node power failure test")
		c.staleMspAfterNodePowerFailureTest()
	})
})

func (c *nodepowerfailureConfig) staleMspAfterNodePowerFailureTest() {

	//Get node name on which moac pod is scheduled
	moacNodeName, err := k8stest.GetMoacNodeName()
	Expect(err).ToNot(HaveOccurred())
	Expect(moacNodeName).ToNot(BeEmpty(), fmt.Sprintf("moac pod not found in running state, %v", err))

	//List mayastor pools in the cluster
	pools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to list msp's in cluster, %v", err))

	var poolName, nodeName string
	var diskName []string

	//Select a test MSP from the cluster, which is not present on the
	//same node on which moac pod is scheduled
	for _, pool := range pools {
		if pool.Spec.Node != moacNodeName {
			poolName = pool.GetName()
			nodeName = pool.Spec.Node
			diskName = pool.Spec.Disks
			break
		}
	}

	//Power off the node on which test MSP is running
	poweredOffNode = nodeName
	Expect(c.platform.PowerOffNode(nodeName)).ToNot(HaveOccurred(), nodeName+" failed to powered off")
	time.Sleep(1 * time.Minute)

	//Verify that node is in not ready state
	c.verifyNodeNotReady(nodeName)

	Eventually(func() error {
		err = custom_resources.DeleteMsPool(poolName)
		Expect(err).ToNot(HaveOccurred())
		return nil
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(BeNil(), "Failed to delete test MSP "+poolName)

	// Power on the node
	Expect(c.platform.PowerOnNode(nodeName)).ToNot(HaveOccurred(), nodeName+" failed to powered on")
	poweredOffNode = ""
	time.Sleep(1 * time.Minute)

	// Create MSP with new name on same node and same disk
	var newPoolName = poolName + "-stale-msp"
	_, err = custom_resources.CreateMsPool(newPoolName, nodeName, diskName)
	Expect(err).ToNot(HaveOccurred())

	// Check for the pool status. It should not be in online state
	const timeSecs = 30
	const timeSleepSecs = 10
	for ix := 0; ix < timeSecs/timeSleepSecs; ix++ {
		time.Sleep(timeSleepSecs * time.Second)
		err = IsMsPoolOnline(newPoolName)
	}
	Expect(err).ToNot(BeNil(), "Unexpected: All pools are online, expected "+newPoolName+" to be offline")

	//Restart moac pod in mayastor namespace
	moacPod, _ := k8stest.GetMoacPodName()
	moacPodName := moacPod[0]

	err = k8stest.DeletePod(moacPodName, common.NSMayastor())
	Expect(err).ToNot(HaveOccurred())

	// check mayastor status
	ready, err := k8stest.MayastorReady(2, 540)
	Expect(err).ToNot(HaveOccurred())
	Expect(ready).To(BeTrue())

	// Check test MSP status, should have come in online state
	for ix := 0; ix < timeSecs/timeSleepSecs; ix++ {
		time.Sleep(timeSleepSecs * time.Second)
		err = IsMsPoolOnline(newPoolName)
	}
	Expect(err).To(BeNil(), newPoolName+" still not in online state, expected to be online after moac pod restart")

	err = k8stest.RestartMayastor(240, 240, 240)
	Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")
	c.verifyMayastorComponentStates()

}

// Check if a MSP is in online state
func IsMsPoolOnline(poolName string) error {
	poolHealthy := true
	pool, err := custom_resources.GetMsPool(poolName)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to get mayastor pool %s %v", poolName, err))

	if pool.Status.State != "online" {
		log.Log.Info("IsMsPoolOnline", "pool", poolName, "state", pool.Status.State)
		poolHealthy = false
	}

	if !poolHealthy {
		return fmt.Errorf(poolName + " is not online")
	}
	return err
}
