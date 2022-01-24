package pvc_delete

import (
	"fmt"
	"strings"
	"testing"

	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	grpc "mayastor-e2e/common/mayastorclient/protobuf"

	coreV1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	mayastorRegexp = "^mayastor-.....$"
	defTimeoutSecs = "120s"
)

func TestPvcDelete(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Delete PVC when some pools are unavailable", "pvc_delete")
}

func testPvcDeleteTest(
	protocol common.ShareProto,
	volumeType common.VolumeType,
	mode storageV1.VolumeBindingMode,
	replica int) {
	params := e2e_config.GetConfig().PvcDelete
	logf.Log.Info("Test", "parameters", params)
	scName := strings.ToLower(
		fmt.Sprintf(
			"pvc-delete-%d-%s",
			replica,
			string(protocol),
		),
	)
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithProtocol(protocol).
		WithReplicas(replica).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	volName := strings.ToLower(
		fmt.Sprintf(
			"pvc-delete-%d-%s",
			replica,
			string(protocol),
		),
	)

	// Create the volume
	uid := k8stest.MkPVC(
		params.VolSizeMb,
		volName,
		scName,
		volumeType,
		common.NSDefault,
	)
	logf.Log.Info("Volume", "uid", uid)

	// create fio pod
	fioPodName, err := createFioPod(volName)
	Expect(err).ToNot(HaveOccurred(), "failed to create fio pod")
	Expect(fioPodName).ToNot(BeNil(), "failed to get fio pod name")

	//check for MayastorVolume CR status
	msv, err := k8stest.GetMSV(uid)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(msv).ToNot(BeNil(), "failed to get msv")

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// get nexus node name
	nexusnode := getNexusNode(uid)

	// get non nexus node name for which label will be removed
	nodeName := getNonNexusNode(nexusnode)

	// get mayastor node address i.e non nexus node address
	address, err := getMayastorNodeIpAddres(nodeName)
	Expect(err).ToNot(HaveOccurred(), "failed to get mayastor node address")

	// get mayastor node address i.e nexus node address
	nexusNodeaddress, err := getMayastorNodeIpAddres(nexusnode)
	Expect(err).ToNot(HaveOccurred(), "failed to get mayastor non nexus node address")

	// verify if pool is online
	status := verifyMayastorPoolStatus(grpc.PoolState_POOL_ONLINE, address)
	Expect(status).Should(Equal(true))

	// Prevent mayastor pod from running on the given node.
	suppressMayastorPodOn(nodeName, params.PodUnscheduleTimeoutSecs)

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	// verify that the nexus is removed from the node that previously had a nexus.
	nexusList, err := mayastorclient.ListNexuses(nexusNodeaddress)
	Expect(err).ToNot(HaveOccurred(), "failed to fetch nexus list")
	Expect(len(nexusList)).To(Equal(0), "Expected to find 0 nexus")

	// Allow mayastor pod to run on the given node.
	unsuppressMayastorPodOn(nodeName, params.PodRescheduleTimeoutSecs)

	// check mayastor status
	ready, err := k8stest.MayastorReady(2, 540)
	Expect(err).ToNot(HaveOccurred(), "failed to get mayastor readyness status")
	Expect(ready).To(BeTrue())

	// verify if pools are online on node where mayastor is rescheduled
	status = verifyMayastorPoolStatus(grpc.PoolState_POOL_ONLINE, address)
	Expect(status).Should(Equal(true))

	// verify old replica status
	if controlplane.MajorVersion() == 1 {
		// Expect orphaned replica to be garbage collected
		Eventually(func() int {
			replicas, err := k8stest.ListReplicasInCluster()
			Expect(err).ToNot(HaveOccurred(), "failed to list replicas in cluster")
			if replicas != nil {
				logf.Log.Info("Found", "replicas", replicas)
			}
			return len(replicas)
		},
			"600s",
			"15s",
		).Should(Equal(0), "replicas have not been garbage collected")
	} else if controlplane.MajorVersion() == 0 {
		status, err = verifyOldReplicas(uid)
		Expect(err).ToNot(HaveOccurred(), "failed to verify old replica status")
		Expect(status).Should(Equal(true))
	}

	// Create the volume to check orphaned replica behavior
	uidSec := k8stest.MkPVC(
		params.VolSizeMb,
		volName,
		scName,
		volumeType,
		common.NSDefault,
	)
	logf.Log.Info("Volume", "uid", uidSec)

	//check for MayastorVolume CR status
	msv, err = k8stest.GetMSV(uidSec)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(msv).ToNot(BeNil(), "failed to get msv")

	// verify orphaned replica status, it should not be part of any msv in case of MOAC
	if controlplane.MajorVersion() == 0 {
		status, err = verifyOrphanedReplicas(uidSec, replica)
		Expect(err).ToNot(HaveOccurred(), "failed to verify orphaned replica status")
		Expect(status).Should(Equal(true))
	}

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	// Delete storageclass
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)

	// Remove replicas from cluster if exist
	err = k8stest.RmReplicasInCluster()
	Expect(err).ToNot(HaveOccurred(), "failed to remove replica")

}

