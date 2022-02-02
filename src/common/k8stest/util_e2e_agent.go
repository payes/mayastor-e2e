package k8stest

import (
	"context"
	"mayastor-e2e/common"
	"mayastor-e2e/common/locations"
	"time"

	"github.com/onsi/gomega"
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

func e2eReadyPodCount() int {
	var daemonSet appsV1.DaemonSet
	if gTestEnv.K8sClient.Get(context.TODO(), types.NamespacedName{Name: "e2e-rest-agent", Namespace: common.NSE2EAgent}, &daemonSet) != nil {
		return -1
	}
	return int(daemonSet.Status.NumberAvailable)
}

// EnsureE2EAgent ensure that e2e agent daemonSet is running, if already deployed
// does nothing, otherwise creates the e2e agent namespace and deploys the daemonSet.
// asserts if creating the namespace fails. This function can be called repeatedly.
func EnsureE2EAgent() bool {
	const sleepTime = 5
	const duration = 60
	count := (duration + sleepTime - 1) / sleepTime
	ready := false
	err := EnsureNamespace(common.NSE2EAgent)
	gomega.Expect(err).To(gomega.BeNil())

	nodes, _ := GetNodeLocs()
	instances := 0
	for _, node := range nodes {
		if node.MayastorNode && !node.MasterNode {
			instances += 1
		}
	}

	if e2eReadyPodCount() == instances {
		return true
	}

	err = KubeCtlApplyYaml("e2e-agent.yaml", locations.GetE2EAgentPath())

	if err != nil {
		return false
	}
	for ix := 0; ix < count && !ready; ix++ {
		time.Sleep(time.Duration(sleepTime) * time.Second)
		ready = e2eReadyPodCount() == instances
	}
	return ready
}
