package ms_pool_delete

import (
	"fmt"
	"strings"
	"testing"

	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/crds"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
)

func TestPooldeletion(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Pool Deletion Test", "ms_pool_delete")
}

var defTimeoutSecs = "60s"

func pooldeletionTest(protocol common.ShareProto, volumeType common.VolumeType, fsType common.FileSystemType, mode storageV1.VolumeBindingMode) {

	params := e2e_config.GetConfig().BasicVolumeIO
	logf.Log.Info("Test", "parameters", params)
	scName := strings.ToLower(fmt.Sprintf("pool-deletion-%d-%s-%s", params.Replicas, string(protocol), volumeType))
	volName := strings.ToLower(fmt.Sprintf("pool-deletion-%d-%s-%s", params.Replicas, string(protocol), volumeType))

	// Create storage class obj
	scObj, err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithReplicas(params.Replicas).
		WithVolumeBindingMode(mode).
		WithProtocol(protocol).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating storage class definition %s", scName)
	if fsType != "" {
		scObj.Parameters[string(common.ScFsType)] = string(fsType)
	}
	// Create storage class
	err = k8stest.CreateSc(scObj)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	// Create the volume
	uid := k8stest.MkPVC(
		params.VolSizeMb,
		volName,
		scName,
		volumeType,
		common.NSDefault,
	)

	// Confirm the PVC has been created.
	pvc, getPvcErr := k8stest.GetPVC(volName, common.NSDefault)
	Expect(getPvcErr).To(BeNil(), "PVC creation failed")
	Expect(pvc).ToNot(BeNil(), "PVC creation failed")

	// Wait for the PVC to be bound.
	Eventually(func() coreV1.PersistentVolumeClaimPhase {
		return k8stest.GetPvcStatusPhase(volName, common.NSDefault)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(coreV1.ClaimBound))

	// Refresh the PVC contents, so that we can get the PV name.
	pvc, getPvcErr = k8stest.GetPVC(volName, common.NSDefault)
	Expect(getPvcErr).To(BeNil(), "PVC content is nil")
	Expect(pvc).ToNot(BeNil(), "PVC content is nil")

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

	// Create pod
	fioPodName := "fio-" + volName
	pod, err := k8stest.CreateFioPod(fioPodName, volName, common.VolFileSystem, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())
	Expect(pod).ToNot(BeNil())

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

	// Get pool name from mayastorvolume
	replicas, err := k8stest.GetReplicas(uid)
	Expect(err).ToNot(HaveOccurred(), "Failed to get pool name")

	var poolName string
	for _, replica := range replicas {
		poolName = replica.Pool
		break
	}

	// Delete pool
	err = crds.DeletePool(poolName)
	Expect(err).ToNot(HaveOccurred())

	// Get pool
	pool, err := crds.GetPool(poolName)
	Expect(err).ToNot(HaveOccurred())
	Expect(pool).ToNot(BeNil())

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	// Delete the storage class
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)

	// RestoreConfiguredPools (re)create pools as defined by the configuration.
	// As part of the tests we may modify the pools, in such test cases
	// the test should delete all pools and recreate the configured set of pools.
	err = k8stest.RestoreConfiguredPools()
	Expect(err).To(BeNil(), "Not all pools are online after restoration")

}

var _ = Describe("Pool deletion check test", func() {

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

	It("Should verify mayastorpool deletion ", func() {
		pooldeletionTest(common.ShareProtoNvmf, common.VolFileSystem, common.XfsFsType, storageV1.VolumeBindingImmediate)
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
