package pvc_readwriteonce

import (
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defTimeoutSecs = "60s" // in seconds
	pvcSize        = 1024  // In Mb
)

// Create a PVC and verify that
//	1. The PVC status transitions to bound,
//	2. The associated PV is created and its status transitions bound
//	3. The associated MV is created and has a State "healthy"
//  4. Test application (fio) can read and write to the volume on two different node
// then Delete the PVC and verify that
//	1. The PVC is deleted
//	2. The associated PV is deleted
//  3. The associated MV is deleted
func testPVC(volName string, protocol common.ShareProto, fsType common.FileSystemType) {
	params := e2e_config.GetConfig().PvcReadWriteOnce
	logf.Log.Info("Test", "parameters", params)
	logf.Log.Info("testPVC", "volume", volName, "protocol", protocol, "fsType", fsType)
	scName := "pvc-rwo-test-" + string(protocol) + "-" + string(fsType)
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithProtocol(protocol).
		WithFileSystemType(fsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	// create PVC
	uid := k8stest.MkPVC(pvcSize, volName, scName, common.VolFileSystem, common.NSDefault)
	Expect(uid).ToNot(BeNil(), "Failed to create PVC")

	// list all nodes
	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())

	var workerNodes []string

	for _, node := range nodes {
		if node.MayastorNode && !node.MasterNode {
			workerNodes = append(workerNodes, node.NodeName)
		}
	}
	// Create the fio Pod on first worker node
	fioPodFirstNodeName := "fio-" + volName + "-" + workerNodes[0]
	// fio pod labels
	label := map[string]string{
		"app": "fio",
	}
	// fio pod container
	firstPodContainer := coreV1.Container{
		Name:            fioPodFirstNodeName,
		Image:           common.GetFioImage(),
		ImagePullPolicy: coreV1.PullAlways,
		Args:            []string{"sleep", "1000000"},
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
		WithName(fioPodFirstNodeName).
		WithNamespace(common.NSDefault).
		WithNodeSelectorHostnameNew(workerNodes[0]).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(firstPodContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(common.VolFileSystem).
		WithLabels(label).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", fioPodFirstNodeName)
	Expect(podObj).ToNot(BeNil())
	// Create first fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", fioPodFirstNodeName)
	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodFirstNodeName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

	err = k8stest.MsvConsistencyCheckAll(common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// Run the fio test
	_, err = k8stest.RunFio(fioPodFirstNodeName, 5, common.FioFsFilename, common.DefaultFioSizeMb)
	Expect(err).ToNot(HaveOccurred())

	// Create the fio Pod on second worker node which will consume same pvc
	fioPodSecNodeName := "fio-" + volName + "-" + workerNodes[1]

	//Update pod name while creating second fio pod
	podObj.Name = fioPodSecNodeName

	// Update nodeselector for second node
	podObj.Spec.NodeSelector = map[string]string{
		k8stest.K8sNodeLabelKeyHostname: workerNodes[1],
	}

	// Create second fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", fioPodSecNodeName)
	logf.Log.Info("Checking fio pod status", "timeout", params.FioTimeout)
	tSecs := 0
	var phase coreV1.PodPhase
	for {
		if tSecs > params.FioTimeout {
			break
		}
		time.Sleep(1 * time.Second)
		tSecs += 1
		phase, err = k8stest.CheckPodCompleted(fioPodSecNodeName, common.NSDefault)
		Expect(err).To(BeNil(), "CheckPodComplete got error %s", err)
		if phase != coreV1.PodPending {
			break
		}
	}
	Expect(phase == coreV1.PodPending).To(BeTrue(), "fio pod phase is %s", phase)

	// Delete the fio pod scheduled on first node
	err = k8stest.DeletePod(fioPodFirstNodeName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the fio pod scheduled on second node
	err = k8stest.DeletePod(fioPodSecNodeName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the PVC
	k8stest.RmPVC(volName, scName, common.NSDefault)

	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

func readwriteonceTestPVC(fsType common.FileSystemType) {

	// Sadly we cannot enumerate over enums so we have to explicitly invoke
	testPVC("rwo-pvc-nvmf-io", common.ShareProtoNvmf, fsType)
	// FIXME: HACK disable iSCSI tests temporarily till Mayastor is fixed.
	//	testPVC(fmt.Sprintf("rwo-pvc-iscsi%s", decoration), common.ShareProtoIscsi, runFio, fsType)

}

func TestPVCReadWriteOnce(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "PVC ReadWriteOnce Test Suite", "pvc_readwriteonce")
}

var _ = Describe("Mayastor PVC ReadWriteOnce test", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Check resource leakage
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should readwriteonce test creation and deletion of PVCs having ext4 as fsType provisioned over iSCSI and NVMe-of", func() {
		readwriteonceTestPVC(common.Ext4FsType)
	})
	It("should readwriteonce test creation and deletion of PVCs having xfs as fsType provisioned over iSCSI and NVMe-of", func() {
		readwriteonceTestPVC(common.XfsFsType)
	})

})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

// Create Delete iterations
var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
