package stale_msp_after_node_power_failure

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/platform"

	"strings"
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
			err := k8stest.WaitForMCPPath(defWaitTimeout)
			Expect(err).ToNot(HaveOccurred())
			err = k8stest.WaitForMayastorSockets(k8stest.GetMayastorNodeIPAddresses(), defWaitTimeout)
			Expect(err).ToNot(HaveOccurred())
		}
		if controlplane.MajorVersion() == 1 {
			var errs common.ErrorAccumulator
			for _, deploy := range msDeployment {
				err := k8stest.RemoveAllNodeSelectorsFromDeployment(deploy, common.NSMayastor())
				if err != nil {
					errs.Accumulate(err)
				}
			}
			Expect(errs.GetError()).ToNot(HaveOccurred(), "Failed to  remove node selectors from deployment, %v", errs.GetError())
		} else {
			Expect(controlplane.MajorVersion).Should(Equal(1), "unsupported control plane version %d/n", controlplane.MajorVersion())
		}

		err := k8stest.RestartMayastor(240, 240, 240)
		Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")

		// RestoreConfiguredPools (re)create pools as defined by the configuration.
		// As part of the tests we may modify the pools, in such test cases
		// the test should delete all pools and recreate the configured set of pools.
		err = k8stest.RestoreConfiguredPools()
		Expect(err).To(BeNil(), "Not all pools are online after restoration")

		//Check resource leakage.
		err = k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Stale msp after node power failure test", func() {
		c := generateNodePowerFailureConfig("Stale MSP after node power failure test")
		c.staleMspAfterNodePowerFailureTest()
	})
})

func (c *nodepowerfailureConfig) staleMspAfterNodePowerFailureTest() {
	pools, err := k8stest.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to list msp's in cluster, %v", err))

	var poolName, nodeName string
	var diskName []string
	var errs common.ErrorAccumulator

	if controlplane.MajorVersion() == 1 {
		coreAgentNodeName, err := k8stest.GetCoreAgentNodeName()
		Expect(err).ToNot(HaveOccurred())
		Expect(coreAgentNodeName).ToNot(BeEmpty(), fmt.Sprintf("core agent pod not found in running state, %v", err))

		for _, deploy := range msDeployment {
			err = k8stest.ApplyNodeSelectorToDeployment(deploy, common.NSMayastor(), "kubernetes.io/hostname", coreAgentNodeName)
			if err != nil {
				errs.Accumulate(err)
			}
		}
		Expect(errs.GetError()).ToNot(HaveOccurred(), "Failed to  apply node selectors to deployment, %v", errs.GetError())
		err = k8stest.VerifyPodsOnNode([]string{"msp-operator", "rest", "csi-controller"}, coreAgentNodeName, common.NSMayastor())
		Expect(err).ToNot(HaveOccurred())
		err = k8stest.WaitForMCPPath(defWaitTimeout)
		Expect(err).ToNot(HaveOccurred())
		err = k8stest.WaitForMayastorSockets(k8stest.GetMayastorNodeIPAddresses(), defWaitTimeout)
		Expect(err).ToNot(HaveOccurred())
		ready, err := k8stest.MayastorReady(5, 60)
		Expect(err).ToNot(HaveOccurred(), "error check mayastor is ready after applying node selectors")
		Expect(ready).To(BeTrue(), "mayastor is not ready after applying node selectors")
		//Select a test MSP from the cluster
		for _, pool := range pools {
			if pool.Spec.Node != coreAgentNodeName {
				poolName = pool.Name
				nodeName = pool.Spec.Node
				diskName = pool.Spec.Disks
				break
			}
		}
	} else {
		Expect(controlplane.MajorVersion).Should(Equal(1), "unsupported control plane version %d/n", controlplane.MajorVersion())
	}

	//Power off the node on which test MSP is running
	poweredOffNode = nodeName
	Expect(c.platform.PowerOffNode(nodeName)).ToNot(HaveOccurred(), nodeName+" failed to power off")
	err = k8stest.WaitForMCPPath(defWaitTimeout)
	Expect(err).ToNot(HaveOccurred())
	//Verify that node is in not ready state
	verifyNodeNotReady(nodeName)

	Eventually(func() error {
		err = custom_resources.DeleteMsPool(poolName)
		Expect(err).ToNot(HaveOccurred())
		return nil
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(BeNil(), "Failed to delete test MSP "+poolName)

	// Power on the node
	Expect(c.platform.PowerOnNode(nodeName)).ToNot(HaveOccurred(), nodeName+" failed to power on")
	poweredOffNode = ""
	err = k8stest.WaitForMCPPath(defWaitTimeout)
	Expect(err).ToNot(HaveOccurred())
	err = k8stest.WaitForMayastorSockets(k8stest.GetMayastorNodeIPAddresses(), defWaitTimeout)
	Expect(err).ToNot(HaveOccurred())

	Eventually(func() bool {
		isPoolDeleted, err := IsMsPoolDeleted(poolName)
		if err == nil && isPoolDeleted {
			return true
		}
		log.Log.Info("IsMsPoolDeleted", "pool", poolName, "isPoolDeleted", isPoolDeleted)
		return false
	},
		defTimeoutSecs, // timeout
		"5s",           // polling interval
	).Should(BeTrue(), "Test pool is still present "+poolName)

	// Create MSP with new name on same node and same disk
	var newPoolName = poolName + "-stale-msp"
	_, err = custom_resources.CreateMsPool(newPoolName, nodeName, diskName)
	Expect(err).ToNot(HaveOccurred())

	log.Log.Info("Verify pool online", "pool", poolName)
	// Check for the pool status
	const timeSecs = 300
	const timeSleepSecs = 10
	for ix := 0; ix < timeSecs/timeSleepSecs; ix++ {
		time.Sleep(timeSleepSecs * time.Second)
		err = IsMsPoolOnline(newPoolName)
		if err == nil {
			break
		}
	}
	Expect(err).To(BeNil(), "Unexpected: All pools should be online, but "+newPoolName+" is not in online state")
}

// Check if a MSP is in online state
func IsMsPoolOnline(poolName string) error {
	poolHealthy := true
	pool, err := k8stest.GetMsPool(poolName)
	if err != nil {
		log.Log.Info("failed to get mayastor pool", "poolName", poolName, "err", err)
		return err
	}

	if strings.ToLower(pool.Status.State) != "online" {
		log.Log.Info("IsMsPoolOnline", "pool", poolName, "state", pool.Status.State)
		poolHealthy = false
	}

	if !poolHealthy {
		return fmt.Errorf(poolName + " is not online")
	}
	return err
}

// Check if a MSP is in online state
func IsMsPoolDeleted(poolName string) (bool, error) {
	pools, err := k8stest.ListMsPools()
	if err != nil {
		log.Log.Error(err, "failed to list mayastor pools")
		return false, err
	}
	for _, pool := range pools {
		if pool.Name == poolName {
			return false, nil
		}
	}

	return true, nil
}

// Verify that node is in not ready state
func verifyNodeNotReady(nodeName string) {
	Eventually(func() bool {
		readyStatus, err := k8stest.IsNodeReady(nodeName, nil)
		Expect(err).ToNot(HaveOccurred())
		return readyStatus
	},
		defTimeoutSecs, // timeout
		5,              // polling interval
	).Should(Equal(false))
}
