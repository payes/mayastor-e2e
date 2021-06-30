package exp

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	mayastorgrpc "mayastor-e2e/common/mayastorclient/grpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type replicaDetails struct {
	node        string
	replicaUuid types.UID
}

func (rd replicaDetails) String() string {
	return fmt.Sprintf("node:%v, replicaUuid:%v", rd.node, rd.replicaUuid)
}

type poolDetails struct {
	pool mayastorclient.MayastorPool
	node string
}

var sizeTable []uint64
var mayastorNodePools []poolDetails

func TestPrimitiveReplicas(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "Primitive Replicas Tests", "primitive_replicas")
}

func makeInvalidReplica(pd poolDetails, replicaSize uint64, fuzzVal uint) error {
	replicaUuid := string(uuid.NewUUID())
	thin := false
	node := pd.node
	poolName := pd.pool.Name
	share := mayastorgrpc.ShareProtocolReplica_REPLICA_NVMF

	fv := fuzzVal % 4

	switch {
	case fv <= 1:
		replicaUuid += replicaUuid
		logf.Log.Info("makeInvalidReplica CASE 1:", "invalid UUID", replicaUuid)
	case fv == 2:
		poolName = "NotExist"
		logf.Log.Info("makeInvalidReplica CASE 2:", "invalid pool name", poolName)
	case fv == 3:
		replicaSize = 1024 * 1024 * 1024 * 1024 * 1024 * 1024
		logf.Log.Info("makeInvalidReplica CASE 3:", "invalid replica size (too large)", replicaSize)
		//	case 4:
		//		replicaSize = 0
		//		logf.Log.Info("makeInvalidReplica CASE 4:", "invalid replica size", replicaSize)
		//	case 5:
		//		thin = true
		//		logf.Log.Info("makeInvalidReplica CASE 5:", "thin", thin)
	}
	var err error
	logf.Log.Info("makeInvalidReplica", "replicaSize", replicaSize, "replicaSizeMiB",
		replicaSize/(1024*1024),
		"replicaUuid", replicaUuid,
		"node", node,
		"pool", poolName,
		"share", share,
	)

	err = mayastorclient.CreateReplicaExt(node, replicaUuid, replicaSize, poolName, thin, share)
	if err == nil {
		return fmt.Errorf("creating replica with UUID=%v succeeded", replicaUuid)
	}
	return nil
}

func makeReplica(pd poolDetails, replicaSize uint64) (*replicaDetails, error) {
	replicaUuid := uuid.NewUUID()
	var rd *replicaDetails
	var err error
	logf.Log.Info("makeReplica",
		"replicaSize", replicaSize,
		"replicaSizeMiB", replicaSize/(1024*1024),
		"replicaUuid", replicaUuid,
		"node", pd.node,
		"pool", pd.pool,
	)

	err = mayastorclient.CreateReplica(pd.node, string(replicaUuid), replicaSize, pd.pool.Name)
	if err == nil {
		rd = &replicaDetails{node: pd.node, replicaUuid: replicaUuid}
		logf.Log.Info("makeReplica", "replica details", rd)
	} else {
		logf.Log.Info("makeReplica",
			"replicaUuid", replicaUuid,
			"error", err)
	}
	return rd, err
}

func checkPoolsOnline(nodes []string) {
	for _, node := range nodes {
		pools, err := mayastorclient.ListPools([]string{node})
		Expect(err).ToNot(HaveOccurred(), "failed to List pools on %s", node)
		for _, pool := range pools {
			Expect(pool.State).To(Equal(mayastorgrpc.PoolState_POOL_ONLINE))
		}
	}
}

func initPoolsList() {
	nodes, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred())

	mayastorNodePools = []poolDetails{}
	for _, node := range nodes {
		if !node.MayastorNode {
			continue
		}
		pools, err := mayastorclient.ListPools([]string{node.IPAddress})
		Expect(err).ToNot(HaveOccurred(), "failed to retrieve pools for %s: %v", node.IPAddress, err)
		for _, pool := range pools {
			mayastorNodePools = append(mayastorNodePools, poolDetails{node: node.IPAddress, pool: pool})
		}
	}
}

