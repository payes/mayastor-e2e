package primitive_device_retirement

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/platform"

	client "mayastor-e2e/common/e2e-agent"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	sleepTime = 1
	timeout   = 900
)

func TestPrimitiveDeviceRetirement(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Primitive Device Retirement Test", "primitive_device_retirement")
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

var _ = Describe("Mayastor primitive device retirement tests", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {

		if len(rebootNode) != 0 {
			platform := platform.Create()

			// Reattaching the detached volume
			Expect(platform.AttachVolume("mayastor-"+rebootNode, rebootNode)).To(BeNil())

			// Node is being rebooted so that the volume comes back at the same location
			// as was there before detaching
			_ = platform.PowerOffNode(rebootNode)
			time.Sleep(30 * time.Second)
			_ = platform.PowerOnNode(rebootNode)
			rebootNode = ""
			restartMayastor = true
			time.Sleep(5 * time.Minute)
		}
		err := k8stest.RestartMayastor(240, 240, 240)
		Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods, err: %v", err)
		// Check resource leakage.
		err = k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})
	It("should verify primitive device retirement test (detached nexus)", func() {
		c := generatePrimitiveDeviceRetirementConfig("detach-nexus-volume")
		c.DetachCloudVolumeTest(true)
	})
	It("should verify primitive device retirement test (detached non-nexus)", func() {
		c := generatePrimitiveDeviceRetirementConfig("detach-non-nexus-volume")
		c.DetachCloudVolumeTest(false)
	})
	It("should verify primitive device retirement test (crashed nexus)", func() {
		c := generatePrimitiveDeviceRetirementConfig("crash-nexus-mayastor")
		c.CrashMayastorTest(true)
	})
	It("should verify primitive device retirement test (crashed non nexus)", func() {
		c := generatePrimitiveDeviceRetirementConfig("crash-non-nexus-mayastor")
		c.CrashMayastorTest(false)
	})
})

var (
	rebootNode      string
	restartMayastor bool
)

func (c *primitiveDeviceRetirementConfig) DetachCloudVolumeTest(testNexusNode bool) {
	var (
		err               error
		testNode, appNode string
	)

	c.createSC()
	uuid := c.createPVC()

	nodes, err := k8stest.GetMayastorNodeNames()
	Expect(err).ToNot(HaveOccurred())
	Expect(len(nodes) >= 2).To(BeTrue())
	if testNexusNode {
		appNode = nodes[0]
		testNode = nodes[0]
	} else {
		appNode = nodes[0]
		testNode = nodes[1]
	}

	// Create the fio Pod, The first time is just to write a verification pattern to the volume
	c.podName = "write-" + c.pvcName
	c.createFioPod(appNode, false)

	time.Sleep(10 * time.Second)
	Expect(c.platform.DetachVolume("mayastor-" + testNode)).To(BeNil())
	rebootNode = testNode

	Expect(k8stest.WaitPodComplete(c.podName, sleepTime, timeout)).To(BeNil())

	// Delete the fio pod
	err = k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Create a new fio Pod, This time just to verify the data.
	c.podName = "verify-" + c.pvcName
	c.createFioPod(appNode, true)

	Expect(k8stest.WaitPodComplete(c.podName, sleepTime, timeout)).To(BeNil())

	// Delete the fio pod
	err = k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	if testNexusNode {
		err = k8stest.SetMsvReplicaCount(uuid, c.replicas-1)
		Expect(err).ToNot(HaveOccurred(), "Failed to patch Mayastor volume %s", uuid)

		c.replicaIPs = c.GetReplicaAddressesForNonTestNodes(uuid, testNode)
		Expect(len(c.replicaIPs)).To(Equal(2))

		// Match data between both the healthy replicas
		c.PrimitiveDataIntegrity()
	}

	c.deletePVC()
	c.deleteSC()
}

// CrashMayastorTest crashes pod with nexus and verifies data consistency
func (c *primitiveDeviceRetirementConfig) CrashMayastorTest(testNexusNode bool) {
	var (
		err               error
		testNode, appNode string
	)

	c.createSC()
	uuid := c.createPVC()

	nodes, err := k8stest.GetMayastorNodeNames()
	Expect(err).ToNot(HaveOccurred())
	Expect(len(nodes) >= 2).To(BeTrue())
	if testNexusNode {
		appNode = nodes[0]
		testNode = nodes[0]
	} else {
		appNode = nodes[0]
		testNode = nodes[1]
	}

	err = k8stest.MsvConsistencyCheck(uuid)
	Expect(err).ToNot(HaveOccurred())

	// Create the fio Pod, The first time is just to write a verification pattern to the volume
	c.podName = "fio-write-" + c.pvcName
	c.createFioPod(appNode, false)

	time.Sleep(time.Second)

	// Kill Mayastor pod
	testNodeIP, err := k8stest.GetNodeIPAddress(testNode)
	Expect(err).ToNot(HaveOccurred())
	pid, err := client.Exec(*testNodeIP, "pidof mayastor")
	Expect(err).ToNot(HaveOccurred())
	Expect(pid).ToNot(Equal(""))
	restartMayastor = true
	killCmd := fmt.Sprintf("kill -9 %s", strings.TrimSuffix(string(pid), "\n"))
	_, err = client.Exec(*testNodeIP, killCmd)
	Expect(err).ToNot(HaveOccurred())

	// Wait for application pod to complete IOs
	time.Sleep(time.Minute)
	Expect(k8stest.WaitPodComplete(c.podName, sleepTime, timeout)).To(BeNil())

	// Delete the fio pod
	err = k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	time.Sleep(time.Minute)
	// Create a new fio Pod, This time just to verify the data.
	c.podName = "fio-verify-" + c.pvcName
	c.createFioPod(appNode, true)

	// Wait for application pod to complete IOs
	Expect(k8stest.WaitPodComplete(c.podName, sleepTime, timeout)).To(BeNil())

	newNexusNode, _ := k8stest.GetMsvNodes(uuid)
	Expect(newNexusNode).NotTo(Equal(""))

	// Delete the fio pod
	err = k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	time.Sleep(time.Minute)

	if testNexusNode {
		err = k8stest.MsvConsistencyCheck(uuid)
		Expect(err).ToNot(HaveOccurred())

		// Non nexus replicas are selected
		c.replicaIPs = c.GetReplicaAddressesForNonTestNodes(uuid, testNode)
		Expect(len(c.replicaIPs)).To(Equal(2))

		// Match data between the healthy replicas
		c.PrimitiveDataIntegrity()
	}

	c.deletePVC()
	c.deleteSC()
}
