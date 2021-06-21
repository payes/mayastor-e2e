package msv_rebuild

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	v1alpha1Api "mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/common/k8stest"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestMayastorNode(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Mayastor rebuild test", "msv_rebuild")
}

var (
	podName      = "rebuild-test-fio"
	pvcName      = "rebuild-test-pvc"
	storageClass = "rebuild-test-nvmf"
)

func mayastorRebuildTest(protocol common.ShareProto) {
	err := k8stest.MkStorageClass(storageClass, common.DefaultReplicaCount, protocol, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", storageClass)
	// Create a PVC
	uuid := k8stest.MkPVC(common.DefaultVolumeSizeMb, pvcName, storageClass, common.VolFileSystem, common.NSDefault)
	timeout := "90s"
	pollPeriod := "1s"
	durationSecs := 250
	volumeFileSizeMb := 50
	podName = "fio-" + pvcName
	args := []string{
		"--",
		"--time_based",
		fmt.Sprintf("--runtime=%d", durationSecs),
		fmt.Sprintf("--filename=%s", common.FioFsFilename),
		fmt.Sprintf("--size=%dm", volumeFileSizeMb),
	}
	fioArgs := append(args, common.GetFioArgs()...)
	// fio pod container
	podContainer := coreV1.Container{
		Name:  podName,
		Image: "mayadata/e2e-fio",
		Args:  fioArgs,
	}
	// volume claim details
	volume := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
	}
	podObj, err := k8stest.NewPodBuilder().
		WithName(podName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(common.VolFileSystem).Build()
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Generating fio pod definition %s", podName))
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")
	// Create first fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Creating fio pod %s %v", podName, err))
	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(podName, common.NSDefault)
	},
		timeout,
		pollPeriod,
	).Should(Equal(true))
	replicas, err := custom_resources.GetMsVolReplicas(uuid)
	Expect(err).To(BeNil())
	Expect(len(replicas)).Should(Equal(1))
	// Wait for volume to be published before adding a child.
	// This ensures that a nexus exists when the child is added.
	Eventually(func() bool { return custom_resources.IsMsVolPublished(uuid) }, timeout, pollPeriod).Should(Equal(true))
	for i := 0; i < 2; i++ {
		// Add another child which should kick off a rebuild.
		err = custom_resources.UpdateMsVolReplicaCount(uuid, 2)
		Expect(err).ToNot(HaveOccurred(), "Update the number of replicas")
		replicas, err := custom_resources.GetMsVolReplicas(uuid)
		Expect(err).To(BeNil())
		Expect(len(replicas)).Should(Equal(2))
		// Wait for the added child to show up.
		time.Sleep(20 * time.Second)
		Eventually(func() int {
			msv, err := custom_resources.GetMsVol(uuid)
			if err == nil {
				return len(msv.Status.Nexus.Children)
			}
			return 0
		}, timeout, pollPeriod).Should(BeEquivalentTo(2))
		getChildrenFunc := func(uuid string) []v1alpha1Api.NexusChild {
			children, err := custom_resources.GetMsVolNexusChildren(uuid)
			if err != nil {
				panic("Failed to get children")
			}
			Expect(len(children)).Should(Equal(2))
			return children
		}
		// Check everything eventually goes healthy following a rebuild.
		Eventually(func() string { return getChildrenFunc(uuid)[0].State }, timeout, pollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
		Eventually(func() string { return getChildrenFunc(uuid)[1].State }, timeout, pollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
		Eventually(func() (string, error) { return custom_resources.GetMsVolNexusState(uuid) }, timeout, pollPeriod).Should(BeEquivalentTo("NEXUS_ONLINE"))

		// remove one child of nexus
		err = custom_resources.UpdateMsVolReplicaCount(uuid, 1)
		Expect(err).ToNot(HaveOccurred(), "Update the number of replicas")
		replicas, err = custom_resources.GetMsVolReplicas(uuid)
		Expect(err).To(BeNil())
		Expect(len(replicas)).Should(Equal(1))
		// Check everything remains in healthy state.
		Eventually(func() string { return getChildrenFunc(uuid)[0].State }, timeout, pollPeriod).Should(BeEquivalentTo("CHILD_ONLINE"))
		Eventually(func() (string, error) { return custom_resources.GetMsVolNexusState(uuid) }, timeout, pollPeriod).Should(BeEquivalentTo("NEXUS_ONLINE"))
	}
	logf.Log.Info("Waiting for run to complete", "timeout", durationSecs)
	tSecs := 0
	var phase coreV1.PodPhase
	for {
		if tSecs > durationSecs {
			break
		}
		time.Sleep(1 * time.Second)
		tSecs += 1
		phase, err = k8stest.CheckPodCompleted(podName, common.NSDefault)
		Expect(err).To(BeNil(), "CheckPodComplete got error %s", err)
		if phase != coreV1.PodRunning {
			break
		}
	}
	Expect(phase == coreV1.PodSucceeded).To(BeTrue(), "fio pod phase is %s", phase)
	logf.Log.Info("fio completed", "duration", tSecs)

	err = k8stest.DeletePod(podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Deleting fio pod")
	k8stest.RmPVC(pvcName, storageClass, common.NSDefault)
	err = k8stest.RmStorageClass(storageClass)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", storageClass)
}

var _ = Describe("Rebuild mayastor check", func() {
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
	It("Verify mayastor rebuild check", func() {
		mayastorRebuildTest(common.ShareProtoNvmf)
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
