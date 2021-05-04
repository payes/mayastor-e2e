package mayastorpool_schema

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/common/crds"
	"mayastor-e2e/common/k8stest"
)

var defTimeoutSecs = "120s"

func TestMayastorPoolSchema(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MayastorPool Schema Test, NVMe-oF TCP and iSCSI", "mayastorpool_schema")
}

func mayastorPoolSchemaTest() {
	pools, _ := crds.ListPools()
	logf.Log.Info("Creating Mayastor Pool without specifying a device schema")
	for _, pool := range pools {
		err := crds.DeletePool(pool.Name)
		Expect(err).ToNot(HaveOccurred())
		err = crds.CreatePool(pool.Name, pool.Spec.Node, pool.Spec.Disks)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() bool {
			return strings.Contains(pool.Status.Disks[0], "uring")
		},
		).Should(Equal(true))
		logf.Log.Info("Verified Mayastor Pool device schema")
		logf.Log.Info("Creating Mayastor Pool with device schema aio")
		diskPath := make([]string, 1)
		err = crds.DeletePool(pool.Name)
		Expect(err).ToNot(HaveOccurred())
		diskPath[0] = "aio://" + pool.Spec.Disks[0]
		err = crds.CreatePool(pool.Name, pool.Spec.Node, diskPath)
		Expect(err).ToNot(HaveOccurred())
	}
	pools, _ = crds.ListPools()
	for _, pool := range pools {
		Eventually(func() bool {
			return strings.Contains(pool.Status.Disks[0], "aio")
		},
			defTimeoutSecs,
			"1s",
		).Should(Equal(true))
	}
	logf.Log.Info("Verified Mayastor Pool with device schema aio")
}

var _ = Describe("Mayastor Pool schema tesr IO", func() {

	// BeforeEach(func() {
	// 	// Check ready to run
	// 	err := k8stest.BeforeEachCheck()
	// 	Expect(err).ToNot(HaveOccurred())
	// })

	// AfterEach(func() {
	// 	// Check resource leakage.
	// 	err := k8stest.AfterEachCheck()
	// 	Expect(err).ToNot(HaveOccurred())
	// })

	It("should verify MayastorPool schema", func() {
		mayastorPoolSchemaTest()
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
