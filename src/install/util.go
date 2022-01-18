package install

import (
	"fmt"
	"os/exec"

	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/locations"

	. "github.com/onsi/gomega"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func GenerateYamlFiles() {
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
			poolDirectives += fmt.Sprintf(" -p '%s,%s'", node.NodeName, poolDevice)
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
