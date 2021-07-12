package primitive_fault_injection

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	"strings"

	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// createSC will create storageclass
func (c *primitiveFaultInjectionConfig) createSC() {
	err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithFileSystemType(c.fsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.scName)
}

// deleteSC will delete storageclass
func (c *primitiveFaultInjectionConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createPVCs will create pvc
func (c *primitiveFaultInjectionConfig) createPVC() *primitiveFaultInjectionConfig {
	c.uuid = k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, common.VolFileSystem, common.NSDefault)
	return c
}

// deletePVC will delete pvc
func (c *primitiveFaultInjectionConfig) deletePVC() {
	k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
}

// createFio will create fio pod and run fio
func (c *primitiveFaultInjectionConfig) createFio() {

	// fio pod container
	podContainer := coreV1.Container{
		Name:            c.fioPodName,
		Image:           common.GetFioImage(),
		ImagePullPolicy: coreV1.PullAlways,
		Args:            []string{"sleep", "1000000"},
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
		WithName(c.fioPodName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(common.VolFileSystem).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", c.fioPodName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

	// Construct argument list for fio to run a single instance of fio,
	// with multiple jobs, one for each volume.
	var podArgs []string

	podArgs = append(podArgs, "--")
	podArgs = append(podArgs, common.GetDefaultFioArguments()...)
	podArgs = append(podArgs, []string{
		"--time_based",
		fmt.Sprintf("--runtime=%d", int(c.duration.Seconds())),
		fmt.Sprintf("--thinktime=%d", int(c.thinkTime.Microseconds())),
	}...,
	)
	logf.Log.Info("fio", "arguments", podArgs)
	podObj.Spec.Containers[0].Args = podArgs
	// Create fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", c.fioPodName)
	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(c.fioPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
}

// delete fio pod
func (c *primitiveFaultInjectionConfig) deleteFio() {
	// Delete the fio pod
	err := k8stest.DeletePod(c.fioPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete fio pod")
}

// Identify the nexus IP address,
// the uri of the replica local to the nexus,
// and the non-nexus node hosting a replica.
func (c *primitiveFaultInjectionConfig) getNexusDetail() {
	nodeList, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "Failed to get mayastor nodes")

	nexus, _ := k8stest.GetMsvNodes(c.uuid)
	Expect(nexus).NotTo(Equal(""), "failed to get Nexus")

	// identify the nexus IP address
	nexusIP := ""
	for _, node := range nodeList {
		if node.NodeName == nexus {
			nexusIP = node.IPAddress
			break
		}
	}
	Expect(nexusIP).NotTo(Equal(""), "failed to get Nexus IP")
	c.nexusIP = nexusIP

	var nxList []string
	nxList = append(nxList, nexusIP)

	nexusList, _ := mayastorclient.ListNexuses(nxList)
	Expect(len(nexusList)).To(Equal(1), "Expected to find 1 nexus")
	nx := nexusList[0]

	// identify the replica local to the nexus
	nxChild := ""
	for _, ch := range nx.Children {
		if strings.HasPrefix(ch.Uri, "bdev:///") {
			Expect(nxChild).To(Equal(""), "More than 1 nexus local replica found")
			nxChild = ch.Uri
			break
		}
	}
	Expect(nxChild).NotTo(Equal(""), "Could not find nexus local replica")
	c.nexusLocalRep = nxChild
	logf.Log.Info("Nexus details", "nexus IP", c.nexusIP, "local replica", c.nexusLocalRep)
}

// Fault the replica hosted on the nexus node
func (c *primitiveFaultInjectionConfig) faultNexusChild() {
	logf.Log.Info("faulting the nexus replica")
	err := mayastorclient.FaultNexusChild(c.nexusIP, c.uuid, c.nexusLocalRep)
	Expect(err).ToNot(HaveOccurred(), "failed to fault local replica")
}

// Validate that all state representations have converged in the expected state (gRPC and CRD)
func (c *primitiveFaultInjectionConfig) verifyVolumeStateOverGrpcAndCrd() {
	//msv := k8stest.GetMSV(c.uuid)
	// TODO verify msv state
}

// verify status of IO after fault injection
func (c *primitiveFaultInjectionConfig) verifyUninterruptedIO() {
	status := k8stest.IsPodRunning(c.fioPodName, common.NSDefault)
	Expect(status).To(Equal(true), "fio pod %s not running", c.fioPodName)
}
