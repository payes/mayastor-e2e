package pvc_readwriteonce

import (
	"fmt"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "60s"

// Create a PVC and verify that
//	1. The PVC status transitions to bound,
//	2. The associated PV is created and its status transitions bound
//	3. The associated MV is created and has a State "healthy"
//  4. Test application (fio) can read and write to the volume on two different node
// then Delete the PVC and verify that
//	1. The PVC is deleted
//	2. The associated PV is deleted
//  3. The associated MV is deleted
func testPVC(volName string, protocol common.ShareProto, runFio bool, fsType common.FileSystemType) {
	params := e2e_config.GetConfig().BasicVolumeIO
	logf.Log.Info("Test", "parameters", params)
	logf.Log.Info("testPVC", "volume", volName, "protocol", protocol, "run FIO", runFio, "fsType", fsType)
	scName := "pvc-rwo-test-" + string(protocol) + "-" + string(fsType)
	scObj, err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithProtocol(protocol).
		WithFileSystemType(fsType).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating storage class definition %s", scName)
	err = k8stest.CreateSc(scObj)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)
	// PVC create options
	createOpts := &coreV1.PersistentVolumeClaim{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      volName,
			Namespace: common.NSDefault,
		},
		Spec: coreV1.PersistentVolumeClaimSpec{
			StorageClassName: &scName,
			AccessModes:      []coreV1.PersistentVolumeAccessMode{coreV1.ReadWriteOnce},
			Resources: coreV1.ResourceRequirements{
				Requests: coreV1.ResourceList{
					coreV1.ResourceStorage: resource.MustParse("64Mi"),
				},
			},
		},
	}
	// Create the PVC.
	_, createErr := k8stest.CreatePVC(createOpts, common.NSDefault)
	Expect(createErr).To(BeNil())

	// Confirm the PVC has been created.
	pvc, getPvcErr := k8stest.GetPVC(volName, common.NSDefault)
	Expect(getPvcErr).To(BeNil())
	Expect(pvc).ToNot(BeNil())

	// Wait for the PVC to be bound.
	Eventually(func() coreV1.PersistentVolumeClaimPhase {
		return k8stest.GetPvcStatusPhase(volName, common.NSDefault)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(coreV1.ClaimBound))

	// Refresh the PVC contents, so that we can get the PV name.
	pvc, getPvcErr = k8stest.GetPVC(volName, common.NSDefault)
	Expect(getPvcErr).To(BeNil())
	Expect(pvc).ToNot(BeNil())

	// Wait for the PV to be provisioned
	Eventually(func() *coreV1.PersistentVolume {
		pv, getPvErr := k8stest.GetPV(pvc.Spec.VolumeName)
		if getPvErr != nil {
			return nil
		}
		return pv

	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Not(BeNil()))

	// Wait for the PV to be bound.
	Eventually(func() coreV1.PersistentVolumePhase {
		return k8stest.GetPvStatusPhase(pvc.Spec.VolumeName)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(coreV1.VolumeBound))

	// Wait for the MSV to be provisioned
	Eventually(func() *k8stest.MayastorVolStatus {
		return k8stest.GetMSV(string(pvc.ObjectMeta.UID))
	},
		defTimeoutSecs, //timeout
		"1s",           // polling interval
	).Should(Not(BeNil()))

	// Wait for the MSV to be healthy
	Eventually(func() string {
		return k8stest.GetMsvState(string(pvc.ObjectMeta.UID))
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal("healthy"))
	// list all nodes
	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())

	var workerNodes []string

	for _, node := range nodes {
		if node.MayastorNode && !node.MasterNode {
			workerNodes = append(workerNodes, node.NodeName)
		}
	}
	if runFio {
		// Create the fio Pod on first worker node
		fioPodFirstNodeName := "fio-" + volName + "-" + workerNodes[0]
		// fio pod labels
		label := map[string]string{
			"app": "fio",
		}
		// fio pod container
		firstPodContainer := coreV1.Container{
			Name:  fioPodFirstNodeName,
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

	}

	// Delete the PVC
	deleteErr := k8stest.DeletePVC(volName, common.NSDefault)
	Expect(deleteErr).To(BeNil())

	// Wait for the PVC to be deleted.
	Eventually(func() bool {
		return k8stest.IsPVCDeleted(volName, common.NSDefault)
	},
		"120s", // timeout
		"1s",   // polling interval
	).Should(Equal(true))

	// Wait for the PV to be deleted.
	Eventually(func() bool {
		return k8stest.IsPVDeleted(pvc.Spec.VolumeName)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(true))

	// Wait for the MSV to be deleted.
	Eventually(func() bool {
		return k8stest.IsMSVDeleted(string(pvc.ObjectMeta.UID))
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(true))

	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

func readwriteonceTestPVC(runFio bool, fsType common.FileSystemType) {
	decoration := ""
	if runFio {
		decoration = "-io"
	}

	// Sadly we cannot enumerate over enums so we have to explicitly invoke
	testPVC(fmt.Sprintf("rwo-pvc-nvmf%s", decoration), common.ShareProtoNvmf, runFio, fsType)
	testPVC(fmt.Sprintf("rwo-pvc-iscsi%s", decoration), common.ShareProtoIscsi, runFio, fsType)

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
		readwriteonceTestPVC(true, common.Ext4FsType)
	})
	It("should readwriteonce test creation and deletion of PVCs having xfs as fsType provisioned over iSCSI and NVMe-of", func() {
		readwriteonceTestPVC(true, common.XfsFsType)
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
