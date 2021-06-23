// JIRA: CAS-500
package pvc_stress_fio_concurrent

import (
	"fmt"
	"sync"
	"testing"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/resource"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defTimeoutSecs = "60s"

// Create Delete iterations
var cdIterations = e2e_config.GetConfig().PVCStress.CdCycles

// Create Read Update Delete iterations
var crudIterations = e2e_config.GetConfig().PVCStress.CrudCycles

// Replica count for the test
var replicaCount = e2e_config.GetConfig().PVCStress.Replicas

// Size of each volume in Mi
var volSizeMb = "10"

// VolumeMode: Filesystem
var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem

// Attempt to create n volumes simultaneously
func testPVCConcurrent(n int) {
	volSizeMbStr := fmt.Sprintf("%dMi", volSizeMb)

	// Create the storage class
	scName := "pvc-stress-test-concurrent-" + string(common.ShareProtoNvmf)
	err := k8stest.MkStorageClass("pvc-stress-test-concurrent-"+string(common.ShareProtoNvmf), replicaCount, common.ShareProtoNvmf, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	//	Create the specs of the list of volumes to be created
	var optsList = make([]coreV1.PersistentVolumeClaim, n)
	var errChannels = make([]chan error, n)

	// TODO: Ensure the presence of a pool of size larger than n * volSizeMb

	for i := 0; i < n; i++ {
		optsList[i] = coreV1.PersistentVolumeClaim{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", "vol", i),
				Namespace: common.NSDefault,
			},
			Spec: coreV1.PersistentVolumeClaimSpec{
				StorageClassName: &scName,
				AccessModes:      []coreV1.PersistentVolumeAccessMode{coreV1.ReadWriteOnce},
				Resources: coreV1.ResourceRequirements{
					Requests: coreV1.ResourceList{
						coreV1.ResourceStorage: resource.MustParse(volSizeMbStr),
					},
				},
				VolumeMode: &fileSystemVolumeMode,
			},
		}

		errChannels[i] = (make(chan error))
	}

	// Create the volumes
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i > n; i++ {
		go MkPVCMinimalHelper(&optsList[i], errChannels[i], wg)
	}

	// Wait for all the volumes to be created
	wg.Wait()

	// TODO: check that all volumes have been created successfully, that none of them are in the pending state,
	// that all of them have the right size, that all of them have the right storageClass, etc.....
}

func MkPVCMinimalHelper(createOpts *coreV1.PersistentVolumeClaim, ch chan error, wg sync.WaitGroup) {
	ch <- k8stest.MkPVCMinimal(createOpts)
	wg.Done()
}

// // Create a PVC and verify that (also see and keep in sync with README.md#pvc_stress_fio)
// //	1. The PVC status transitions to bound,
// //	2. The associated PV is created and its status transitions bound
// //	3. The associated MV is created and has a State "healthy"
// //  4. Optionally that a test application (fio) can read and write to the volume
// // then Delete the PVC and verify that
// //	1. The PVC is deleted
// //	2. The associated PV is deleted
// //  3. The associated MV is deleted
// func testPVC(volName string, protocol common.ShareProto, runFio bool) {
// 	logf.Log.Info("testPVC", "volume", volName, "protocol", protocol, "run FIO", runFio)
// 	scName := "pvc-stress-test-" + string(protocol)
// 	err := k8stest.MkStorageClass(scName, replicaCount, protocol, common.NSDefault)
// 	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

// 	_ = k8stest.MkPVC(64, volName, scName, common.VolFileSystem, common.NSDefault)

// 	if runFio {
// 		// Create the fio Pod
// 		fioPodName := "fio-" + volName
// 		pod, err := k8stest.CreateFioPod(fioPodName, volName, common.VolFileSystem, common.NSDefault)
// 		Expect(err).ToNot(HaveOccurred())
// 		Expect(pod).ToNot(BeNil())

// 		// Wait for the fio Pod to transition to running
// 		Eventually(func() bool {
// 			return k8stest.IsPodRunning(fioPodName, common.NSDefault)
// 		},
// 			defTimeoutSecs,
// 			"1s",
// 		).Should(Equal(true))

// 		// Run the fio test
// 		_, err = k8stest.RunFio(fioPodName, 5, common.FioFsFilename, common.DefaultFioSizeMb)
// 		Expect(err).ToNot(HaveOccurred())

// 		// Delete the fio pod
// 		err = k8stest.DeletePod(fioPodName, common.NSDefault)
// 		Expect(err).ToNot(HaveOccurred())
// 	}

// 	// Delete the PVC
// 	k8stest.RmPVC(volName, scName, common.NSDefault)

// 	// cleanup
// 	err = k8stest.RmStorageClass(scName)
// 	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
// }

// func stressTestPVC(iters int, runFio bool) {
// 	decoration := ""
// 	if runFio {
// 		decoration = "-io"
// 	}
// 	for ix := 1; ix <= iters; ix++ {
// 		// Sadly we cannot enumerate over enums so we have to explicitly invoke
// 		testPVC(fmt.Sprintf("stress-pvc-nvmf%s-%d", decoration, ix), common.ShareProtoNvmf, runFio)
// 		// FIXME: HACK disable iSCSI tests temporarily till Mayastor is fixed.
// 		//testPVC(fmt.Sprintf("stress-pvc-iscsi%s-%d", decoration, ix), common.ShareProtoIscsi, runFio)
// 	}
// }

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
		logf.Log.Info("Number of cycles are", "Create/Delete", cdIterations)
		stressTestPVC(cdIterations, false)
	})

	It("should stress test creation and deletion of PVCs provisioned over iSCSI and NVMe-of", func() {
		logf.Log.Info("Number of cycles are", "Create/Read/Update/Delete", crudIterations)
		stressTestPVC(crudIterations, true)
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()
	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
