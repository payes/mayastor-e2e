package install

import (
	"mayastor-e2e/common/custom_resources"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/locations"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Install mayastor on the cluster under test.
// We deliberately call out to kubectl, rather than constructing the client-go
// objects, so that we can verify the local deploy yaml files are correct.
func installMayastor() {
	e2eCfg := e2e_config.GetConfig()

	Expect(e2eCfg.ImageTag).ToNot(BeEmpty(),
		"mayastor image tag not defined")
	Expect(e2eCfg.PoolDevice != "").To(BeTrue(),
		"configuration error pools are not defined.")

	mayastorNodes, err := k8stest.GetMayastorNodeNames()
	Expect(err).ToNot(HaveOccurred())

	numMayastorInstances := len(mayastorNodes)
	Expect(numMayastorInstances).ToNot(Equal(0))

	logf.Log.Info("Install", "tag", e2eCfg.ImageTag, "registry", e2eCfg.Registry, "# of mayastor instances", numMayastorInstances)

	GenerateYamlFiles()
	yamlsDir := locations.GetGeneratedYamlsDir()

	k8stest.EnsureE2EAgent()

	err = k8stest.MkNamespace(common.NSMayastor())
	Expect(err).ToNot(HaveOccurred())
	k8stest.KubeCtlApplyYaml("moac-rbac.yaml", yamlsDir)

	k8stest.KubeCtlApplyYaml("etcd", yamlsDir)
	k8stest.KubeCtlApplyYaml("nats-deployment.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("csi-daemonset.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("moac-deployment.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("mayastor-daemonset.yaml", yamlsDir)

	ready, err := k8stest.MayastorReady(2, 540)
	Expect(err).ToNot(HaveOccurred())
	Expect(ready).To(BeTrue())

	crdReady := WaitForPoolCrd()
	Expect(crdReady).To(BeTrue())

	// Now create configured pools on all nodes.
	k8stest.CreateConfiguredPools()

	// Wait for pools to be online
	const timoSecs = 120
	const timoSleepSecs = 10
	for ix := 0; ix < timoSecs/timoSleepSecs; ix++ {
		time.Sleep(timoSleepSecs * time.Second)
		err = custom_resources.CheckAllMsPoolsAreOnline()
		if err == nil {
			break
		}
	}
	Expect(err).To(BeNil(), "One or more pools are offline")
	// Mayastor has been installed and is now ready for use.
}

func TestInstallSuite(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Basic Install Suite", "install")
}

var _ = Describe("Mayastor setup", func() {
	It("should install using yaml files", func() {
		installMayastor()
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	k8stest.TeardownTestEnvNoCleanup()
})
