package rc_reconciliation

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	agent "mayastor-e2e/common/e2e-agent"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"reflect"
	"time"

	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "240s"

var fioWriteParams = []string{
	"--name=benchtest",
	"--numjobs=1",
	"--direct=1",
	"--rw=randwrite",
	"--do_verify=0",
	"--ioengine=libaio",
	"--bs=512",
	"--iodepth=16",
	"--verify=crc32",
}

const (
	mayastorRegexp = "^mayastor-.....$"
	engineLabel    = "openebs.io/engine"
	mayastorLabel  = "mayastor"
)

func (c *Config) createSC() {

	err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithLocal(true).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.scName)
}

func (c *Config) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

func (c *Config) createPVC() string {
	// Create the volume with 1 replica
	pvc, err := k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, c.volType, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", c.pvcName)
	return pvc
}

func (c *Config) deletePVC() {
	err := k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", c.pvcName)
}

func (c *Config) createFioPod(node string) {

	var args = []string{
		"--",
	}
	args = append(args, fmt.Sprintf("--filename=%s", common.FioBlockFilename))

	args = append(args, fioWriteParams...)
	logf.Log.Info("fio", "arguments", args)

	// fio pod container
	container := k8stest.MakeFioContainer(c.podName, args)
	// volume claim details
	volume := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: c.pvcName,
			},
		},
	}

	podObj, err := k8stest.NewPodBuilder().
		WithName(c.podName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(container).
		WithNodeSelectorHostnameNew(node).
		WithVolume(volume).
		WithVolumeDeviceOrMount(c.volType).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", c.podName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", c.podName)

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(c.podName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	logf.Log.Info("fio test pod is running.")
}

func (c *Config) GetReplicaAddressesForNonNexusNodes(uuid, nexusNode string) []string {
	var addrs []string
	_, nodes := k8stest.GetMsvNodes(uuid)

	for _, node := range nodes {
		if node == nexusNode {
			continue
		}
		addr, err := k8stest.GetNodeIPAddress(node)
		Expect(err).To(BeNil())
		addrs = append(addrs, *addr)
	}
	return addrs
}

func (c *Config) DeletePoolOnNode(nodeName string) {
	pools, err := k8stest.ListMsPools()
	if err != nil {
		// This function may be called by AfterSuite by uninstall test so listing MSVs may fail correctly
		logf.Log.Info("DeletePoolOnNode: list MSPs failed.", "Error", err)
	}
	if err == nil && pools != nil && len(pools) != 0 {
		for _, pool := range pools {
			if pool.Spec.Node != nodeName {
				continue
			}
			logf.Log.Info("DeletePoolOnNode:", "pool", pool.Name)
			err = custom_resources.DeleteMsPool(pool.Name)
			if err != nil {
				logf.Log.Info("DeletePoolOnNode: failed to delete pool", "pool", pool.Name, "error", err)
			}
			break
		}
	}
}

func verifyReplicaOnNodes(uuid string, nodes []string) bool {
	Eventually(func() bool {
		_, msvNodes := k8stest.GetMsvNodes(uuid)
		logf.Log.Info("Replicas:", "msvNodes", msvNodes, "expectedNodes", nodes)
		return reflect.DeepEqual(nodes, msvNodes)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	return true
}

// Prevent mayastor pod from running on the given node.
// Action can be delayed to ensure overlap with IO in main thread.
func suppressMayastorPodOn(nodeName string, delay int) {
	time.Sleep(time.Duration(delay) * time.Second)
	logf.Log.Info("suppressing mayastor pod", "node", nodeName)
	err := k8stest.UnlabelNode(nodeName, engineLabel)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	err = k8stest.WaitForPodNotRunningOnNode(mayastorRegexp, common.NSMayastor(), nodeName, 60)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

// Allow mayastor pod to run on the given node.
// Action can be delayed to ensure overlap with IO in main thread.
func unSuppressMayastorPodOn(nodeName string, delay int) {
	// add the mayastor label to the node
	time.Sleep(time.Duration(delay) * time.Second)
	logf.Log.Info("restoring mayastor pod", "node", nodeName)
	err := k8stest.LabelNode(nodeName, engineLabel, mayastorLabel)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	err = k8stest.WaitForPodRunningOnNode(mayastorRegexp, common.NSMayastor(), nodeName, 60)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func createFaultyDisk(node string) {
	addr, err := k8stest.GetNodeIPAddress(node)
	Expect(err).ToNot(HaveOccurred())
	table := "0 50000 linear " + e2e_config.GetConfig().PoolDevice + " 0\n50000 5000000  error\n5050000 8143000 linear " + e2e_config.GetConfig().PoolDevice + " 5050000"
	err = agent.CreateFaultyDevice(*addr, e2e_config.GetConfig().PoolDevice, table)
	Expect(err).ToNot(HaveOccurred(), "Failed to create faulty disk on node %s: ", addr)
}

func deleteFaultyDisk(node string) {
	addr, err := k8stest.GetNodeIPAddress(node)
	Expect(err).ToNot(HaveOccurred())
	table := "dmsetup delete " + e2e_config.GetConfig().PoolDevice
	err = agent.CreateFaultyDevice(*addr, e2e_config.GetConfig().PoolDevice, table)
	Expect(err).ToNot(HaveOccurred(), "Failed to create faulty disk on node %s: ", addr)
}

func CreateFaultyPoolOnNode(spareNode string) {

	createFaultyDisk(spareNode)
	poolName := "spare-faulty-pool"
	logf.Log.Info("Creating msp", "poolName", poolName)
	_, err := custom_resources.CreateMsPool(poolName, spareNode, []string{"/dev/dm-0"})
	Expect(err).To(BeNil(), "Failed to create pool")
}

func DeleteFaultyPoolOnNode(spareNode string) {

	poolName := "spare-faulty-pool"
	logf.Log.Info("Creating msp", "poolName", poolName)
	err := custom_resources.DeleteMsPool(poolName)
	Expect(err).To(BeNil(), "Failed to delete pool")
	deleteFaultyDisk(spareNode)
}
