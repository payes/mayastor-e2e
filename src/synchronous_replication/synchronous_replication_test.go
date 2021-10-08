package synchronous_replication

import (
	"fmt"
	"strings"
	"testing"
	"time"

	coreV1 "k8s.io/api/core/v1"

	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type srJobSpec struct {
	replicaCount int
	durationSecs int
}

type srJobStatus struct {
	scName  string
	volName string
	podName string
	volUid  string
}

type srJob struct {
	spec   srJobSpec
	status srJobStatus
}

// Start the job and populate all fields except replicaCount
func (job srJob) start(upDn string) srJob {
	protocol := common.ShareProtoNvmf
	volumeSizeMb := 512
	volumeFileSizeMb := 450
	volumeType := common.VolFileSystem

	logf.Log.Info("Parameters",
		"protocol", protocol, "volumeSizeMb", volumeSizeMb, "volumeType", volumeType,
		"volumeFileSizeMb", volumeFileSizeMb,
	)
	job.status.scName = strings.ToLower(
		fmt.Sprintf(
			"sync-replication-%s-%s-repl-%d",
			upDn,
			protocol,
			job.spec.replicaCount,
		),
	)
	err := k8stest.NewScBuilder().
		WithName(job.status.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(protocol).
		WithReplicas(job.spec.replicaCount).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Creating storage class %s", job.status.scName))

	job.status.volName = strings.ToLower(fmt.Sprintf("vol-%s", job.status.scName))

	// Create the volume
	job.status.volUid = k8stest.MkPVC(
		volumeSizeMb,
		job.status.volName,
		job.status.scName,
		volumeType,
		common.NSDefault,
	)
	logf.Log.Info("Volume created", "name", job.status.volName, "uid", job.status.volUid)

	// Create the fio Pod
	job.status.podName = "fio-" + job.status.volName
	args := []string{
		"--",
		"--time_based",
		fmt.Sprintf("--runtime=%d", job.spec.durationSecs),
		fmt.Sprintf("--filename=%s", common.FioFsFilename),
		fmt.Sprintf("--size=%dm", volumeFileSizeMb),
	}
	fioArgs := append(args, common.GetFioArgs()...)

	// fio pod container
	podContainer := coreV1.Container{
		Name:            job.status.podName,
		Image:           common.GetFioImage(),
		ImagePullPolicy: coreV1.PullAlways,
		Args:            fioArgs,
	}

	// volume claim details
	volume := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: job.status.volName,
			},
		},
	}

	podObj, err := k8stest.NewPodBuilder().
		WithName(job.status.podName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(common.VolFileSystem).Build()
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Generating fio pod definition %s", job.status.podName))
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

	// Create first fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Creating fio pod %s %v", job.status.podName, err))

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(job.status.podName, common.NSDefault)
	},
		"120s",
		"1s",
	).Should(Equal(true))
	return job
}

func (job srJob) stop() {
	// Delete the fio pod
	err := k8stest.DeletePod(job.status.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(job.status.volName, job.status.scName, common.NSDefault)

	err = k8stest.RmStorageClass(job.status.scName)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Deleting storage class %s", job.status.scName))
}

func (job srJob) updateVolumeCount(replicaCount int) {
	err := k8stest.SetMsvReplicaCount(job.status.volUid, replicaCount)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to update the volume replicas %s", job.status.volName))
}

func (job srJob) waitPodComplete() {
	const sleepTimeSecs = 5
	const timeoutSecs = 360
	var podPhase coreV1.PodPhase
	var err error

	logf.Log.Info("Waiting for pod to complete", "name", job.status.podName)
	for ix := 0; ix < (timeoutSecs+sleepTimeSecs-1)/sleepTimeSecs; ix++ {
		time.Sleep(sleepTimeSecs * time.Second)
		podPhase, err = k8stest.CheckPodCompleted(job.status.podName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to access pods status %s %v", job.status.podName, err))
		if podPhase == coreV1.PodSucceeded {
			return
		}
		Expect(podPhase == coreV1.PodRunning).To(BeTrue(), fmt.Sprintf("Unexpected pod phase %v", podPhase))
		volState, err := k8stest.GetMsvState(job.status.volUid)
		if err == nil {
			logf.Log.Info("MayastorVolume", "Name", job.status.volUid, "State", volState)
		}
	}
	Expect(podPhase == coreV1.PodSucceeded).To(BeTrue(), fmt.Sprintf("pod did not complete, phase %v", podPhase))
}

func (job srJob) waitVolumeHealthy() {
	const sleepTimeSecs = 5
	const timeoutSecs = 360
	var volState string
	var err error

	logf.Log.Info("Waiting for volume to be healthy", "name", job.status.volUid)
	for ix := 0; ix < (timeoutSecs+sleepTimeSecs-1)/sleepTimeSecs; ix++ {
		time.Sleep(sleepTimeSecs * time.Second)
		volState, err = k8stest.GetMsvState(job.status.volUid)
		if err == nil {
			if volState == controlplane.VolStateHealthy() {
				return
			}
			logf.Log.Info("MayastorVolume",
				"Status.State", volState,
				"wanted", controlplane.VolStateHealthy())
		}
	}
	Expect(volState == controlplane.VolStateHealthy()).To(BeTrue(), fmt.Sprintf("volume is not healthy %v", volState))
}

func (job srJob) assertExpectedReplicas(replicaCount int, failMsg string) {
	// FIXME: this works on the assumption that the only mayastor volume provisioned is for this test
	replicas, err := k8stest.ListReplicasInCluster()
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("failed to access the list of replicas %v", err))
	Expect(len(replicas)).To(BeIdenticalTo(replicaCount), fmt.Sprintf("%s %v", failMsg, replicas))
}

func TestSynchronousReplication(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Synchronous Replication tests", "synchronous_replication")
}

var _ = Describe("Synchronous Replication", func() {
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

	const delay = 30 * time.Second

	It("should increase the replicaCount count on a volume by 1 from 1 to 2, and application IO remains unaffected", func() {
		job := srJob{
			spec: srJobSpec{
				replicaCount: 1,
				durationSecs: 180,
			},
		}
		job = job.start("up")
		job.assertExpectedReplicas(job.spec.replicaCount, "unexpected number of replicas")

		logf.Log.Info("Sleeping ", "time", delay)
		time.Sleep(delay)

		updReplCount := job.spec.replicaCount + 1
		logf.Log.Info("Increasing replicaCount count to", "repl", updReplCount)
		job.updateVolumeCount(updReplCount)

		job.waitPodComplete()
		job.assertExpectedReplicas(updReplCount, "expected 2 replicas after increase")
		job.waitVolumeHealthy()

		job.stop()
	})

	It("should decrease the replicaCount count on a volume by 1 from 2 to 1, and the replicaCount is removed", func() {
		job := srJob{
			spec: srJobSpec{
				replicaCount: 2,
				durationSecs: 120,
			},
		}
		job = job.start("dn")
		job.assertExpectedReplicas(job.spec.replicaCount, "unexpected number of replicas")

		logf.Log.Info("Sleeping ", "time", delay)
		time.Sleep(delay)

		updReplCount := job.spec.replicaCount - 1
		logf.Log.Info("Decreasing replicaCount count to", "repl", updReplCount)
		job.updateVolumeCount(updReplCount)

		job.waitPodComplete()
		job.assertExpectedReplicas(updReplCount, "expected 1 replica after reduction")
		job.waitVolumeHealthy()
		job.stop()
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
