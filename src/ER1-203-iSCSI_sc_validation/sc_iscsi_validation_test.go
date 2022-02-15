package iscsi_sc_validation

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common/k8stest"
)

func TestIscsiScValidation(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "ER1-203", "ER1-203")
}

var _ = Describe("iSCSi protocol StorageClass validation Tests:", func() {

	BeforeEach(func() {
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Unsupported iscsi protocol validation test", func() {
		c := GenerateScIscsiValidationConfig("sc-iscsi-validation")
		c.ScIscsiValidation()

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
