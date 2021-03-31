package ms_pod_disruption

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defTimeoutSecs           = "90s"
	disconnectionTimeoutSecs = "180s"
	podUnscheduleTimeoutSecs = 100
	podRescheduleTimeoutSecs = 180
	repairTimeoutSecs        = "180s"
	mayastorRegexp           = "^mayastor-.....$"
	moacRegexp               = "^moac-..........-.....$"
	engineLabel              = "openebs.io/engine"
	mayastorLabel            = "mayastor"
)

type DisruptionEnv struct {
	replicaToRemove string
	unusedNodes     []string
	uuid            string
	volToDelete     string
	storageClass    string
	fioPodName      string
}

// prevent mayastor pod from running on the given node
func SuppressMayastorPodOn(nodeName string) {
	k8stest.UnlabelNode(nodeName, engineLabel)
	err := k8stest.WaitForPodNotRunningOnNode(mayastorRegexp, common.NSMayastor, nodeName, podUnscheduleTimeoutSecs)
	Expect(err).ToNot(HaveOccurred())
}

// allow mayastor pod to run on the given node
func UnsuppressMayastorPodOn(nodeName string) {
	// add the mayastor label to the node
	k8stest.LabelNode(nodeName, engineLabel, mayastorLabel)
	err := k8stest.WaitForPodRunningOnNode(mayastorRegexp, common.NSMayastor, nodeName, podRescheduleTimeoutSecs)
	Expect(err).ToNot(HaveOccurred())
}

// allow mayastor pod to run on the suppressed node
func (env *DisruptionEnv) UnsuppressMayastorPod() {
	if env.replicaToRemove != "" {
		UnsuppressMayastorPodOn(env.replicaToRemove)
		env.replicaToRemove = ""
	}
}

// return the node of the replica to remove,
// and a vector of the mayastor nodes not in the volume
func getNodes(uuid string) (string, []string) {
	nodeList, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())

	nexus, replicaNodes := k8stest.GetMsvNodes(uuid)
	Expect(nexus).NotTo(Equal(""))

	toRemove := ""
	// find a node which is not the nexus and is a replica
	for _, node := range replicaNodes {
		if node != nexus {
			toRemove = node
			break
		}
	}
	Expect(toRemove).NotTo(Equal(""))

	// get a list of all of the unused mayastor nodes in the cluster
	var unusedNodes []string
	for _, node := range nodeList {
		if node.MayastorNode {
			found := false
			for _, repnode := range replicaNodes {
				if repnode == node.NodeName {
					found = true
					break
				}
			}
			if !found {
				unusedNodes = append(unusedNodes, node.NodeName)
			}
		}
	}
	// we need at least 1 spare node to re-enable to allow the volume to become healthy again
	Expect(len(unusedNodes)).NotTo(Equal(0))
	logf.Log.Info("identified nodes", "nexus", nexus, "node of replica to remove", toRemove)
	return toRemove, unusedNodes
}

// Run fio against the cluster while a replica mayastor pod is unscheduled and then rescheduled
// TODO - run fio without a break
// TODO - disable the replica on the nexus node
func (env *DisruptionEnv) PodLossTest() {
	// Disable mayastor on the spare nodes so that moac initially cannot
	// assign one to the volume to replace the faulted one. We want to make
	// the volume degraded long enough for the test to detect it.
	for _, node := range env.unusedNodes {
		logf.Log.Info("suppressing mayastor on unused node", "node", node)
		SuppressMayastorPodOn(node)
	}
	logf.Log.Info("removing mayastor replica", "node", env.replicaToRemove)
	SuppressMayastorPodOn(env.replicaToRemove)

	logf.Log.Info("waiting for pod removal to affect the nexus", "timeout", disconnectionTimeoutSecs)
	Eventually(func() string {
		logf.Log.Info("running fio against the volume")
		_, err := k8stest.RunFio(env.fioPodName, 5, common.FioFsFilename, common.DefaultFioSizeMb)
		Expect(err).ToNot(HaveOccurred())
		return k8stest.GetMsvState(env.uuid)
	},
		disconnectionTimeoutSecs, // timeout
		"1s",                     // polling interval
	).Should(Equal("degraded"))

	logf.Log.Info("volume condition", "state", k8stest.GetMsvState(env.uuid))

	logf.Log.Info("running fio against the degraded volume")
	_, err := k8stest.RunFio(env.fioPodName, 20, common.FioFsFilename, common.DefaultFioSizeMb)
	Expect(err).ToNot(HaveOccurred())

	// re-enable mayastor on all unused nodes
	for _, node := range env.unusedNodes {
		logf.Log.Info("suppressing mayastor on unused node", "node", node)
		UnsuppressMayastorPodOn(node)
	}
	logf.Log.Info("waiting for the volume to be repaired", "timeout", repairTimeoutSecs)
	Eventually(func() string {
		logf.Log.Info("running fio while volume is being repaired")
		_, err := k8stest.RunFio(env.fioPodName, 5, common.FioFsFilename, common.DefaultFioSizeMb)
		Expect(err).ToNot(HaveOccurred())
		return k8stest.GetMsvState(env.uuid)
	},
		repairTimeoutSecs, // timeout
		"1s",              // polling interval
	).Should(Equal("healthy"))

	logf.Log.Info("volume condition", "state", k8stest.GetMsvState(env.uuid))

	logf.Log.Info("running fio against the repaired volume")
	_, err = k8stest.RunFio(env.fioPodName, 20, common.FioFsFilename, common.DefaultFioSizeMb)
	Expect(err).ToNot(HaveOccurred())

	logf.Log.Info("enabling mayastor pod", "node", env.replicaToRemove)
	UnsuppressMayastorPodOn(env.replicaToRemove)
}

// Common steps required when setting up the test.
// Creates the PVC, deploys fio, and records variables needed for the
// test in the DisruptionEnv structure
func Setup(pvcName string, storageClassName string, fioPodName string) DisruptionEnv {
	env := DisruptionEnv{}

	env.volToDelete = pvcName
	env.storageClass = storageClassName
	env.uuid = k8stest.MkPVC(common.DefaultVolumeSizeMb, pvcName, storageClassName, common.VolFileSystem, common.NSDefault)

	podObj := k8stest.CreateFioPodDef(fioPodName, pvcName, common.VolFileSystem, common.NSDefault)
	_, err := k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	env.fioPodName = fioPodName
	logf.Log.Info("waiting for pod", "name", env.fioPodName)
	Eventually(func() bool {
		return k8stest.IsPodRunning(env.fioPodName, common.NSDefault)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(true))

	env.replicaToRemove, env.unusedNodes = getNodes(env.uuid)
	return env
}

// Common steps required when tearing down the test
func (env *DisruptionEnv) Teardown() {
	var err error

	env.UnsuppressMayastorPod()

	for _, node := range env.unusedNodes {
		UnsuppressMayastorPodOn(node)
	}
	if env.fioPodName != "" {
		err = k8stest.DeletePod(env.fioPodName, common.NSDefault)
		env.fioPodName = ""
	}
	if env.volToDelete != "" {
		k8stest.RmPVC(env.volToDelete, env.storageClass, common.NSDefault)
		env.volToDelete = ""
	}
	Expect(err).ToNot(HaveOccurred())
}
