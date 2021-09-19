package install

import (
	"fmt"
	"mayastor-e2e/common"
	"os/exec"
	"time"

	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/locations"

	. "github.com/onsi/gomega"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func GenerateMayastorYamlFiles() {
	e2eCfg := e2e_config.GetConfig()

	coresDirective := ""
	if e2eCfg.Cores != 0 {
		coresDirective = fmt.Sprintf("%s -c %d", coresDirective, e2eCfg.Cores)
	}

	nodeLocs, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "GetNodeLocs failed %v", err)
	poolDirectives := ""
	masterNode := ""
	if len(e2eCfg.PoolDevice) != 0 {
		poolDevice := e2eCfg.PoolDevice
		for _, node := range nodeLocs {
			if node.MasterNode {
				masterNode = node.NodeName
			}
			if !node.MayastorNode {
				continue
			}
			if !k8stest.IsControlPlaneMcp() {
				poolDirectives += fmt.Sprintf(" -p '%s,%s'", node.NodeName, poolDevice)
			}
		}
	}

	registryDirective := ""
	if len(e2eCfg.Registry) != 0 {
		registryDirective = fmt.Sprintf(" -r '%s'", e2eCfg.Registry)
	}

	imageTag := e2eCfg.ImageTag

	etcdOptions := "etcd.replicaCount=1,etcd.nodeSelector=kubernetes.io/hostname: " + masterNode + ",etcd.tolerations=- key: node-role.kubernetes.io/master"
	bashCmd := fmt.Sprintf(
		"%s/generate-deploy-yamls.sh -s '%s' -o %s -t '%s' %s %s %s test",
		locations.GetMayastorScriptsDir(),
		etcdOptions,
		locations.GetGeneratedYamlsDir(),
		imageTag, registryDirective, coresDirective, poolDirectives,
	)
	logf.Log.Info("About to execute", "command", bashCmd)
	cmd := exec.Command("bash", "-c", bashCmd)
	out, err := cmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred(), "%s", out)
}

func WaitForPoolCrd() bool {
	const timoSleepSecs = 5
	const timoSecs = 60
	for ix := 0; ix < timoSecs; ix += timoSleepSecs {
		_, err := custom_resources.ListMsPools()
		if err != nil {
			logf.Log.Info("WaitForPoolCrd", "error", err)
			if k8serrors.IsNotFound(err) {
				logf.Log.Info("WaitForPoolCrd, error := IsNotFound")
			} else {
				Expect(err).ToNot(HaveOccurred(), "%v", err)
			}
		} else {
			return true
		}
	}
	return false
}

func GenerateMCPYamlFiles() {
	e2eCfg := e2e_config.GetConfig()

	if k8stest.IsControlPlaneMcp() {
		registryDirective := ""
		if len(e2eCfg.Registry) != 0 {
			registryDirective = fmt.Sprintf(" -r '%s'", e2eCfg.Registry)
		}

		imageTag := e2eCfg.ImageTag

		bashCmd := fmt.Sprintf(
			"%s/generate-deploy-yamls.sh -o %s -t '%s' %s test",
			locations.GetMCPScriptsDir(),
			locations.GetGeneratedYamlsDir(),
			imageTag, registryDirective,
		)
		logf.Log.Info("About to execute", "command", bashCmd)
		cmd := exec.Command("bash", "-c", bashCmd)
		out, err := cmd.CombinedOutput()
		Expect(err).ToNot(HaveOccurred(), "%s", out)
	}
	// nothing to do for MOAC
}

// Install mayastor on the cluster under test.
// We deliberately call out to kubectl, rather than constructing the client-go
// objects, so that we can verify the local deploy yaml files are correct.
func InstallMayastor() {
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

	GenerateMCPYamlFiles()
	GenerateMayastorYamlFiles()
	yamlsDir := locations.GetGeneratedYamlsDir()
	logf.Log.Info("", "yamlsDir", yamlsDir)

	k8stest.EnsureE2EAgent()

	err = k8stest.MkNamespace(common.NSMayastor())
	Expect(err).ToNot(HaveOccurred())

	k8stest.KubeCtlApplyYaml("etcd", yamlsDir)
	k8stest.KubeCtlApplyYaml("nats-deployment.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("csi-daemonset.yaml", yamlsDir)
	k8stest.KubeCtlApplyYaml("mayastor-daemonset.yaml", yamlsDir)

	if k8stest.IsControlPlaneMcp() {
		k8stest.KubeCtlApplyYaml("operator-rbac.yaml", yamlsDir)
		k8stest.KubeCtlApplyYaml("core-agents-deployment.yaml", yamlsDir)
		k8stest.KubeCtlApplyYaml("rest-deployment.yaml", yamlsDir)
		k8stest.KubeCtlApplyYaml("rest-service.yaml", yamlsDir)
		k8stest.KubeCtlApplyYaml("msp-deployment.yaml", yamlsDir)
	} else {
		k8stest.KubeCtlApplyYaml("moac-rbac.yaml", yamlsDir)
		k8stest.KubeCtlApplyYaml("moac-deployment.yaml", yamlsDir)
	}

	ready, err := k8stest.MayastorReady(2, 540)
	Expect(err).ToNot(HaveOccurred())
	Expect(ready).To(BeTrue())

	ready = k8stest.ControlPlaneReady(10, 180)
	Expect(ready).To(BeTrue())

	crdReady := WaitForPoolCrd()
	Expect(crdReady).To(BeTrue())

	// Now create configured pools on all nodes.
	k8stest.CreateConfiguredPools()

	// Wait for pools to be online
	const timoSecs = 240
	const timoSleepSecs = 10
	for ix := 0; ix < timoSecs/timoSleepSecs; ix++ {
		time.Sleep(timoSleepSecs * time.Second)
		err = custom_resources.CheckAllMsPoolsAreOnline()
		if err == nil {
			break
		}
	}
	Expect(err).To(BeNil(), "One or more pools are offline")

	// hack core-agents typically has 4 restarts after deployment,
	// attempt to clear the number of restarts by re-deploying core-agents
	if k8stest.IsControlPlaneMcp() {
		k8stest.KubeCtlDeleteYaml("core-agents-deployment.yaml", yamlsDir)
		time.Sleep(10 * time.Second)
		k8stest.KubeCtlApplyYaml("core-agents-deployment.yaml", yamlsDir)
		ready = k8stest.ControlPlaneReady(10, 180)
		Expect(ready).To(BeTrue(), "re-deploy core-agents to zero restart count")
	}

	// Mayastor has been installed and is now ready for use.
}
