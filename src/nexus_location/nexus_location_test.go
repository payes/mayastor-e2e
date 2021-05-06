package nexus_location

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const NlNodeSelectorKey = "e2e-nl"
const NlNodeSelectorAppValue = "e2e-app"

var nodeSelector = map[string]string{
	NlNodeSelectorKey: NlNodeSelectorAppValue,
}

const volSizeMb = 512
const volumeType = common.VolRawBlock
const ns = common.NSDefault

func TestNexusLocation(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Nexus location tests", "nexus_location")
}

func makeTestVolume(prefix string, replicas int, local bool) (string, string, string) {
	scName := fmt.Sprintf("%s-repl-%d-local-%v", prefix, replicas, local)
	scObj, err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(ns).
		WithProtocol(common.ShareProtoNvmf).
		WithReplicas(replicas).
		WithLocal(local).
		Build()
	Expect(err).ToNot(HaveOccurred(), "Generating storage class definition %s", scName)
	err = k8stest.CreateSc(scObj)
	Expect(err).ToNot(HaveOccurred(), "failed to create storage class %s", scName)
	volName := fmt.Sprintf("vol-%s", scName)
	uid := k8stest.MkPVC(volSizeMb, volName, scName, volumeType, ns)
	return volName, uid, scName
}

func destroyTestVolume(volName, scName string) {
	// Delete the volume
	k8stest.RmPVC(volName, scName, ns)

	err := k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

func nexusLocality(replicas int, local bool) {
	logf.Log.Info("nexus locality")

	volName, uid, scName := makeTestVolume("nexus-local", replicas, local)

	logf.Log.Info("Volume", "uid", uid)

	// Create the fio Pod
	fioPodName := "fio-" + volName
	podDef := k8stest.CreateFioPodDef(fioPodName, volName, volumeType, ns)
	Expect(podDef).ToNot(BeNil())
	pod, err := k8stest.CreatePod(podDef, ns)
	Expect(err).ToNot(HaveOccurred(), "create test pod")
	Expect(pod).ToNot(BeNil(), "create test pod")

	Expect(k8stest.WaitPodRunning(fioPodName, ns, 120)).To(BeTrue())
	logf.Log.Info("fio test pod is running.")

	podHostIP, err := k8stest.GetPodHostIp(fioPodName, ns)
	Expect(err).To(BeNil(), "failed to retrieve test pod Host IP addr")
	logf.Log.Info("Pod", "node", podHostIP)

	ips := []string{podHostIP}
	nexuses, err := mayastorclient.ListNexuses(ips)
	Expect(err).To(BeNil(), "failed to list nexuses")
	foundLocalNexus := false
	for _, nexus := range nexuses {
		if nexus.Uuid == uid {
			foundLocalNexus = true
			logf.Log.Info("found matching nexus local to consumer pod", "nexus uuid", nexus.Uuid, "nexus", nexus)
		}
	}
	Expect(foundLocalNexus).To(BeTrue(), "nexus is not local to consumer pod")

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, ns)
	Expect(err).ToNot(HaveOccurred(), "deleting test pod")

	destroyTestVolume(volName, scName)
}

