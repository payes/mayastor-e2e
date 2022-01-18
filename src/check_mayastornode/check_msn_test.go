package check_mayastornode

import (
	"mayastor-e2e/common/custom_resources"
	"reflect"
	"sort"
	"testing"

	storageV1 "k8s.io/api/storage/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
)

func TestMayastorNode(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MayastorNode Test", "check_mayastornode")
}

var defTimeoutSecs = "60s"

func mayastorNodeTest(protocol common.ShareProto, volumeType common.VolumeType, fsType common.FileSystemType, mode storageV1.VolumeBindingMode) {

	// List nodes
	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())

	// CheckMayastorNodesAreOnine checks all the msns are
	// in online state or not if any of the msn is not in
	// online state then it returns error.
	err = custom_resources.CheckMsNodesAreOnline()
	Expect(err).ToNot(HaveOccurred())

	var workerNodes []string

	for _, node := range nodes {
		if node.MayastorNode && !node.MasterNode {
			workerNodes = append(workerNodes, node.NodeName)
		}
	}

	// List of MSN
	msnList, err := k8stest.GetMayastorNodeNames()
	Expect(err).ToNot(HaveOccurred())
	Expect(len(workerNodes) == len(msnList)).To(BeTrue(), "Invalid number of nodes in MSN list")

	// Check if mayastornode is present in all nodes
	Eventually(func() bool {
		sort.Strings(workerNodes)
		sort.Strings(msnList)
		return reflect.DeepEqual(workerNodes, msnList)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))

}

var _ = Describe("MayastorNode check test", func() {

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

	It("Should verify MSN on every node", func() {
		mayastorNodeTest(common.ShareProtoNvmf, common.VolFileSystem, common.XfsFsType, storageV1.VolumeBindingImmediate)
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
