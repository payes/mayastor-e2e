package ms_pod_disruption_rm_vol

import (
	"fmt"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	mayastorGrpc "mayastor-e2e/common/mayastorclient/grpc"

	coreV1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defTimeoutSecs = "180s"
	mayastorRegexp = "^mayastor-.....$"
	engineLabel    = "openebs.io/engine"
	mayastorLabel  = "mayastor"
)

type DisruptionEnv struct {
	unusedNodes              []string
	uuid                     string
	volToDelete              string
	storageClass             string
	fioPodName               string
	podUnscheduleTimeoutSecs int
	podRescheduleTimeoutSecs int
}

var env DisruptionEnv

var fioParams = []string{
	"--name=benchtest",
	"--numjobs=1",
	"--direct=1",
	"--rw=randrw",
	"--ioengine=libaio",
	"--bs=4k",
	"--iodepth=16",
	"--verify=crc32",
}

func createFioPod(fioPodName string, volumeName string, volumeType common.VolumeType) {

	var args = []string{
		"--",
	}
	args = append(args, fmt.Sprintf("--filename=%s", common.FioBlockFilename))

	args = append(args, fioParams...)
	logf.Log.Info("fio", "arguments", args)

	// fio pod container
	podContainer := coreV1.Container{
		Name:            fioPodName,
		Image:           common.GetFioImage(),
		ImagePullPolicy: coreV1.PullAlways,
		Args:            args,
	}

	// volume claim details
	volume := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: volumeName,
			},
		},
	}

	podObj, err := k8stest.NewPodBuilder().
		WithName(fioPodName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(volumeType).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", fioPodName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", fioPodName)

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	logf.Log.Info("fio test pod is running.")
}

