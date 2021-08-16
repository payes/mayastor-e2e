package pvc_waitforfirstconsumer

import (
	"fmt"
	"mayastor-e2e/common/custom_resources"
	"strings"
	"testing"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "90s"

func TestPvcWaitForFirstConsumer(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Check nexus and replica creation on PVC with volumeBindingMode: WaitForFirstConsumer", "pvc_waitforfirstconsumer")
}

func testPvcWaitForFirstConsumerTest(
	protocol common.ShareProto,
	volumeType common.VolumeType,
	mode storageV1.VolumeBindingMode,
	replica int) {

	scName := strings.ToLower(
		fmt.Sprintf(
			"pvc-waitforfirstconsumer-%d-%s",
			replica,
			string(protocol),
		),
	)
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithProtocol(protocol).
		WithReplicas(replica).
		WithVolumeBindingMode(mode).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)

	volName := strings.ToLower(
		fmt.Sprintf(
			"pvc-waitforfirstconsumer-%d-%s",
			replica,
			string(protocol),
		),
	)

	// Create the volume
	uid := k8stest.MkPVC(
		common.LargeClaimSizeMb,
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

	//check for MayastorVolume CR status
	msv, err := k8stest.GetMSV(uid)
	Expect(msv).To(BeNil())
	Expect(err).To(HaveOccurred(), "Get MSV succeeded, expected failure")

	//check PVC status i.e Pending
	Expect(pvc.Status.Phase).Should(Equal(coreV1.ClaimPending))

	//verify if nexus is created or not
	children, nexusErr := custom_resources.GetMsVolNexusChildren(uid)
	Expect(children).To(BeNil())
	Expect(nexusErr).ToNot(BeNil(), "Nexus children not created yet")

	// Create the fio pod name
	fioPodName := "fio-" + volName

	// fio pod container
	podContainer := coreV1.Container{
		Name:            fioPodName,
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

	err = k8stest.MsvConsistencyCheckAll(common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	//check for MayastorVolume CR status
	msv, err = k8stest.GetMSV(uid)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(msv).ToNot(BeNil())

	//check PVC status i.e Bound
	Expect(k8stest.GetPvcStatusPhase(volName, common.NSDefault)).Should(Equal(coreV1.ClaimBound))

	//verify if nexus children is created or not
	children, nexusErr = custom_resources.GetMsVolNexusChildren(uid)
	if nexusErr != nil {
		panic("Failed to get nexus children")
	}
	Expect(len(children)).Should(Equal(replica))

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(volName, scName, common.NSDefault)

	// Delete storageclass
	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

var _ = Describe("Check nexus and replica creation on PVC with volumeBindingMode: WaitForFirstConsumer test", func() {

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

	It("Check nexus and replica creation on PVC with volumeBindingMode: WaitForFirstConsumer", func() {
		testPvcWaitForFirstConsumerTest(common.ShareProtoNvmf, common.VolFileSystem, storageV1.VolumeBindingWaitForFirstConsumer, 2)
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
