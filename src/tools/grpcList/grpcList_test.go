package cleanup

import (
	"mayastor-e2e/common/mayastorclient"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"mayastor-e2e/common/k8stest"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// This is run as a test but is really a utility to restore
// the cluster to usable state and restart mayastor.
func TestRestartMayastor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GRPC lists")
}

var _ = Describe("Mayastor utility: gRPC lists", func() {
	It("should use gRPC to list mayastor nexuses, replicas and pools", func() {
		nodes, err := k8stest.GetNodeLocs()
		if err != nil {
			logf.Log.Info("list nodes failed", "error", err)
			return
		}
		for _, node := range nodes {
			if !node.MayastorNode {
				continue
			}
			addrs := []string{node.IPAddress}
			nexuses, err := mayastorclient.ListNexuses(addrs)
			logf.Log.Info(node.NodeName, "IP address", node.IPAddress)
			if err == nil {
				for _, nexus := range nexuses {
					logf.Log.Info("", "nexus", nexus)
				}
			} else {
				logf.Log.Info("nexuses", "error", err)
			}
			replicas, err := mayastorclient.ListReplicas(addrs)
			if err == nil {
				for _, replica := range replicas {
					logf.Log.Info("", "replica", replica)
				}
			} else {
				logf.Log.Info("replicas", "error", err)
			}
			pools, err := mayastorclient.ListPools(addrs)
			if err == nil {
				for _, pool := range pools {
					logf.Log.Info("", "pool", pool)
				}
			} else {
				logf.Log.Info("pools", "error", err)
			}
		}
	})
})

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