// Common steps required when tearing down the test
func (env *DisruptionEnv) teardown() {
	var err error

	for _, node := range env.unusedNodes {
		env.unsuppressMayastorPodOn(node, 0)
	}
	if env.fioPodName != "" {
		err := k8stest.DeletePod(env.fioPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env.fioPodName = ""
	}
	if env.volToDelete != "" {
		k8stest.RmPVC(env.volToDelete, env.storageClass, common.NSDefault)
		env.volToDelete = ""
	}
	if env.storageClass != "" {
		err = k8stest.RmStorageClass(env.storageClass)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env.storageClass = ""
	}
}

// Prevent mayastor pod from running on the given node.
// Action can be delayed to ensure overlap with IO in main thread.
func (env *DisruptionEnv) suppressMayastorPodOn(nodeName string, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	logf.Log.Info("suppressing mayastor pod", "node", nodeName)
	k8stest.UnlabelNode(nodeName, engineLabel)
	err := k8stest.WaitForPodNotRunningOnNode(mayastorRegexp, common.NSMayastor(), nodeName, env.podUnscheduleTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

// Allow mayastor pod to run on the given node.
// Action can be delayed to ensure overlap with IO in main thread.
func (env *DisruptionEnv) unsuppressMayastorPodOn(nodeName string, delay int) {
	// add the mayastor label to the node
	time.Sleep(time.Duration(delay) * time.Second)
	logf.Log.Info("restoring mayastor pod", "node", nodeName)
	k8stest.LabelNode(nodeName, engineLabel, mayastorLabel)
	err := k8stest.WaitForPodRunningOnNode(mayastorRegexp, common.NSMayastor(), nodeName, env.podRescheduleTimeoutSecs)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func listReplicasOnNode(nodeIP string) []mayastorclient.MayastorReplica {
	list, err := mayastorclient.ListReplicas([]string{nodeIP})
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	return list
}

func listPoolsOnNode(nodeIP string) []mayastorclient.MayastorPool {
	list, err := mayastorclient.ListPools([]string{nodeIP})
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	return list
}

func ReplicaLossVolumeDelete(pvcName string, storageClassName string, fioPodName string) {
	e2eCfg := e2e_config.GetConfig()
	env := DisruptionEnv{}
	env.podUnscheduleTimeoutSecs = e2eCfg.MsPodDisruption.PodUnscheduleTimeoutSecs
	env.podRescheduleTimeoutSecs = e2eCfg.MsPodDisruption.PodRescheduleTimeoutSecs

	scObj, err := k8stest.NewScBuilder().
		WithName(storageClassName).
		WithNamespace(common.NSDefault).
		WithProtocol(common.ShareProtoNvmf).
		WithReplicas(3).
		WithLocal(false).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating storage class definition %s", storageClassName)

	err = k8stest.CreateSc(scObj)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", storageClassName)
	env.storageClass = storageClassName

	volMb := common.DefaultVolumeSizeMb

	env.uuid = k8stest.MkPVC(volMb, pvcName, storageClassName, common.VolRawBlock, common.NSDefault)
	env.volToDelete = pvcName

	createFioPod(fioPodName, pvcName, common.VolRawBlock)
	env.fioPodName = fioPodName

	// get the nodes comprising the volume
	nexus, replicaNodes := k8stest.GetMsvNodes(env.uuid)
	logf.Log.Info("Nexus", "node", nexus)
	logf.Log.Info("Replica", "nodes", replicaNodes)
	Expect(nexus).NotTo(Equal(""), "Nexus not found")
	Expect(len(replicaNodes)).To(Equal(3), "Expected 3 replica nodes")

	// we need to delete the pod in order to remove the Volume
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	Eventually(func() bool {
		return k8stest.IsPodRunning(env.fioPodName, common.NSDefault)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(false))

	volList, err := k8stest.ListMsvs()
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(len(volList)).To(Equal(1), "Expected 1 volume")

	// get the IP addresses of the nodes
	nodeList, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(len(nodeList)).ToNot(BeZero())

	var nexusIP = ""
	var replicaIPs []string

	// the IP addresses are needed to examine the replicas via the mayastor pods
	for _, node := range nodeList {
		if node.NodeName == nexus {
			nexusIP = node.IPAddress
		} else {
			for _, replica := range replicaNodes {
				if node.NodeName == replica {
					replicaIPs = append(replicaIPs, node.IPAddress)
				}
			}
		}
	}
	Expect(len(replicaIPs)).To(Equal(2), "Could not find all replica IP addresses")
	Expect(nexusIP).NotTo(Equal(""), "Could not find nexus IP address")

	// for each node call listReplica()
	// verify that each has 1 replica and 1 online pool
	reps := listReplicasOnNode(nexusIP)
	Expect(len(reps)).To(Equal(1), "Expected 1 replica on the nexus node")
	pools := listPoolsOnNode(nexusIP)
	Expect(len(pools)).To(Equal(1), "Expected 1 pool on the nexus node")
	Expect(pools[0].State).To(Equal(mayastorGrpc.PoolState_POOL_ONLINE))

	for _, nodeIP := range replicaIPs {
		reps = listReplicasOnNode(nodeIP)
		Expect(len(reps)).To(Equal(1), "Expected 1 replica on each replica pod")
		pools = listPoolsOnNode(nodeIP)
		Expect(len(pools)).To(Equal(1), "Expected 1 pool on each node")
		Expect(pools[0].State).To(Equal(mayastorGrpc.PoolState_POOL_ONLINE))
	}

	// remove 2 replicas by un-scheduling mayastor
	// iterate through the replica nodes and suppress those not on the nexus
	for _, node := range replicaNodes {
		if node != nexus {
			logf.Log.Info("suppressing replica on", "name", node)
			env.suppressMayastorPodOn(node, 0)
			env.unusedNodes = append(env.unusedNodes, node)
		}
	}

	logf.Log.Info("deleting volume", "name", pvcName)
	k8stest.RmPVC(pvcName, storageClassName, common.NSDefault)

	// wait for the volume to go
	Eventually(func() int {
		volList, err := k8stest.ListMsvs()
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		logf.Log.Info("volume count is", "vols", len(volList))
		return len(volList)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(0))

	Eventually(func() bool {
		// the nexus pool should be online, the others offline
		msps, err := custom_resources.ListMsPools()
		logf.Log.Info("pools", "count", len(msps))
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		for _, msp := range msps {
			if msp.Spec.Node == nexus {
				if msp.Status.State != controlplane.MspStateOnline() {
					logf.Log.Info("pool on nexus node", "name", msp.Name, "state", msp.Status.State)
					return false
				}
			} else if msp.Status.State == controlplane.MspStateOnline() {
				logf.Log.Info("pool on non nexus node", "name", msp.Name, "state", msp.Status.State)
				return false
			}
		}
		return true
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(true))

	// wait for the nexus replica to go
	Eventually(func() int {
		reps := listReplicasOnNode(nexusIP)
		logf.Log.Info("nexus replica count", "reps", reps)
		return len(reps)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(0))

	// re-enable the other replica nodes so we can query them
	for _, node := range replicaNodes {
		if node != nexus {
			logf.Log.Info("unsuppressing replica on", "name", node)
			env.unsuppressMayastorPodOn(node, 0)
		}
	}

	// with the nexus call listReplica()
	// verify that the nexus has 0 replicas
	// verify that each other node has 1 replica
	reps = listReplicasOnNode(nexusIP)
	Expect(len(reps)).To(Equal(0), "Expected 0 replicas on the nexus pod")
	pools = listPoolsOnNode(nexusIP)
	Expect(len(pools)).To(Equal(1), "Expected 1 pool on the nexus")
	logf.Log.Info("nexus pool state is", "state", pools[0].State)
	Expect(pools[0].State).To(Equal(mayastorGrpc.PoolState_POOL_ONLINE), "Expected nexus pool to be online")

	if controlplane.MajorVersion() == 0 {
		for _, nodeIP := range replicaIPs {
			Eventually(func() error {
				reps, err = mayastorclient.ListReplicas([]string{nodeIP})
				return err
			},
				defTimeoutSecs, // timeout
				"1s",           // polling interval
			).Should(BeNil(), "Failed to list replica over gRPC")
			Expect(len(reps)).To(Equal(1), "Expected 1 replica on each replica pod")
		}

		// verify that the replicas do not get removed
		logf.Log.Info("wait for 30s before rechecking replicas")
		time.Sleep(time.Duration(30) * time.Second)

		for _, nodeIP := range replicaIPs {
			reps = listReplicasOnNode(nodeIP)
			Expect(len(reps)).To(Equal(1), "Expected 1 replica on each replica pod")
		}
	} else {
		logf.Log.Info("waiting for replicas to be garbage collected")
		const sleepTimeSecs = 30
		const timeoutSecs = 60 * 30 // 30 minutes
		var repCount = len(replicaIPs)
		for ix := 0; ix < timeoutSecs && repCount != 0; ix += sleepTimeSecs {
			repCount = 0
			time.Sleep(sleepTimeSecs * time.Second)
			for _, nodeIP := range replicaIPs {
				repCount += len(listReplicasOnNode(nodeIP))
			}
			logf.Log.Info("replicas", "count", repCount)
		}
		Expect(repCount).To(Equal(0), "Expected 0 replicas for deleted volume")
	}

	// the cluster is broken at this point so needs repairing
	// TODO - determine how much of this should be necessary
	k8stest.CleanUp()

	err = k8stest.RestoreConfiguredPools()
	Expect(err).To(BeNil(), "Not all pools are online after restoration")
}

func TestMayastorPodLoss(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-2330", "MQ-2330")
}

var _ = Describe("Mayastor replica pod removal test", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred(), "%v", err)
	})

	AfterEach(func() {
		env.teardown()

		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred(), "%v", err)
	})

	It("should verify nexus behaviour when removing an Volume with missing replicas", func() {
		ReplicaLossVolumeDelete("pvc-ms-pod-remove-test-remove-vol", "sc-ms-pod-remove-test-remove-vol", "fio-ms-pod-remove-test-remove-vol")
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