func remotelyProvisionedVolume(replicas int, local bool) {
	deferredAssert := e2e_config.GetConfig().DeferredAssert

	logf.Log.Info("remote volume")
	nodes, err := k8stest.GetMayastorNodeNames()
	Expect(err).To(BeNil(), "failed to retrieve list of nodes")
	Expect(len(nodes)).ToNot(BeZero(), "no nodes found")
	nodeName := nodes[0]

	logf.Log.Info("De-scheduling mayastor on", "node", nodeName)
	k8stest.UnlabelNode(nodeName, common.MayastorEngineLabel)
	logf.Log.Info("Scheduling consumer pod on", "node", nodeName)
	k8stest.LabelNode(nodeName, NlNodeSelectorKey, NlNodeSelectorAppValue)

	volName, uid, scName := makeTestVolume("remote-nexus", replicas, local)
	logf.Log.Info("", "uid", uid)

	// Create the fio Pod
	fioPodName := "fio-" + volName
	podDef := k8stest.CreateFioPodDef(fioPodName, volName, volumeType, ns)
	Expect(podDef).ToNot(BeNil(), "failed to create pod definition")
	podDef.Spec.NodeSelector = nodeSelector
	pod, err := k8stest.CreatePod(podDef, ns)
	Expect(err).To(BeNil(), "failed to create test pod")
	Expect(pod).ToNot(BeNil(), "failed to create test pod")

	var podScheduledStatus coreV1.ConditionStatus
	var podScheduledReason string
	for ix := 0; ix < 6; ix++ {
		time.Sleep(10 * time.Second)
		podScheduledStatus, podScheduledReason, _ = k8stest.GetPodScheduledStatus(fioPodName, ns)
		if podScheduledStatus != coreV1.ConditionUnknown {
			break
		}
	}
	logf.Log.Info("FioPod", "name", fioPodName, "PodScheduledStatus", podScheduledStatus, "reason", podScheduledReason)
	if !deferredAssert {
		Expect(podScheduledStatus == coreV1.ConditionFalse).To(BeTrue(), "remotely provisioned pod was scheduled")
	}

	nexuses, err := k8stest.ListNexusesInCluster()
	if err == nil {
		logf.Log.Info("Nexuses", "list", nexuses)
	} else {
		logf.Log.Info("Failed to list nexuses")
	}

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, ns)
	Expect(err).ToNot(HaveOccurred(), "failed to delete test pod")

	destroyTestVolume(volName, scName)
	k8stest.UnlabelNode(nodeName, NlNodeSelectorKey)
	err = k8stest.EnsureNodeLabels()
	Expect(err).To(BeNil(), "Restoring node labels failed.")

	if deferredAssert {
		// Deferred check so that we clean up and can meaningfully run the next test, this renders postmortem analysis following this test next to useless
		Expect(podScheduledStatus == coreV1.ConditionFalse).To(BeTrue(), "remotely provisioned pod was scheduled")
	}
}

func descheduledTestPod(replicas int, local bool) {
	deferredAssert := e2e_config.GetConfig().DeferredAssert
	volName, uid, scName := makeTestVolume("desched", replicas, local)
	logf.Log.Info("", "uid", uid)

	// Create the fio Pod
	fioPodName := "fio-" + volName
	podDef := k8stest.CreateFioPodDef(fioPodName, volName, volumeType, ns)
	Expect(podDef).ToNot(BeNil(), "failed to create test pod definition")
	pod, err := k8stest.CreatePod(podDef, ns)
	Expect(err).To(BeNil(), "failed to create test pod")
	Expect(pod).ToNot(BeNil())

	// Wait for the fio Pod to transition to running
	Expect(k8stest.WaitPodRunning(fioPodName, ns, 120)).To(BeTrue())
	logf.Log.Info("fio test pod is running.")

	replicasAppRunning, err := k8stest.ListReplicasInCluster()
	Expect(err).To(BeNil(), "failed to list replicas")
	logf.Log.Info("", "replicas", replicasAppRunning)
	Expect(len(replicasAppRunning) == replicas).To(BeTrue(), "number of listed replicas does not match")

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, ns)
	Expect(err).ToNot(HaveOccurred(), "failed to delete test pod")
	for ix := 1; ix < 120; ix++ {
		if !k8stest.IsPodRunning(fioPodName, ns) {
			break
		}
		time.Sleep(1 * time.Second)
	}
	Expect(k8stest.IsPodRunning(fioPodName, ns)).To(BeFalse(), "de-scheduled fio pod is still running")
	replicasAppStopped, err := k8stest.ListReplicasInCluster()
	Expect(err).To(BeNil(), "failed to list replicas")
	logf.Log.Info("", "replicas", replicasAppStopped)

	// Compare the sets of replicas returned
	Expect(len(replicasAppRunning) == len(replicasAppStopped)).To(BeTrue(), "%v != %v", replicasAppRunning, replicasAppStopped)
	sort.Sort(mayastorclient.MayastorReplicaArray(replicasAppRunning))
	sort.Sort(mayastorclient.MayastorReplicaArray(replicasAppStopped))
	for ix := range replicasAppRunning {
		Expect(reflect.DeepEqual(replicasAppRunning[ix], replicasAppStopped[ix])).To(BeTrue(), "replicas do not match")
	}

	nexuses, err := k8stest.ListNexusesInCluster()
	Expect(err).To(BeNil(), "failed to list nexuses")
	logf.Log.Info("", "nexuses", nexuses)
	if !deferredAssert {
		Expect(len(nexuses)).To(BeZero(), "nexus was not destroyed when the consumer pods was de-scheduled")
	}

	destroyTestVolume(volName, scName)

	if deferredAssert {
		// Deferred check so that we clean up and can meaningfully run the next test, this renders postmortem analysis following this test next to useless
		Expect(len(nexuses)).To(BeZero(), "nexus was not destroyed when the consumer pods was de-scheduled")
	}
}

