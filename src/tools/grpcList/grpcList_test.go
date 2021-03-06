package cleanup

import (
	"mayastor-e2e/common/mayastorclient"
	"testing"

	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// This is run as a test but is really a utility to list
// resources using gRPC calls to mayastor instances
func TestMayastorClientList(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GRPC lists")
}

var _ = Describe("Mayastor utility: gRPC lists", func() {
	It("should use gRPC to list mayastor nexuses, replicas and pools", func() {
		Expect(mayastorclient.CanConnect()).To(BeTrue(), "unable to connect to all mayastor instances")
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
				logf.Log.Info("nexuses", "count", len(nexuses))
				for _, nexus := range nexuses {
					logf.Log.Info("", "nexus", nexus)
				}
			} else {
				logf.Log.Info("nexuses", "error", err)
			}
			replicas, err := mayastorclient.ListReplicas(addrs)
			if err == nil {
				logf.Log.Info("replicas", "count", len(replicas))
				for _, replica := range replicas {
					logf.Log.Info("", "replica", replica)
				}
			} else {
				logf.Log.Info("replicas", "error", err)
			}
			pools, err := mayastorclient.ListPools(addrs)
			logf.Log.Info("pools", "count", len(pools))
			if err == nil {
				for _, pool := range pools {
					logf.Log.Info("", "pool", pool)
				}
			} else {
				logf.Log.Info("pools", "error", err)
			}
			nvmeControllers, err := mayastorclient.ListNvmeControllers(addrs)
			logf.Log.Info("nvmeControllers", "count", len(nvmeControllers))
			if err == nil {
				for _, controller := range nvmeControllers {
					logf.Log.Info("", "nvmeController", controller)
				}
			} else {
				logf.Log.Info("nvmeControllers", "error", err)
			}
		}
	})
})

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(GinkgoWriter)))
	err := k8stest.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to setup test environment in BeforeSuite : SetupTestEnv %v", err)
	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	err := k8stest.TeardownTestEnv()
	Expect(err).ToNot(HaveOccurred(), "failed to tear down test environment in AfterSuite : TeardownTestEnv %v", err)
})
