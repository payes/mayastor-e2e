package primitive_device_retirement

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"

	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "120s"

var fioWriteParams = []string{
	"--name=benchtest",
	"--numjobs=1",
	"--direct=1",
	"--rw=randwrite",
	"--do_verify=0",
	"--ioengine=libaio",
	"--bs=4k",
	"--iodepth=16",
	"--verify=crc32",
}

var fioVerifyParams = []string{
	"--name=benchtest",
	"--numjobs=1",
	"--direct=1",
	"--rw=randread",
	"--ioengine=libaio",
	"--bs=4k",
	"--iodepth=16",
	"--verify=crc32",
	"--verify_fatal=1",
	"--verify_async=2",
}

func (c *primitiveDeviceRetirementConfig) createSC() {

	err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithLocal(true).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.scName)
}

func (c *primitiveDeviceRetirementConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

func (c *primitiveDeviceRetirementConfig) createPVC() string {
	// Create the volume with 1 replica
	return k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, c.volType, common.NSDefault)
}

func (c *primitiveDeviceRetirementConfig) deletePVC() {
	k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
}

func (c *primitiveDeviceRetirementConfig) createFioPod(nodeName string, verify bool) {

	var args = []string{
		"--",
	}
	args = append(args, fmt.Sprintf("--filename=%s", common.FioBlockFilename))

	if verify {
		args = append(args, fioVerifyParams...)
	} else {
		args = append(args, fioWriteParams...)
	}
	logf.Log.Info("fio", "arguments", args)

	// fio pod container
	podContainer := coreV1.Container{
		Name:            c.podName,
		Image:           common.GetFioImage(),
		ImagePullPolicy: coreV1.PullAlways,
		Args:            args,
	}
	// volume claim details
	volume := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: c.pvcName,
			},
		},
	}

	podObj, err := k8stest.NewPodBuilder().
		WithName(c.podName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithNodeSelectorHostnameNew(nodeName).
		WithVolumeDeviceOrMount(c.volType).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", c.podName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", c.podName)

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(c.podName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	logf.Log.Info("fio test pod is running.")
}

// use the e2e-agent run on each non-nexus node:
//    for each non-nexus replica node
//        nvme connect to a replica
//        checksum /dev/nvme0n1p2
//        disconnect
//    compare the checksum results, they should match
func (c *primitiveDeviceRetirementConfig) PrimitiveDataIntegrity() {

	// checksum first replica
	replicas, err := mayastorclient.ListReplicas([]string{c.replicaIPs[0]})
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(len(replicas)).To(Equal(1), "Expected to find 1 replica")
	uri := replicas[0].Uri
	logf.Log.Info("uri", "uri", uri)
	firstchecksum, err := k8stest.ChecksumReplica(c.replicaIPs[0], c.replicaIPs[0], uri)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// checksum second replica
	replicas, err = mayastorclient.ListReplicas([]string{c.replicaIPs[1]})
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(len(replicas)).To(Equal(1), "Expected to find 1 replica")
	uri = replicas[0].Uri
	logf.Log.Info("uri", "uri", uri)
	secondchecksum, err := k8stest.ChecksumReplica(c.replicaIPs[1], c.replicaIPs[1], uri)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// verify that they match
	logf.Log.Info("match", "first", firstchecksum, "this", secondchecksum)
	Expect(secondchecksum).To(Equal(firstchecksum), "checksums differ")
}

func (c *primitiveDeviceRetirementConfig) GetReplicaAddressesForNonTestNodes(uuid, testNode string) []string {
	var addrs []string
	_, nodes := k8stest.GetMsvNodes(uuid)

	for _, node := range nodes {
		if node == testNode {
			continue
		}
		addr, err := k8stest.GetNodeIPAddress(node)
		Expect(err).To(BeNil())
		addrs = append(addrs, *addr)
	}
	return addrs
}
