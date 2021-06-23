// JIRA: CAS-500
package pvc_stress_fio_concurrent

import (
	"context"
	"fmt"
	"mayastor-e2e/common/custom_resources/api/types/v1alpha1"
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

	custom_resources "mayastor-e2e/common/custom_resources"
)

// Multi-use timeout for the PVs to be provisioned, bound, and associated with MSVs
var defTimeoutSecs = 60

// Number of PVCs to create concurrently
var numPVCs = 100

// Replica count for the test
var replicaCount = e2e_config.GetConfig().PVCStress.Replicas

// Size of each volume in Mi
var volSizeMb = 10

// VolumeMode: Filesystem
var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem

// Attempt to create n volumes simultaneously
func testPVCConcurrent(n int, timeoutSecs int) {

	// Get the test environment from the k8stest package
	gTestEnv := k8stest.GetGTestEnv()

	volSizeMbStr := fmt.Sprintf("%dMi", volSizeMb)

	// Create the storage class
	scName := "pvc-stress-test-concurrent-" + string(common.ShareProtoNvmf)
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithReplicas(replicaCount).
		WithProtocol(common.ShareProtoNvmf).
		WithNamespace(common.NSDefault).
		WithVolumeBindingMode("Immediate").
		BuildAndCreate()
	Expect(err).To(BeNil(), "Creating storage class %s", scName)

	// Ensure that the cluster has only a single pool of size larger than n * volSizeMb
	poolsList, poolsListErr := custom_resources.ListMsPools()
	Expect(poolsListErr).To(BeNil())
	Expect(len(poolsList)).To(Equal(1))
	var poolCapacity int64 = poolsList[0].Status.Capacity // Pool capacity in bytes
	Expect(poolCapacity).Should(BeNumerically(">=", n*volSizeMb*1024*1024))
	logf.Log.Info("Number of pools currently present:", "count", len(poolsList))

	//	Create the descriptions of the volumes to be created
	var optsList = make([]coreV1.PersistentVolumeClaim, n)
	var pvcNames = make([]string, n)
	var errs = make([]error, n)

	for i := 0; i < n; i++ {
		pvcNames[i] = fmt.Sprintf("%s-%d", "vol", i)
		optsList[i] = coreV1.PersistentVolumeClaim{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      pvcNames[i],
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

		// errChannels[i] = make(chan error)
	}

	// Create the volumes
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		logf.Log.Info("Attempting to create PVC:", "pvcName", pvcNames[i])
		go MkPVCMinimal(&optsList[i], &errs[i], &wg, gTestEnv)
	}
	wg.Wait()

	logf.Log.Info("Finished calling the create methods for all PVC candidates.")

	// Check that all volumes have been created successfully, that none of them are in the pending state,
	// that all of them have the right size, that all of them have the right storageClass, tha.
	for i := 0; i < n; i++ {

		// Confirm that the PVC has been created
		Expect(errs[i]).To(BeNil())
		pvcApi := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims
		pvc, getPvcErr := pvcApi(common.NSDefault).Get(context.TODO(), pvcNames[i], metaV1.GetOptions{})
		Expect(getPvcErr).To(BeNil())
		Expect(pvc).ToNot(BeNil())
		// // Check that we can still get the storage class
		// ScApi := gTestEnv.KubeInt.StorageV1().StorageClasses
		// _, getScErr := ScApi().Get(context.TODO(), scName, metaV1.GetOptions{})
		// Expect(getScErr).To(BeNil())

		// Wait for the PVC to be bound.
		Eventually(func() coreV1.PersistentVolumeClaimPhase {
			return k8stest.GetPvcStatusPhase(pvcNames[i], common.NSDefault)
		},
			timeoutSecs, // timeout
			"1s",        // polling interval
		).Should(Equal(coreV1.ClaimBound))

		// Refresh the PVC contents, so that we can get the PV name.
		pvc, getPvcErr = pvcApi(common.NSDefault).Get(context.TODO(), pvcNames[i], metaV1.GetOptions{})
		Expect(getPvcErr).To(BeNil())
		Expect(pvc).ToNot(BeNil())

		// Wait for the PV to be provisioned
		Eventually(func() *coreV1.PersistentVolume {
			pv, getPvErr := gTestEnv.KubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), pvc.Spec.VolumeName, metaV1.GetOptions{})
			if getPvErr != nil {
				return nil
			}
			return pv

		},
			timeoutSecs, // timeout
			"1s",        // polling interval
		).Should(Not(BeNil()))

		// Wait for the PV to be bound.
		Eventually(func() coreV1.PersistentVolumePhase {
			return k8stest.GetPvStatusPhase(pvc.Spec.VolumeName)
		},
			timeoutSecs, // timeout
			"1s",        // polling interval
		).Should(Equal(coreV1.VolumeBound))

		// Check that an MSV is linked with the PVC
		Eventually(func() *v1alpha1.MayastorVolume {
			return k8stest.GetMSV(string(pvc.ObjectMeta.UID))
		},
			timeoutSecs,
			"1s",
			timeoutSecs, // timeout
			"1s",        // polling interval
		).Should(Not(BeNil()))

		logf.Log.Info("Created", "volume", pvc.Spec.VolumeName, "uuid", pvc.ObjectMeta.UID)
	}

}

func MkPVCMinimal(createOpts *coreV1.PersistentVolumeClaim, errBuf *error, wg *sync.WaitGroup, gTestEnv k8stest.TestEnvironment) {
	// Create the PVC.
	_, err := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(createOpts.ObjectMeta.Namespace).Create(context.TODO(), createOpts, metaV1.CreateOptions{})
	*errBuf = err
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
// 			timeout,
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

func TestPVCStressConcurrent(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "PVC Concurrent Stress Test Suite", "pvc_stress_concurrent")
}

var _ = Describe("Mayastor PVC Concurrent Stress test", func() {

	//  Check ready to run
	BeforeEach(func() {
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	// AfterEach(func() {
	// 	// Check resource leakage
	// 	err := k8stest.AfterEachCheck()
	// 	Expect(err).ToNot(HaveOccurred())
	// })

	It("should stress test concurrent creation (concurrent deletion is TODO) of PVCs provisioned over NVMe-of", func() {
		logf.Log.Info("Testing concurrent pvc creation:", "PVC count to create", fmt.Sprintf("%d", numPVCs))
		testPVCConcurrent(numPVCs, defTimeoutSecs)
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
