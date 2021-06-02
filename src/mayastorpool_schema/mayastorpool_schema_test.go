package mayastorpool_schema

import (
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/k8stest"
)

func TestMayastorPoolSchema(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MayastorPool Schema Test", "mayastorpool_schema")
}

func mayastorPoolSchemaTest(schema string) {
	const timoSecs = 60
	const timoSleepSecs = 5
	pools, err := custom_resources.ListPools()
	Expect(err).ToNot(HaveOccurred())
	logf.Log.Info("Creating Mayastor Pool")
	for _, pool := range pools {
		err := custom_resources.DeletePool(pool.Name)
		Expect(err).ToNot(HaveOccurred())
	}
	for _, pool := range pools {
		if schema == "default" {
			k8stest.CreateConfiguredPools()
			break
		} else {
			diskPath := make([]string, 1)
			diskPath[0] = schema + "://" + pool.Spec.Disks[0]
			_, err = custom_resources.CreatePool(pool.Name, pool.Spec.Node, diskPath)
			Expect(err).ToNot(HaveOccurred())
		}
	}
	// Wait for pools to be online
	for ix := 0; ix < timoSecs/timoSleepSecs; ix++ {
		time.Sleep(timoSleepSecs * time.Second)
		err := k8stest.CheckAllPoolsAreOnline()
		if err == nil {
			break
		}
	}
	Expect(err).To(BeNil(), "One or more pools are offline")
	logf.Log.Info("Verifying Mayastor Pool device schema")
	pools, err = custom_resources.ListPools()
	Expect(err).ToNot(HaveOccurred())

	for _, pool := range pools {
		Expect(len(pool.Status.Disks) == 1).To(BeTrue(), "unexpected pool disks %v", pool.Status.Disks)
		Eventually(func() bool {
			if schema == "default" {
				return strings.Contains(pool.Status.Disks[0], "aio")
			} else {
				return strings.Contains(pool.Status.Disks[0], schema)
			}
		},
		).Should(Equal(true))
	}
	// RestoreConfiguredPools (re)create pools as defined by the configuration.
	// As part of the tests we may modify the pools, in such test cases
	// the test should delete all pools and recreate the configured set of pools.
	err = k8stest.RestoreConfiguredPools()
	Expect(err).To(BeNil(), "Not all pools are online after restoration")
}

var _ = Describe("Mayastor Pool schema test IO", func() {

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

	It("should verify MayastorPool schema with default schema", func() {
		mayastorPoolSchemaTest("default")
	})
	It("should verify MayastorPool schema with aio schema", func() {
		mayastorPoolSchemaTest("aio")
	})
	It("should verify MayastorPool schema with uring schema", func() {
		mayastorPoolSchemaTest("uring")
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