var _ = Describe("Nexus location tests", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should verify that when a consumer pod is scheduled on a Mayastor node, the nexus is located on the same node, single replica, local=true", func() {
		nexusLocality(1, true)
	})

	It("should verify that when a consumer pod is scheduled on a Mayastor node, the nexus is located on the same node, single replica, local=false", func() {
		nexusLocality(1, false)
	})

	It("should verify that when a consumer pod is scheduled on a Mayastor node, the nexus is located on the same node, 2 replicas, local=true", func() {
		nexusLocality(2, true)
	})

	It("should verify that when a consumer pod is scheduled on a Mayastor node, the nexus is located on the same node, 2 replicas, local=false", func() {
		nexusLocality(2, false)
	})

	It("should verify that when consumer pod is de-scheduled, nexus is destroyed but replicas remain, 1 replica, local=true", func() {
		descheduledTestPod(1, true)
	})

	It("should verify that when consumer pod is de-scheduled, nexus is destroyed but replicas remain, 1 replica, local=false", func() {
		descheduledTestPod(1, false)
	})

	It("should verify that when consumer pod is de-scheduled, nexus is destroyed but replicas remain, 2 replicas, local=true", func() {
		descheduledTestPod(1, true)
	})

	It("should verify that when consumer pod is de-scheduled, nexus is destroyed but replicas remain, 2 replicas, local=false", func() {
		descheduledTestPod(2, false)
	})
})

var _ = Describe("Nexus location tests", func() {
	// We run remotelyProvisionedVolume test in a separate Describe section because
	// the test manipulates node labels and we need to do extra work to restore the
	// cluster to usable state after the test has run, in case the test fails.

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		//Restore node labels, wait for all pools to transition to online.
		_ = k8stest.EnsureNodeLabels()
		err := k8stest.WaitForPoolsToBeOnline(120)
		Expect(err).ToNot(HaveOccurred(), "Not all pools are online")
		// Check resource leakage.
		err = k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should verify volume is not published if consumer pod is scheduled on a node not running Mayastor, 1 replica, local=true ", func() {
		remotelyProvisionedVolume(1, true)
	})

	It("should verify volume is not published if consumer pod is scheduled on a node not running Mayastor, 1 replica, local=false ", func() {
		remotelyProvisionedVolume(1, false)
	})

	It("should verify volume is not published if consumer pod is scheduled on a node not running Mayastor, 2 replicas, local=true ", func() {
		remotelyProvisionedVolume(2, true)
	})

	It("should verify volume is not published if consumer pod is scheduled on a node not running Mayastor, 1 replicas, local=false ", func() {
		remotelyProvisionedVolume(2, false)
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
