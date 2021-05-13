package install

import (
	"fmt"
	"os/exec"
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

func generateYamlFiles(imageTag string, mayastorNodes []string, e2eCfg *e2e_config.E2EConfig) {
	coresDirective := ""
	if e2eCfg.Cores != 0 {
		coresDirective = fmt.Sprintf("%s -c %d", coresDirective, e2eCfg.Cores)
	}

	poolDirectives := ""
	if len(e2eCfg.PoolDevice) != 0 {
		poolDevice := e2eCfg.PoolDevice
		for _, mayastorNode := range mayastorNodes {
			poolDirectives += fmt.Sprintf(" -p '%s,%s'", mayastorNode, poolDevice)
		}
	}

	registryDirective := ""
	if len(e2eCfg.Registry) != 0 {
		registryDirective = fmt.Sprintf(" -r '%s'", e2eCfg.Registry)
	}

	bashCmd := fmt.Sprintf(
		"%s/generate-deploy-yamls.sh -o %s -t '%s' %s %s %s test",
		locations.GetMayastorScriptsDir(),
		locations.GetGeneratedYamlsDir(),
		imageTag, registryDirective, coresDirective, poolDirectives,
	)
	cmd := exec.Command("bash", "-c", bashCmd)
	out, err := cmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred(), "%s", out)
}

// Install mayastor on the cluster under test.
// We deliberately call out to kubectl, rather than constructing the client-go
// objects, so that we can verify the local deploy yaml files are correct.
func installMayastor() {
	e2eCfg := e2e_config.GetConfig()

	Expect(e2eCfg.ImageTag).ToNot(BeEmpty(),
		"mayastor image tag not defined")
	Expect(e2eCfg.PoolDevice != "").To(BeTrue(),
		"configuration error pools are not defined.")

	imageTag := e2eCfg.ImageTag
	registry := e2eCfg.Registry

	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())

	var mayastorNodes []string
	numMayastorInstances := 0

	for _, node := range nodes {
		if node.MayastorNode && !node.MasterNode {
			mayastorNodes = append(mayastorNodes, node.NodeName)
			numMayastorInstances += 1
		}
	}
	Expect(numMayastorInstances).ToNot(Equal(0))

	logf.Log.Info("Install", "tag", imageTag, "registry", registry, "# of mayastor instances", numMayastorInstances)

	generateYamlFiles(imageTag, mayastorNodes, &e2eCfg)
	deployDir := locations.GetMayastorDeployDir()
	yamlsDir := locations.GetGeneratedYamlsDir()

	k8stest.EnsureE2EAgent()

	err = k8stest.MkNamespace(common.NSMayastor())
	Expect(err).ToNot(HaveOccurred())
	k8stest.KubeCtlApplyYaml("moac-rbac.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("mayastorpoolcrd.yaml", deployDir)
	k8stest.KubeCtlApplyYaml("nats-deployment.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("csi-daemonset.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("moac-deployment.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("mayastor-daemonset.yaml", yamlsDir)

	ready, err := k8stest.MayastorReady(2, 540)
	Expect(err).ToNot(HaveOccurred())
	Expect(ready).To(BeTrue())

	// Now create configured pools on all nodes.
	k8stest.CreateConfiguredPools()

	// Wait for pools to be online
	const timoSecs = 120
	const timoSleepSecs = 10
	for ix := 0; ix < timoSecs/timoSleepSecs; ix++ {
		time.Sleep(timoSleepSecs * time.Second)
		err = k8stest.CheckAllPoolsAreOnline()
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
