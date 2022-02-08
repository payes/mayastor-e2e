package etcd_inaccessibility

import (
	"fmt"
	"mayastor-e2e/common"
	e2e_agent "mayastor-e2e/common/e2e-agent"
	"mayastor-e2e/common/k8stest"

	coreV1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var fioWriteParams = []string{
	"--name=benchtest",
	"--numjobs=1",
	"--direct=1",
	"--rw=randwrite",
	"--ioengine=libaio",
	"--bs=4k",
	"--iodepth=16",
	"--verify=crc32",
}

func DisablePoolDeviceAtNode(nodeName string, device string) {
	logf.Log.Info("disabling device", "nodeName", nodeName, "device", device)
	pIpAddress, err := k8stest.GetNodeIPAddress(nodeName)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	DisablePoolDeviceAtIp(*pIpAddress, device)
}

func DisablePoolDeviceAtIp(ipAddress string, device string) {
	logf.Log.Info("disabling device", "address", ipAddress, "device", device)
	_, err := e2e_agent.ControlDevice(ipAddress, device, "offline")
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func EnablePoolDeviceAtIp(ipAddress string, device string) {
	logf.Log.Info("enabling device", "address", ipAddress, "device", device)
	_, err := e2e_agent.ControlDevice(ipAddress, device, "running")
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func EnablePoolDeviceAtNode(nodeName string, device string) {
	logf.Log.Info("enabling device", "NodeName", nodeName, "device", device)
	pIpAddress, err := k8stest.GetNodeIPAddress(nodeName)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	EnablePoolDeviceAtIp(*pIpAddress, device)
}

func GetNonNexusNodes(uid string) []string {
	var nonNexusNodes []string
	nexusNode, nodes := k8stest.GetMsvNodes(uid)
	logf.Log.Info("GetNonNexusNodes", "nexusNode", nexusNode, "nodes", nodes)

	for _, node := range nodes {
		if node == nexusNode {
			continue
		}
		nonNexusNodes = append(nonNexusNodes, node)
	}
	logf.Log.Info("GetNonNexusNodes", "nexusNode", nexusNode, "nonNexusNodes", nonNexusNodes)
	return nonNexusNodes
}

func (c *inaccessibleEtcdTestConfig) createSC() {

	err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithLocal(true).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.scName)
}

func (c *inaccessibleEtcdTestConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

func (c *inaccessibleEtcdTestConfig) createPVC() string {
	pvc, err := k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, c.volType, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", c.pvcName)
	return pvc
}

func (c *inaccessibleEtcdTestConfig) deletePVC() {
	err := k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", c.pvcName)
}

func (c *inaccessibleEtcdTestConfig) createFioPod() {

	var args = []string{
		"--",
	}
	args = append(args, fmt.Sprintf("--filename=%s", common.FioBlockFilename))

	args = append(args, fioWriteParams...)

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
