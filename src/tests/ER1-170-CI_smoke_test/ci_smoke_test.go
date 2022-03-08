// JIRA: CAS-505
// JIRA: CAS-506
package ci_smoke_test

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "120s"

func CISmokeTest() {
	params := e2e_config.GetConfig().CISmokeTest
	log.Log.Info("Test", "parameters", params)
	scName := strings.ToLower(fmt.Sprintf("ci-smoke-test-repl-%d", params.ReplicaCount))
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithReplicas(params.ReplicaCount).
		WithProtocol(common.ShareProtoNvmf).
		WithNamespace(common.NSDefault).
		WithVolumeBindingMode(storageV1.VolumeBindingImmediate).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "failed to create storage class %s", scName)

	volName := strings.ToLower(fmt.Sprintf("ci-smoke-test-repl-%d", params.ReplicaCount))

	// Create the volume
	uid, err := k8stest.MkPVC(params.VolSizeMb, volName, scName, common.VolFileSystem, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", volName)
	log.Log.Info("Volume", "uid", uid)

	// Create the fio Pod
	fioPodName := "fio-" + volName
	pod := k8stest.CreateFioPodDef(fioPodName, volName, common.VolFileSystem, common.NSDefault)
	Expect(pod).ToNot(BeNil())

	var args = []string{
		"--",
	}
	args = append(args, fmt.Sprintf("--filename=%s", common.FioFsFilename))
	args = append(args, fmt.Sprintf("--size=%dm", params.FsVolSizeMb))
	args = append(args, common.GetFioArgs()...)
	log.Log.Info("fio", "arguments", args)
	pod.Spec.Containers[0].Args = args

	pod, err = k8stest.CreatePod(pod, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())
	Expect(pod).ToNot(BeNil())

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	log.Log.Info("fio test pod is running.")

	msvc_err := k8stest.MsvConsistencyCheck(uid)
	Expect(msvc_err).ToNot(HaveOccurred(), "%v", msvc_err)

	log.Log.Info("Waiting for run to complete", "timeout", params.FioTimeout)
	tSecs := 0
	var phase coreV1.PodPhase
	for {
		if tSecs > params.FioTimeout {
			break
		}
		time.Sleep(1 * time.Second)
		tSecs += 1
		phase, err = k8stest.CheckPodCompleted(fioPodName, common.NSDefault)
		Expect(err).To(BeNil(), "CheckPodComplete got error %s", err)
		if phase != coreV1.PodRunning {
			break
		}
	}
	Expect(phase == coreV1.PodSucceeded).To(BeTrue(), "fio pod phase is %s", phase)
	log.Log.Info("fio completed", "duration", tSecs)

	// Delete the fio pod
	err = k8stest.DeletePod(fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	err = k8stest.RmPVC(volName, scName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", volName)

	err = k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

func TestCISmokeTest(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "CI_Smoke_Test", "CI_Smoke_Test")
}

var _ = Describe("CI_Smoke_Test", func() {

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

	It("should verify a volume can be created, used and deleted", func() {
		CISmokeTest()
	})
})

var _ = BeforeSuite(func(done Done) {
	err := k8stest.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnv %v", err)
	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	err := k8stest.TeardownTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)
})
