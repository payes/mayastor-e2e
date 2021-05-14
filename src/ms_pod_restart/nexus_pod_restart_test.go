package ms_pod_restart

import (
	"fmt"
	"strings"
	"testing"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	coreV1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	storageV1 "k8s.io/api/storage/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "90s"

func TestMsPodRestart(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Restart mayastor pod hosting the nexus", "ms_pod_restart")
}

func testMsPodRestartTest(
	protocol common.ShareProto,
	volumeType common.VolumeType,
	mode storageV1.VolumeBindingMode,
	local bool,
	replica int) {
	params := e2e_config.GetConfig().BasicVolumeIO
	logf.Log.Info("Test", "parameters", params)
	scName := strings.ToLower(
		fmt.Sprintf(
			"ms-pod-restart-%d-%s-%s-%s",
			replica,
			string(protocol),
			volumeType,
			mode,
		),
	)
	scObj, err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithProtocol(protocol).
		WithReplicas(replica).
		WithLocal(local).Build()

	Expect(err).ToNot(HaveOccurred(), "Generating storage class definition %s", scName)

	err = k8stest.CreateSc(scObj)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	volName := strings.ToLower(
		fmt.Sprintf(
			"ms-pod-restart-%d-%s-%s-%s",
			replica,
			string(protocol),
			volumeType,
			mode,
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

	// Confirm the PVC has been created.
	pvc, getPvcErr := k8stest.GetPVC(volName, common.NSDefault)
	Expect(getPvcErr).To(BeNil())
	Expect(pvc).ToNot(BeNil(), "failed to get pvc")

	// Create the fio Pod
	fioPodName := "fio-" + volName

	// fio pod container
	podContainer := coreV1.Container{
		Name:  fioPodName,
		Image: "mayadata/e2e-fio",
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
	// Create first fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", fioPodName)
	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

	// get node name where nexus is hostted
	node := getNexusNode(uid)
	Expect(node).ToNot(BeNil(), "Failed to get nexus node name")

	//verify one replica is local to the nexus
	localRepl := verifyLocalReplica(uid, node, replica)
	Expect(localRepl).Should(Equal(true))
	// get mayastor pod name which needs to be restarted
	msPodName := getMayastorPodName(common.NSMayastor(), node)
	Expect(msPodName).ToNot(BeNil(), "failed to get mayastor pod name hosting nexus")

	//Restart mayastor pod hosting the nexus
	err = k8stest.DeletePod(msPodName, common.NSMayastor())
	Expect(err).ToNot(HaveOccurred())

	// check mayastor status
	ready, err := k8stest.MayastorReady(2, 540)
	Expect(err).ToNot(HaveOccurred())
	Expect(ready).To(BeTrue())

	//verify one replica is local to the nexus
	localRepl = verifyLocalReplica(uid, node, replica)
	Expect(localRepl).Should(Equal(true))

	//verify the remote replicas are children of the newly (re) created nexus
	children := verifyRemoteReplica(uid, node, replica)
	Expect(children).Should(Equal(true))

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

// getNexusNodes return the node where nexus is hosted
func getNexusNode(uuid string) string {

	nexus, replicaNodes := k8stest.GetMsvNodes(uuid)
	Expect(nexus).NotTo(Equal(""))

	nexusNode := ""
	// find a node which is the nexus
	for _, node := range replicaNodes {
		if node == nexus {
			nexusNode = node
			break
		}
	}
	Expect(nexusNode).NotTo(Equal(""))
	return nexusNode
}

// getMayastorPodName return the mayastor pod name where nexus is hosted
func getMayastorPodName(ns string, nodeName string) string {
	logf.Log.Info("CheckMsPodName")
	podList, _ := k8stest.ListPod(ns)
	msPodName := ""
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == nodeName &&
			pod.GenerateName == "mayastor-" {
			msPodName = pod.Name
			break
		}
	}
	Expect(msPodName).NotTo(Equal(""))
	return msPodName
}

// verifyLocalReplica return the true when one replica is local to the nexus
func verifyLocalReplica(uuid string, nexusNode string, replCount int) bool {
	logf.Log.Info("VerifyLocalReplica")
	replicas, err := k8stest.GetReplicas(uuid)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(replicas) == replCount).To(BeTrue(), "number of listed replicas does not match")
	var status bool
	for _, replica := range replicas {
		if replica.Node == nexusNode &&
			strings.HasPrefix(replica.URI, "bdev:///") {
			status = true
		}
	}
	return status
}

// verifyRemoteReplica the remote replicas are children of the newly (re) created nexus
func verifyRemoteReplica(uuid string, nexusNode string, replCount int) bool {

	replicas, err := k8stest.GetReplicas(uuid)
	Expect(err).ToNot(HaveOccurred())
	Expect(len(replicas) == replCount).To(BeTrue(), "number of listed replicas does not match")
	var status bool
	for _, replica := range replicas {
		if replica.Node == nexusNode && strings.HasPrefix(replica.URI, "bdev:///") {
			status = true
		} else if replica.Node != nexusNode &&
			strings.HasPrefix(replica.URI, "nvmf://") {
			status = true
		} else {
			status = false
			break
		}
	}
	return status
}

var _ = Describe("Restart mayastor pod hosting the nexus test", func() {

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

	It("should verify nexus recreation with whichever shared replica(s) remain available after mayastor pod restart hosting nexus", func() {
		testMsPodRestartTest(common.ShareProtoNvmf, common.VolFileSystem, storageV1.VolumeBindingImmediate, true, 3)
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
