package etcd_inaccessibility

import (
	"testing"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegrityTest(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-2307", "MQ-2307")
}

func (c *inaccessibleEtcdTestConfig) etcdInaccessibilityTest() {

	c.createSC()
	c.createPVC()

	// Create pod
	c.podName = "write-" + c.pvcName
	c.createFioPod()

	//make etcd replicas zero
	logf.Log.Info("Scaling down etcd sts")
	var replicas int32 = 0
	k8stest.SetStatefulsetReplication("mayastor-etcd", e2e_config.GetConfig().Platform.MayastorNamespace, &replicas)
	WaitForEtcdState(false)

	// Wait for the fio Pod to complete
	Eventually(func() coreV1.PodPhase {
		phase, err := k8stest.CheckPodCompleted(c.podName, common.NSDefault)
		logf.Log.Info("CheckPodCompleted phase", "actual", phase, "desired", coreV1.PodSucceeded)
		if err != nil {
			return coreV1.PodUnknown
		}
		return phase
	},
		defTimeoutSecs,
		"5s",
	).Should(Equal(coreV1.PodSucceeded))

	logf.Log.Info("Scaling up etcd sts")
	replicas = 1
	k8stest.SetStatefulsetReplication("mayastor-etcd", e2e_config.GetConfig().Platform.MayastorNamespace, &replicas)
	WaitForEtcdState(true)

	k8stest.WaitForMCPPath(defTimeoutSecs)

	// Delete the fio pod
	err := k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	c.deletePVC()
	c.deleteSC()
}

func (c *inaccessibleEtcdTestConfig) etcdInaccessibilityWhenReplicaFaulted() {

	c.createSC()
	uuid := c.createPVC()

	// Create pod
	c.podName = "write-" + c.pvcName
	c.createFioPod()

	//make etcd replicas zero
	logf.Log.Info("Scaling down etcd sts")
	var replicas int32 = 0
	k8stest.SetStatefulsetReplication("mayastor-etcd", e2e_config.GetConfig().Platform.MayastorNamespace, &replicas)
	WaitForEtcdState(false)

	nodes := GetNonNexusNodes(uuid)

	DisablePoolDeviceAtNode(nodes[0], poolDevice)

	// Wait for the fio Pod to Fail
	Eventually(func() coreV1.PodPhase {
		phase, err := k8stest.CheckPodCompleted(c.podName, common.NSDefault)
		logf.Log.Info("CheckPodCompleted phase", "actual", phase, "desired", coreV1.PodFailed)
		if err != nil {
			return coreV1.PodUnknown
		}
		return phase
	},
		defTimeoutSecs,
		"5s",
	).Should(Equal(coreV1.PodFailed))

	EnablePoolDeviceAtNode(nodes[0], poolDevice)

	logf.Log.Info("Scaling up etcd sts")
	replicas = 1
	k8stest.SetStatefulsetReplication("mayastor-etcd", e2e_config.GetConfig().Platform.MayastorNamespace, &replicas)

	WaitForEtcdState(true)

	k8stest.WaitForMCPPath(defTimeoutSecs)

	// Delete the fio pod
	err := k8stest.DeletePod(c.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	c.deletePVC()
	c.deleteSC()

}

func WaitForEtcdState(ready bool) {
	Eventually(func() bool {
		status := k8stest.StatefulSetReady("mayastor-etcd", common.NSMayastor())
		logf.Log.Info("etcd Statefulset status", "actual", status, "desired", ready)
		return status
	},
		defTimeoutSecs,
		"5s",
	).Should(Equal(ready))
}

var _ = Describe("Mayastor Volume IO test", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {

		if detachedDeviceAtNode != "" {
			EnablePoolDeviceAtNode(detachedDeviceAtNode, poolDevice)
		}
		var replicas int32 = 1
		k8stest.SetStatefulsetReplication("mayastor-etcd", e2e_config.GetConfig().Platform.MayastorNamespace, &replicas)
		k8stest.WaitForMCPPath(defTimeoutSecs)
	})

	It("should verify data consistency when etcd is brought down", func() {
		c := generateInaccessibleEtcdTestConfig("etcd-inaccessibility")
		c.etcdInaccessibilityTest()
	})
	It("should verify write failures when etcd is brought down and one replica is faulted", func() {
		c := generateInaccessibleEtcdTestConfig("etcd-inaccessibility-faulted-replica")
		c.etcdInaccessibilityWhenReplicaFaulted()
	})

})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	k8stest.TeardownTestEnv()
})
