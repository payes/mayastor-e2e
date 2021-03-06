package pool_modify

import (
	"mayastor-e2e/common/custom_resources"
	"reflect"
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestPoolModify(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Pool modification tests", "pool_modify")
}

var _ = Describe("Mayastor Pool Modification test", func() {
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

	It("should fail to expand existing lvol store", func() {
		pools, err := custom_resources.ListMsPools()
		Expect(err).To(BeNil(), "failed to list pools using custom resources API")
		for _, pool := range pools {
			updatedPool := pool
			updatedPool.Spec.Disks = append(pool.Spec.Disks, "/dev/sdx")

			logf.Log.Info("Updating pool", "from", pool.Spec, "to", updatedPool.Spec)
			poolRet, err := custom_resources.UpdateMsPool(updatedPool)
			logf.Log.Info("After UpdateMsPool", "err", err)
			Expect(reflect.DeepEqual(pool.Status, poolRet.Status)).To(BeTrue(), "updated pool status was modified on pool update %v", poolRet.Status)

			poolAgain, err := custom_resources.GetMsPool(pool.Name)
			Expect(err).To(BeNil(), "GetMsPool failed for %s", pool.Name)
			Expect(reflect.DeepEqual(pool.Status, poolAgain.Status)).To(BeTrue(), "pool status was modified on failed pool update %v", poolAgain.Status)
		}
		err = k8stest.RestoreConfiguredPools()
		Expect(err).To(BeNil(), "Not all pools are online after restoration")
	})

	// Suppresing this test case for HC1 release (v1.0).  At this time, the behaviour of the Mayastor Pool Operator in regards to
	// a change in spec.Node is undefined behaviour
	/*
		It("should fail to change status when pool spec node is modified", func() {
			pools, err := custom_resources.ListMsPools()
			Expect(err).To(BeNil(), "failed to list pools using custom resources API")
			var nodes []string
			for _, pool := range pools {
				nodes = append(nodes, pool.Spec.Node)
			}
			for ix, pool := range pools {
				updatedPool := pool
				updatedPool.Spec.Node = nodes[(ix+1)%len(nodes)]

				logf.Log.Info("Updating pool", "was", pool.Spec, "to", updatedPool.Spec)
				poolRet, err := custom_resources.UpdateMsPool(updatedPool)
				logf.Log.Info("After UpdateMsPool", "err", err)
				Expect(reflect.DeepEqual(pool.Status, poolRet.Status)).To(BeTrue(), "updated pool status was modified on pool update %v", poolRet.Status)

				poolAgain, err := custom_resources.GetMsPool(pool.Name)
				Expect(err).To(BeNil(), "GetMsPool failed for %s", pool.Name)
				Expect(reflect.DeepEqual(pool.Status, poolAgain.Status)).To(BeTrue(), "pool status was modified on failed pool update %v", poolAgain.Status)
			}
			err = k8stest.RestoreConfiguredPools()
			Expect(err).To(BeNil(), "Not all pools are online after restoration")
		})*/
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
