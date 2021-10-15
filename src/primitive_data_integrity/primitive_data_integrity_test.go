package primitivedataintegrity

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/e2e_config"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	defTimeoutSecs = "180s"
)

type IntegrityEnv struct {
	volUuid        string
	pvc            string
	storageClass   string
	fioPodName     string
	nexusIP        string
	replicaIPs     []string
	fioTimeoutSecs int
	nexusUuid      string
}

var env IntegrityEnv

func (env *IntegrityEnv) setupReplicas() {
	nodeList, err := k8stest.GetNodeLocs() // all 3 nodes as IP address + name
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	nexus, replicaNodes := k8stest.GetMsvNodes(env.volUuid) // names of nodes in volume
	Expect(nexus).NotTo(Equal(""), "Nexus not found")

	// identify the nexus IP address
	nexusIP := ""
	for _, node := range nodeList {
		if node.NodeName == nexus {
			nexusIP = node.IPAddress
			break
		}
	}
	Expect(nexusIP).NotTo(Equal(""), "Nexus IP address not found")
	env.nexusIP = nexusIP
	msv, err := k8stest.GetMSV(env.volUuid)
	Expect(err).ToNot(HaveOccurred(), "failed to retrieve MSV for volume %s", env.volUuid)
	env.nexusUuid = msv.Status.Nexus.Uuid
	// if necessary, reconfigure the volume not to include the nexus node as a replica
	changed, err := k8stest.ExcludeNexusReplica(nexusIP, env.nexusUuid, env.volUuid)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	if changed {
		nexus, replicaNodes = k8stest.GetMsvNodes(env.volUuid) // names of nodes in volume
	}
	var replicaIPs []string

	for _, node := range nodeList {
		if node.NodeName != nexus {
			for _, replica := range replicaNodes {
				if replica == node.NodeName {
					replicaIPs = append(replicaIPs, node.IPAddress)
					break
				}
			}
		}
	}

	Expect(len(replicaIPs)).To(Equal(2), "Expected to find 2 non-nexus replicas")
	env.replicaIPs = replicaIPs
	logf.Log.Info("identified", "nexus", env.nexusIP, "replica1", env.replicaIPs[0], "replica2", env.replicaIPs[1])
}

