package primitive_device_retirement

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
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
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
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

			Expect(platform.AttachVolume("mayastor-"+rebootNode, rebootNode)).To(BeNil())
			_ = platform.PowerOffNode(rebootNode)
			time.Sleep(30 * time.Second)
			_ = platform.PowerOnNode(rebootNode)
			rebootNode = ""
			restartMayastor = true
			time.Sleep(2 * time.Minute)
		}
		if restartMayastor {
			err := k8stest.RestartMayastor(240, 240, 240)
			Expect(err).ToNot(HaveOccurred(), "Restart Mayastor pods")
		}
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})
	It("should verify primitive device retirement test", func() {
		c := generatePrimitiveDeviceRetirementConfig("primitive-device-retirement-detach-nexus-volume")
		c.DetachCloudVolumeTest()
	})
	It("should verify primitive device retirement test", func() {
		c := generatePrimitiveDeviceRetirementConfig("primitive-device-retirement-crash-nexus-mayastor")
		c.CrashMayastorTest()
	})
})

var (
	rebootNode      string
	restartMayastor bool
)

func (c *primitiveDeviceRetirementConfig) DetachCloudVolumeTest() {
	var err error

	c.createSC()
	uuid := c.createPVC()

	// Create the fio Pod, The first time is just to write a verification pattern to the volume
	c.podName = "fio-write-" + c.pvcName
	c.createFioPod(false)

	nexusNode, _ := k8stest.GetMsvNodes(uuid)
	Expect(nexusNode).NotTo(Equal(""))

	time.Sleep(10 * time.Second)
	Expect(c.platform.DetachVolume("mayastor-" + nexusNode)).To(BeNil())
	rebootNode = nexusNode

	Expect(k8stest.WaitPodComplete(c.podName, sleepTime, timeout)).To(BeNil())

	// Delete the fio pod
	err = k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Create a new fio Pod, This time just to verify the data.
	c.podName = "fio-verify-" + c.pvcName
	c.createFioPod(true)

	Expect(k8stest.WaitPodComplete(c.podName, sleepTime, timeout)).To(BeNil())

	// Delete the fio pod
	err = k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	msv, err := k8stest.GetMSV(uuid)
	Expect(err).ToNot(HaveOccurred())

	msv.Spec.ReplicaCount = c.replicas - 1
	_, err = custom_resources.UpdateMsVol(msv)
	Expect(err).ToNot(HaveOccurred(), "Failed to patch Mayastor volume %s", uuid)

	c.replicaIPs = c.GetReplicaAddressesForNonNexusNodes(uuid, nexusNode)
	Expect(len(c.replicaIPs)).To(Equal(2))

	// Match data between both the healthy replicas
	c.PrimitiveDataIntegrity()

	Expect(c.platform.AttachVolume("mayastor-"+rebootNode, rebootNode)).To(BeNil())
	c.deletePVC()
	c.deleteSC()
}

// CrashMayastorTest crashes pod with nexus and verifies data consistency
func (c *primitiveDeviceRetirementConfig) CrashMayastorTest() {
	var err error

	c.createSC()
	uuid := c.createPVC()

	// Create the fio Pod, The first time is just to write a verification pattern to the volume
	c.podName = "fio-write-" + c.pvcName
	c.createFioPod(false)

	err = k8stest.MsvConsistencyCheck(uuid)
	Expect(err).ToNot(HaveOccurred())

	nexusNode, _ := k8stest.GetMsvNodes(uuid)
	Expect(nexusNode).NotTo(Equal(""))

	testNode := nexusNode

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
	c.createFioPod(true)

	// Wait for application pod to complete IOs
	Expect(k8stest.WaitPodComplete(c.podName, sleepTime, timeout)).To(BeNil())

	newNexusNode, _ := k8stest.GetMsvNodes(uuid)
	Expect(newNexusNode).NotTo(Equal(""))

	// Delete the fio pod
	err = k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	time.Sleep(time.Minute)
	err = k8stest.MsvConsistencyCheck(uuid)
	Expect(err).ToNot(HaveOccurred())

	// Non nexus replicas are selected
	c.replicaIPs = c.GetReplicaAddressesForNonNexusNodes(uuid, newNexusNode)
	Expect(len(c.replicaIPs)).To(Equal(2))

	// Match data between the healthy replicas
	c.PrimitiveDataIntegrity()

	c.deletePVC()
	c.deleteSC()
}