// getNonNexusNode return mayastor node where nexus is absent
func getNonNexusNode(name string) string {
	nodeList, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())

	var nodeName string

	for _, node := range nodeList {
		if node.MayastorNode && !node.MasterNode && node.NodeName != name {
			nodeName = node.NodeName
			break
		}

	}
	return nodeName
}

// getNexusNode return mayastor node hosting nexus
func getNexusNode(uid string) string {
	nexusNode, _ := k8stest.GetMsvNodes(uid)
	Expect(nexusNode).NotTo(Equal(""), "Nexus not found")
	return nexusNode

}

// getMayastorNodeIpAddres return ip addres of mayastor node
func getMayastorNodeIpAddres(nodeName string) ([]string, error) {
	nodeAddres := []string{}
	nodes, err := k8stest.GetNodeLocs()
	if err != nil {
		return nodeAddres, err
	}
	for _, node := range nodes {
		if node.NodeName == nodeName {
			nodeAddres = append(nodeAddres, node.IPAddress)
			break
		}
	}
	return nodeAddres, err
}

// verify old replica status
func verifyOldReplicas(uid string) (bool, error) {
	replicas, err := k8stest.ListReplicasInCluster()
	var status bool
	if err != nil {
		logf.Log.Info("failed to retrieve list of replicas")
		return status, err
	}
	for _, replica := range replicas {
		if replica.Uuid == uid {
			status = true
			break
		}
	}
	return status, nil
}

// verify orphaned replica status i.e it should not form part of any msv
func verifyOrphanedReplicas(uid string, replicaCount int) (bool, error) {
	replicas, err := k8stest.ListReplicasInCluster()
	var status bool
	count := 0
	if err != nil {
		logf.Log.Info("ResourceEachCheck: failed to retrieve list of replicas")
		return status, err
	}
	for _, replica := range replicas {
		if replica.Uuid == uid {
			count++
		}
	}
	if count == replicaCount {
		status = true
	}
	return status, nil
}

// verify mayastor pool status on node address
func verifyMayastorPoolStatus(poolStatus grpc.PoolState, address []string) bool {
	pools, err := mayastorclient.ListPools(address)
	var status bool
	if err == nil {
		for _, pool := range pools {

			if pool.State != poolStatus {
				status = false
				break
			} else {
				status = true
			}
		}
	} else {
		logf.Log.Info("pools", "error", err)
	}
	return status
}

// create fio pod
func createFioPod(volName string) (string, error) {
	// Create the fio Pod
	fioPodName := "fio-" + volName

	// fio pod container
	podContainer := coreV1.Container{
		Name:  fioPodName,
		Image: common.GetFioImage(),
		Args:  []string{"sleep", "1000000"},
	}

	// volume claim details
	volume := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: volName,
			},
		},
	}

	podObj, err := k8stest.NewPodBuilder().
		WithName(fioPodName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(common.VolFileSystem).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", fioPodName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")
	// Create fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", fioPodName)

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

	return fioPodName, err
}

// Prevent mayastor pod from running on the given node.
func suppressMayastorPodOn(nodeName string, timeout int) {
	logf.Log.Info("suppressing mayastor pod", "node", nodeName)
	k8stest.UnlabelNode(nodeName, common.MayastorEngineLabel)
	err := k8stest.WaitForPodNotRunningOnNode(mayastorRegexp, common.NSMayastor(), nodeName, timeout)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

// Allow mayastor pod to run on the given node.
func unsuppressMayastorPodOn(nodeName string, timeout int) {
	// add the mayastor label to the node
	logf.Log.Info("restoring mayastor pod", "node", nodeName)
	k8stest.LabelNode(nodeName, common.MayastorEngineLabel, common.MayastorEngineLabelValue)
	err := k8stest.WaitForPodRunningOnNode(mayastorRegexp, common.NSMayastor(), nodeName, timeout)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

var _ = Describe("Delete PVC when some pools are unavailable test", func() {

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

	It("Delete PVC when some pools are unavailable", func() {
		testPvcDeleteTest(common.ShareProtoNvmf, common.VolFileSystem, storageV1.VolumeBindingWaitForFirstConsumer, 3)
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
