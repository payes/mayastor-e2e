package single_msn_shutdown

import (
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (c *shutdownConfig) nonCoreAgentNodeShutdownTest() {

	var conf *appConfig
	var errs common.ErrorAccumulator

	coreAgentNodeName, err := k8stest.GetCoreAgentNodeName()
	Expect(err).ToNot(HaveOccurred())
	Expect(coreAgentNodeName).ToNot(BeEmpty(), "core agent pod not found in running state")
	for _, deploy := range msDeployment {
		err = k8stest.ApplyNodeSelectorToDeployment(deploy, common.NSMayastor(), "kubernetes.io/hostname", coreAgentNodeName)
		if err != nil {
			errs.Accumulate(err)
		}
	}
	Expect(errs.GetError()).ToNot(HaveOccurred(), "Failed to  apply node selectors to deployment, %v", errs.GetError())

	err = k8stest.VerifyPodsOnNode([]string{"msp-operator", "rest", "csi-controller"}, coreAgentNodeName, common.NSMayastor())
	Expect(err).ToNot(HaveOccurred())

	verifyMayastorComponentStates(c.numMayastorInstances)

	// Create SC, PVC and Application Deployment
	for _, config := range c.config {
		config.createSC()
		config.uuid = config.createPVC()
		config.createDeployment()
		if config.nodeName != coreAgentNodeName {
			conf = config
		}
	}

	// Power off one non core agent node on which application is running
	Expect(c.platform.PowerOffNode(conf.nodeName)).ToNot(HaveOccurred(), "PowerOffNode")
	poweredOffNode = conf.nodeName
	logf.Log.Info("Sleeping for 2 mins... for all the mayastor pods to be in running status")
	time.Sleep(2 * time.Minute)
	verifyNodeNotReady(conf.nodeName, true)

	verifyMayastorComponentStates(c.numMayastorInstances - 1)

	for _, config := range c.config {
		if config.nodeName == conf.nodeName {
			continue
		}
		// Verify mayastor pods at the other nodes are still running
		Expect(getMsvState(config.uuid)).To(Equal(controlplane.VolStateDegraded()), "Unexpected MSV state")
		config.verifyApplicationPodRunning(true)
		config.verifyTaskCompletionStatus("success")
	}

	// Poweron the node for other tests to proceed
	Expect(c.platform.PowerOnNode(conf.nodeName)).ToNot(HaveOccurred(), "PowerOnNode")
	poweredOffNode = ""
	verifyNodesReady()
	err = k8stest.WaitForMCPPath(defWaitTimeout)
	Expect(err).ToNot(HaveOccurred())
	err = k8stest.WaitForMayastorSockets(k8stest.GetMayastorNodeIPAddresses(), defWaitTimeout)
	Expect(err).ToNot(HaveOccurred())
	// Delete deployment, PVC and SC
	for _, config := range c.config {
		config.deleteDeployment()
		config.deletePVC()
		config.deleteSC()
	}
	for _, deploy := range msDeployment {
		err := k8stest.RemoveAllNodeSelectorsFromDeployment(deploy, common.NSMayastor())
		if err != nil {
			errs.Accumulate(err)
		}
	}
	Expect(errs.GetError()).ToNot(HaveOccurred(), "Failed to  remove node selectors from deployment, %v", errs.GetError())
	verifyMayastorComponentStates(c.numMayastorInstances)
	err = k8stest.RestartMayastor(240, 240, 240)
	Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")
}

func (c *shutdownConfig) coreAgentNodeShutdownTest() {

	var conf *appConfig
	var errs common.ErrorAccumulator
	coreAgentNodeName, err := k8stest.GetCoreAgentNodeName()
	Expect(err).ToNot(HaveOccurred())
	Expect(coreAgentNodeName).ToNot(BeEmpty(), "core agent pod not found in running state")
	for _, deploy := range msDeployment {
		err = k8stest.ApplyNodeSelectorToDeployment(deploy, common.NSMayastor(), "kubernetes.io/hostname", coreAgentNodeName)
		if err != nil {
			errs.Accumulate(err)
		}
	}
	Expect(errs.GetError()).ToNot(HaveOccurred(), "Failed to  apply node selectors to deployment, %v", errs.GetError())

	err = k8stest.VerifyPodsOnNode([]string{"msp-operator", "rest", "csi-controller"}, coreAgentNodeName, common.NSMayastor())
	Expect(err).ToNot(HaveOccurred())

	verifyMayastorComponentStates(c.numMayastorInstances)
	for _, deploy := range msDeployment {
		err := k8stest.RemoveAllNodeSelectorsFromDeployment(deploy, common.NSMayastor())
		if err != nil {
			errs.Accumulate(err)
		}
	}
	Expect(errs.GetError()).ToNot(HaveOccurred(), "Failed to  remove node selectors from deployment, %v", errs.GetError())

	// Create SC, PVC and Application Deployment
	for _, config := range c.config {
		config.createSC()
		config.uuid = config.createPVC()
		config.createDeployment()
		if config.nodeName == coreAgentNodeName {
			conf = config
		}
	}

	// Power off coreAgent node
	Expect(c.platform.PowerOffNode(conf.nodeName)).ToNot(HaveOccurred(), "PowerOffNode")
	poweredOffNode = conf.nodeName
	logf.Log.Info("Sleeping for 2 mins... for IO paths to error out")
	time.Sleep(2 * time.Minute)
	verifyNodeNotReady(conf.nodeName, false)

	for _, config := range c.config {
		if config.nodeName == conf.nodeName {
			continue
		}
		// mayastor volume will not be marked as degraded unless core agent is up
		//Expect(getMsvState(config.uuid)).To(Equal(controlplane.VolStateDegraded()), "Unexpected MSV state")
		// Verify mayastor pods at the other nodes are still running
		config.verifyApplicationPodRunning(true)
		config.verifyTaskCompletionStatus("success")
	}

	// After 5 mins [(2(Earlier)+4(now)], core agent will be scheduled to some other node
	logf.Log.Info("Sleeping for 4 more mins... for core agent to be scheduled on a different node")
	time.Sleep(4 * time.Minute)
	err = k8stest.WaitForMCPPath(defWaitTimeout)
	Expect(err).ToNot(HaveOccurred())

	verifyMayastorComponentStates(c.numMayastorInstances - 1)
	for _, config := range c.config {
		if config.nodeName == conf.nodeName {
			continue
		}
		Expect(getMsvState(config.uuid)).To(Equal(controlplane.VolStateDegraded()), "Unexpected MSV state")
	}

	// Poweron the node for other tests to proceed
	Expect(c.platform.PowerOnNode(conf.nodeName)).ToNot(HaveOccurred(), "PowerOnNode")
	poweredOffNode = ""
	verifyNodesReady()

	err = k8stest.WaitForMCPPath(defWaitTimeout)
	Expect(err).ToNot(HaveOccurred())
	err = k8stest.WaitForMayastorSockets(k8stest.GetMayastorNodeIPAddresses(), defWaitTimeout)
	Expect(err).ToNot(HaveOccurred())

	// Delete deployment, PVC and SC
	for _, config := range c.config {
		config.deleteDeployment()
		config.deletePVC()
		config.deleteSC()
	}

	verifyMayastorComponentStates(c.numMayastorInstances)
	err = k8stest.RestartMayastor(240, 240, 240)
	Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")
}
