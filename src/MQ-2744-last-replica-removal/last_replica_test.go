// JIRA: CAS-1284
package last_replica

import (
	"fmt"
	"mayastor-e2e/common/k8stest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"

	//	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})

func TestLastReplica(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-2744", "MQ-2744")
}

var _ = Describe("Mayastor last replica test", func() {

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

	It("should verify that the last replica of volume is not removed, when faulted using gRPC", func() {
		ctx := makeTestContext("last-replica-grpc-faulting")
		setupTestVolAndPod(ctx)

		assertCheckChildren(ctx)
		faultReplica(ctx)
		assertCheckChildren(ctx)

		exitCode, err := waitPodComplete(ctx)
		logf.Log.Info("", "exitCode", exitCode, "error", err)

		assertCheckChildren(ctx)

		cleanup(ctx)
		Expect(exitCode).To(BeZero(), "fio pod had error")
	})

	It("should verify that the last replica of volume is not removed, when pools devices are briefly taken offline", func() {
		for ix := 0; ix < 100; ix++ {
			ctx := makeTestContext(fmt.Sprintf("last-replica-pool-blip-%d", ix))

			setupTestVolAndPod(ctx)
			time.Sleep(10 * time.Second)

			assertCheckChildren(ctx)
			blipPoolDevices(ctx)
			assertCheckChildren(ctx)

			exitCode, err := waitPodComplete(ctx)
			logf.Log.Info("", "exitCode", exitCode, "error", err)
			// ignore exitCode it is irrelevant whether the fio pod ran successfully
			assertCheckChildren(ctx)

			//		_ = k8stest.DumpPodInfo(ctx.podName, common.NSDefault)
			cleanup(ctx)
		}
	})

	It("should verify that the last replica of volume is not removed, when pools devices are taken offline", func() {
		ctx := makeTestContext("last-replica-pool-offline")
		setupTestVolAndPod(ctx)
		offlinePoolDevices(ctx)

		phase, err := waitPodComplete(ctx)
		logf.Log.Info("", "pod phase", phase, "error", err)

		nc, err := checkChildren(ctx)

		//Restore the pools
		onlinePoolDevices(ctx)
		time.Sleep(30 * time.Second)

		Expect(err).ToNot(HaveOccurred(), "Failure when checking nexus children")
		Expect(nc > 0).To(BeTrue())

		cleanup(ctx)
	})

})