func creatDelTest() {
	initPoolsList()

	for ix := 0; ix < e2e_config.GetConfig().PrimitiveReplicas.Iterations; ix += 1 {
		replicaSize := sizeTable[ix%len(sizeTable)]
		pd := mayastorNodePools[ix%len(mayastorNodePools)]

		rd, err := makeReplica(pd, replicaSize)
		Expect(err).ToNot(HaveOccurred(), "failed to create replica of size %d on %v", replicaSize, pd)
		Expect(rd).ToNot(BeNil(), "got nil pointer to replica details")

		checkPoolsOnline([]string{pd.node})

		replicas, err := k8stest.ListReplicasInCluster()
		Expect(err).ToNot(HaveOccurred(), "failed to list replicas")
		Expect(len(replicas)).To(Equal(1))

		err = mayastorclient.RmReplica(rd.node, string(rd.replicaUuid))
		Expect(err).ToNot(HaveOccurred())

		checkPoolsOnline([]string{pd.node})

		err = k8stest.CheckTestPodsHealth(common.NSMayastor())
		Expect(err).ToNot(HaveOccurred(), "mayastor pods not healthy")

		replicas, err = k8stest.ListReplicasInCluster()
		Expect(err).ToNot(HaveOccurred(), "failed to list replicas")
		Expect(len(replicas)).To(BeZero())
	}
}

func fuzzTest() {
	initPoolsList()

	var err error
	for ix := 0; ix < e2e_config.GetConfig().PrimitiveReplicas.Iterations; ix += 1 {
		pd := mayastorNodePools[ix%len(mayastorNodePools)]
		replicaSize := sizeTable[ix%len(sizeTable)]
		err = makeInvalidReplica(pd, replicaSize, uint(ix))
		Expect(err).ToNot(HaveOccurred(), "%v", err)
	}
}

func createDeleteReplica(pd poolDetails, doneC chan string, errC chan<- error) {
	for ix := 0; ix < e2e_config.GetConfig().PrimitiveReplicas.Iterations; ix += 1 {
		replicaSize := sizeTable[ix%len(sizeTable)]

		rd, err := makeReplica(pd, replicaSize)
		if err != nil {
			errC <- err
			return
		}

		if rd == nil {
			errC <- fmt.Errorf("makeReplica returned nil pointer")
			return
		}

		err = mayastorclient.RmReplica(rd.node, string(rd.replicaUuid))
		if err != nil {
			errC <- err
			return
		}
	}
	doneC <- pd.pool.Name
}

func concurrentFuzz(errC chan<- error, quitC chan bool) {
	var err error
	for ix := 1; ; ix += 1 {
		select {
		case _, ok := <-quitC:
			if ok == false {
				return
			}
		default:
			pd := mayastorNodePools[ix%len(mayastorNodePools)]
			replicaSize := sizeTable[ix%len(sizeTable)]
			err = makeInvalidReplica(pd, replicaSize, uint(ix))
			if err != nil {
				errC <- err
				return
			}
		}
	}
}

func concurrentTest() {
	initPoolsList()
	doneC, errC, quitC := make(chan string), make(chan error), make(chan bool)
	go concurrentFuzz(errC, quitC)
	for _, pd := range mayastorNodePools {
		go createDeleteReplica(pd, doneC, errC)
	}

	var err error

	for ix := 1; ix <= len(mayastorNodePools); ix += 1 {
		select {
		case poolName := <-doneC:
			logf.Log.Info("Completed", "pool name", poolName)
		case err = <-errC:
			logf.Log.Info("", "error", err)
		}
	}
	close(quitC)
	Expect(err).ToNot(HaveOccurred())

	var nodes []string
	for _, n := range mayastorNodePools {
		nodes = append(nodes, n.node)
	}
	checkPoolsOnline(nodes)

	err = k8stest.CheckTestPodsHealth(common.NSMayastor())
	Expect(err).ToNot(HaveOccurred(), "mayastor pods not healthy")

	replicas, err := k8stest.ListReplicasInCluster()
	Expect(err).ToNot(HaveOccurred(), "failed to list replicas")
	Expect(len(replicas)).To(BeZero())
}

var _ = Describe("Mayastor Volume IO test", func() {

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

	cfg := e2e_config.GetConfig().PrimitiveReplicas
	logf.Log.Info("Using", "configuration", cfg)
	const mb = 1024 * 1024
	for rSize := uint64(cfg.StartSizeMb) * mb; rSize <= uint64(cfg.EndSizeMb)*mb; rSize += uint64(cfg.SizeStepMb) * mb {
		sizeTable = append(sizeTable, rSize)
	}

	It("should repeatedly create and delete replicas with health checks ", func() {
		creatDelTest()
	})

	It("should not create replicas with invalid parameters ", func() {
		fuzzTest()
	})

	It("should concurrently, repeatedly create and delete replicas with health checks and not create replicas with invalid parameters", func() {
		concurrentTest()
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	// NB This only tears down the local structures for talking to the cluster,
	// not the kubernetes cluster itself.
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