func setup(pvcName string, storageClassName string, fioPodName string) IntegrityEnv {
	e2eCfg := e2e_config.GetConfig()
	volMb := e2eCfg.PrimitiveDataIntegrity.VolMb
	env := IntegrityEnv{}

	env.fioTimeoutSecs = volMb / 2

	env.pvc = pvcName
	env.storageClass = storageClassName
	env.volUuid = k8stest.MkPVC(volMb, pvcName, storageClassName, common.VolRawBlock, common.NSDefault)

	podObj := k8stest.CreateFioPodDef(fioPodName, pvcName, common.VolRawBlock, common.NSDefault)
	_, err := k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	env.fioPodName = fioPodName
	logf.Log.Info("waiting for pod", "name", env.fioPodName)
	Eventually(func() bool {
		return k8stest.IsPodRunning(env.fioPodName, common.NSDefault)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(true))

	env.setupReplicas()
	return env
}

// Common steps required when tearing down the test
func (env *IntegrityEnv) teardown() {
	var err error

	if env.fioPodName != "" {
		err := k8stest.DeletePod(env.fioPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env.fioPodName = ""
	}
	if env.pvc != "" {
		k8stest.RmPVC(env.pvc, env.storageClass, common.NSDefault)
		env.pvc = ""
	}
	if env.storageClass != "" {
		err = k8stest.RmStorageClass(env.storageClass)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env.storageClass = ""
	}
}

// Run fio against the device, finish when all blocks are accessed
func runFio(podName string, filename string, args ...string) ([]byte, error) {
	argFilename := fmt.Sprintf("--filename=%s", filename)

	logf.Log.Info("RunFio",
		"podName", podName,
		"filename", filename,
		"args", args)

	cmdArgs := []string{
		"exec",
		"-it",
		podName,
		"--",
		"fio",
		"--name=benchtest",
		"--verify_fatal=1",
		"--verify_async=2",
		argFilename,
		"--direct=1",
		"--ioengine=libaio",
		"--bs=4k",
		"--iodepth=16",
		"--numjobs=1",
	}

	if args != nil {
		cmdArgs = append(cmdArgs, args...)
	}

	cmd := exec.Command(
		"kubectl",
		cmdArgs...,
	)
	cmd.Dir = ""
	output, err := cmd.CombinedOutput()
	if err != nil {
		logf.Log.Info("Running fio failed", "error", err, "output", string(output))
	}
	return output, err
}

// write to all blocks with a block-specific pattern and its checksum
func (env *IntegrityEnv) fioWriteAndVerify(fioPodName string) error {
	var err error
	ch := make(chan bool, 1)

	go func() {
		_, err = runFio(
			fioPodName,
			common.FioBlockFilename,
			"--rw=randwrite",
			"--verify=crc32",
			"--verify_pattern=%o")
		ch <- true
	}()
	select {
	case <-ch:
		return err
	case <-time.After(time.Duration(env.fioTimeoutSecs) * time.Second):
		return fmt.Errorf("Fio timed out")
	}
}

// 1) create a 2-replica volume
// 2) ensure the nexus is not one of the replicas
// 3) get fio to write to the entire volume
// 4) use the e2e-agent running on each non-nexus node:
//    for each non-nexus replica node
//        nvme connect to its own target
//        cksum /dev/nvme0n1p2
//        nvme disconnect
//    compare the checksum results, they should match
func (env *IntegrityEnv) PrimitiveDataIntegrity() {
	logf.Log.Info("writing to the volume")
	err := env.fioWriteAndVerify(env.fioPodName)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// the first replica
	replicas, err := mayastorclient.ListReplicas([]string{env.replicaIPs[0]})
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(len(replicas)).To(Equal(1), "Expected to find 1 replica")
	uri := replicas[0].Uri
	logf.Log.Info("uri", "uri", uri)
	firstchecksum, err := k8stest.ChecksumReplica(env.replicaIPs[0], env.replicaIPs[0], uri)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// the second replica
	replicas, err = mayastorclient.ListReplicas([]string{env.replicaIPs[1]})
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(len(replicas)).To(Equal(1), "Expected to find 1 replica")
	uri = replicas[0].Uri
	logf.Log.Info("uri", "uri", uri)
	secondchecksum, err := k8stest.ChecksumReplica(env.replicaIPs[1], env.replicaIPs[1], uri)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// verify that they match
	logf.Log.Info("match", "first", firstchecksum, "this", secondchecksum)
	Expect(secondchecksum).To(Equal(firstchecksum), "checksums differ")
}

func TestPrimitiveDataIntegrity(t *testing.T) {
	// Initialise test and set class and file names for reports
	k8stest.InitTesting(t, "MQ-1510", "MQ-1510")
}

var _ = Describe("Primitive data integrity:", func() {

	BeforeEach(func() {
		// Check ready to run
		err := k8stest.BeforeEachCheck()
		Expect(err).ToNot(HaveOccurred(), "%v", err)
	})

	AfterEach(func() {
		env.teardown()
		// Check resource leakage.
		err := k8stest.AfterEachCheck()
		Expect(err).ToNot(HaveOccurred(), "%v", err)
	})

	It("should verify data is duplicated to replicas", func() {
		sc := "sc-primitive-data-integrity"
		err := k8stest.MkStorageClass(sc, 2, common.ShareProtoNvmf, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		env = setup("pvc-primitive-data-integrity", sc, "fio-primiive-data-integrity")
		env.PrimitiveDataIntegrity()
	})
})

var _ = BeforeSuite(func(done Done) {
	k8stest.SetupTestEnv()
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	k8stest.TeardownTestEnv()
})
