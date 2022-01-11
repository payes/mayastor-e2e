package k8stest

import (
	"mayastor-e2e/common/mayastorclient"

	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func WaitForMCPPath(timeout string) {
	Eventually(func() error {
		// If this call goes through implies
		// REST, Core Agent and etcd pods are up and running
		_, err := ListMsvs()
		if err != nil {
			logf.Log.Info("Failed t list msvs", "error", err)
		}
		return err
	},
		timeout,
		"5s",
	).Should(BeNil())
}

func WaitForMayastorSockets(addrs []string, timeout string) {
	Eventually(func() error {
		// If this call goes through without an error imples
		// the listeners at the pod have started
		_, err := mayastorclient.ListReplicas(addrs)
		if err != nil {
			logf.Log.Info("Failed to list replicas", "error", err)
		}
		return err
	},
		timeout,
		"5s",
	).Should(BeNil())
}
