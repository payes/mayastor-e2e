// JIRA: CAS-500
package pvc_stress_fio_test

import (
	"fmt"
	"testing"

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

// Create Delete iterations
var cdIterations = 100

// Create Read Update Delete iterations
var crudIterations = 100

// volume name and associated storage class name
// parameters required by RmPVC
type volSc struct {
	volName string
	scName  string
}

var podNames []string
var volNames []volSc

// Create a PVC and verify that (also see and keep in sync with README.md#pvc_stress_fio)
//	1. The PVC status transitions to bound,
//	2. The associated PV is created and its status transitions bound
//	3. The associated MV is created and has a State "healthy"
//  4. Optionally that a test application (fio) can read and write to the volume
// then Delete the PVC and verify that
//	1. The PVC is deleted
//	2. The associated PV is deleted
//  3. The associated MV is deleted
func testPVC(volName string, protocol common.ShareProto, runFio bool) {
	logf.Log.Info("testPVC", "volume", volName, "protocol", protocol, "run FIO", runFio)
	scName := "pvc-stress-test-" + string(protocol)
	err := k8stest.MkStorageClass(scName, e2e_config.GetConfig().BasicVolumeIO.Replicas, protocol, common.NSDefault)
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

	// For cleanup
	tmp := volSc{volName, scName}
	volNames = append(volNames, tmp)

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

	if runFio {
		// Create the fio Pod
		fioPodName := "fio-" + volName
		pod, err := k8stest.CreateFioPod(fioPodName, volName, common.VolFileSystem, common.NSDefault)
		Expect(err).ToNot(HaveOccurred())
		Expect(pod).ToNot(BeNil())

		// For cleanup
		podNames = append(podNames, fioPodName)

		// Wait for the fio Pod to transition to running
		Eventually(func() bool {
			return k8stest.IsPodRunning(fioPodName, common.NSDefault)
		},
			defTimeoutSecs,
			"1s",
		).Should(Equal(true))

		// Run the fio test
		_, err = k8stest.RunFio(fioPodName, 5, common.FioFsFilename, common.DefaultFioSizeMb)
		Expect(err).ToNot(HaveOccurred())

		// Delete the fio pod
		err = k8stest.DeletePod(fioPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred())

		// cleanup
		podNames = podNames[:len(podNames)-1]
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

	// cleanup
	volNames = volNames[:len(volNames)-1]
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

func stressTestPVC(iters int, runFio bool) {
	decoration := ""
	if runFio {
		decoration = "-io"
	}
	for ix := 1; ix <= iters; ix++ {
		// Sadly we cannot enumerate over enums so we have to explicitly invoke
		testPVC(fmt.Sprintf("stress-pvc-nvmf%s-%d", decoration, ix), common.ShareProtoNvmf, runFio)
		testPVC(fmt.Sprintf("stress-pvc-iscsi%s-%d", decoration, ix), common.ShareProtoIscsi, runFio)
	}
}

func TestPVCStress(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "PVC Stress Test Suite", "pvc_stress_fio")
}

var _ = Describe("Mayastor PVC Stress test", func() {

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

	It("should stress test creation and deletion of PVCs provisioned over iSCSI and NVMe-of", func() {
		stressTestPVC(cdIterations, false)
	})

	It("should stress test creation and deletion of PVCs provisioned over iSCSI and NVMe-of", func() {
		stressTestPVC(crudIterations, true)
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()
	e2eCfg := e2e_config.GetConfig()
	cdIterations = e2eCfg.PVCStress.CdCycles
	crudIterations = e2eCfg.PVCStress.CrudCycles

	logf.Log.Info("Number of cycles are", "Create/Delete", cdIterations, "Create/Read/Update/Delete", crudIterations)

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
