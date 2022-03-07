package primitive_fault_injection

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	"strings"
	"time"

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
		WithReplicas(c.replicaCount).
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
	var err error
	c.uuid, err = k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, common.VolRawBlock, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", c.pvcName)
	return c
}

// deletePVC will delete pvc
func (c *primitiveFaultInjectionConfig) deletePVC() {
	err := k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", c.pvcName)
}

// createFio will create fio pod and run fio
func (c *primitiveFaultInjectionConfig) createFio() {
	var volFioArgs [][]string

	volFioArgs = append(volFioArgs, []string{
		fmt.Sprintf("--filename=%s", common.FioBlockFilename),
	})
	// Construct argument list for fio
	var podArgs []string

	podArgs = append(podArgs, "--")
	podArgs = append(podArgs, common.GetDefaultFioArguments()...)
	podArgs = append(podArgs, []string{
		"--time_based",
		fmt.Sprintf("--runtime=%d", int(c.duration.Seconds())),
		fmt.Sprintf("--thinktime=%d", int(c.thinkTime.Microseconds())),
		fmt.Sprintf("--status-interval=%d", 30),
	}...,
	)
	for ix, v := range volFioArgs {
		podArgs = append(podArgs, v...)
		podArgs = append(podArgs, fmt.Sprintf("--name=benchtest-%d", ix))
	}
	podArgs = append(podArgs, "&")
	logf.Log.Info("fio", "arguments", podArgs)

	// fio pod container
	podContainer := k8stest.MakeFioContainer(c.fioPodName, podArgs)

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
		WithVolumeDeviceOrMount(common.VolRawBlock).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", c.fioPodName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

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
// the uri of the replica ,
func (c *primitiveFaultInjectionConfig) getNexusDetail() {
	nodeList, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "Failed to get mayastor nodes")

	nexus, replicaNodes := k8stest.GetMsvNodes(c.uuid)
	Expect(nexus).NotTo(Equal(""), "Nexus not found")

	var replicaIPs []string
	// identify the nexus IP address
	nexusIP := ""
	for _, node := range nodeList {
		if node.NodeName == nexus {
			nexusIP = node.IPAddress
		} else {
			for _, replica := range replicaNodes {
				if replica == node.NodeName {
					replicaIPs = append(replicaIPs, node.IPAddress)
					break
				}
			}

		}
	}
	Expect(nexusIP).NotTo(Equal(""), "Nexus IP not found")
	c.nexusNodeIP = nexusIP

	Expect(len(replicaIPs)).To(Equal(c.replicaCount-1), "Expected to find %d non-nexus replicas", c.replicaCount-1)
	c.replicaIPs = replicaIPs

	var nxList []string
	nxList = append(nxList, nexusIP)

	nexusList, _ := mayastorclient.ListNexuses(nxList)
	Expect(len(nexusList)).NotTo(Equal(BeZero()), "Expected to find at least 1 nexus")
	nx := nexusList[0]

	// identify the local replica to be faulted
	nxChildUri := ""
	for _, ch := range nx.Children {
		if strings.HasPrefix(ch.Uri, "bdev:///") {
			nxChildUri = ch.Uri
			break
		}
	}
	Expect(nxChildUri).NotTo(Equal(""), "Could not find nexus replica")
	c.nexusRep = nxChildUri
	msv, err := k8stest.GetMSV(c.uuid)
	Expect(err).ToNot(HaveOccurred(), "failed to retrieve MSV for volume %s", c.uuid)
	c.nexusUuid = msv.Status.Nexus.Uuid

	logf.Log.Info("identified", "nexus", c.nexusNodeIP, "replica1", c.replicaIPs[0], "replica2", c.replicaIPs[1])
}

// Fault the replica hosted on the nexus node
func (c *primitiveFaultInjectionConfig) faultNexusChild() {
	logf.Log.Info("faulting the nexus replica")
	err := mayastorclient.FaultNexusChild(c.nexusNodeIP, c.nexusUuid, c.nexusRep)
	Expect(err).ToNot(HaveOccurred(), "failed to fault local replica")
}

// Validate that all state representations have converged in the expected state (gRPC and CRD)
func (c *primitiveFaultInjectionConfig) verifyVolumeStateOverGrpcAndCrd() {
	logf.Log.Info("Verify crd and grpc status", "msv", c.uuid)
	var statusCount int
	Eventually(func() int {
		msv, err := k8stest.GetMSV(c.uuid)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		Expect(msv).ToNot(BeNil(), "got nil msv for %v", c.uuid)
		nexusChildren := msv.Status.Nexus.Children
		statusCount = 0
		for _, nxChild := range nexusChildren {

			if controlplane.MajorVersion() == 1 {
				if nxChild.State == controlplane.ChildStateOnline() ||
					nxChild.State == controlplane.ChildStateDegraded() {
					statusCount++
				}
				logf.Log.Info("Nexus child state", "uri", nxChild.Uri, "child state", nxChild.State)
			} else {
				Expect(controlplane.MajorVersion).Should(Equal(1), "unsupported control plane version %d/n", controlplane.MajorVersion())
			}

		}
		return statusCount
	},
		defTimeoutSecs,
		"6s",
	).Should(Equal(3), "Nexus children are not in online or degraded state")

	nodeList, err := k8stest.GetNodeLocs()
	Expect(err).ToNot(HaveOccurred(), "Failed to get mayastor nodes")

	// identify the nexus IP address
	var nexusIP []string
	for _, node := range nodeList {
		nexusIP = append(nexusIP, node.IPAddress)
	}
	Expect(len(nexusIP)).NotTo(Equal(BeZero()), "failed to get Nexus IPs")

	nexusList, _ := mayastorclient.ListNexuses(nexusIP)
	Expect(len(nexusList)).NotTo(Equal(BeZero()), "Expected to find at least 1 nexus")
	nx := nexusList[0]

	for _, ch := range nx.Children {
		Expect(ch.State).NotTo(Equal(1), "Nexus child is not online")
	}

}

// verify status of IO after fault injection
func (c *primitiveFaultInjectionConfig) verifyUninterruptedIO() {
	logf.Log.Info("Verify status", "pod", c.fioPodName)
	var fioPodPhase coreV1.PodPhase
	var err error
	var status bool
	Eventually(func() bool {
		status = k8stest.IsPodRunning(c.fioPodName, common.NSDefault)
		return status
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	if !status {
		// check pod phase
		fioPodPhase, err = k8stest.CheckPodContainerCompleted(c.fioPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "Failed to get %s pod phase ", c.fioPodName)
	}
	if fioPodPhase == coreV1.PodSucceeded {
		logf.Log.Info("pod", "name", c.fioPodName, "phase", fioPodPhase)
	} else {
		Expect(status).To(Equal(true), "fio pod %s phase is %v", c.fioPodName, fioPodPhase)
	}
}

// check volume status
func (c *primitiveFaultInjectionConfig) verifyMsvStatus() {
	logf.Log.Info("Verify msv", "uuid", c.uuid)
	namespace := common.NSDefault
	volName := c.pvcName
	pvc, getPvcErr := k8stest.GetPVC(volName, namespace)
	Expect(getPvcErr).To(BeNil(), "Failed to get PVC %s", volName)
	Expect(pvc).ToNot(BeNil())

	// Wait for the PVC to be bound.
	Eventually(func() coreV1.PersistentVolumeClaimPhase {
		phase, err := k8stest.GetPvcStatusPhase(volName, namespace)
		Expect(err).ToNot(HaveOccurred(), "failed to get pvc %s phase", volName)
		return phase
	},
		defTimeoutSecs, // timeout
		"5s",           // polling interval
	).Should(Equal(coreV1.ClaimBound))

	// Refresh the PVC contents, so that we can get the PV name.
	pvc, getPvcErr = k8stest.GetPVC(volName, namespace)
	Expect(getPvcErr).To(BeNil())
	Expect(pvc).ToNot(BeNil())

	// Wait for the PV to be provisioned
	Eventually(func() *coreV1.PersistentVolume {
		pv, getPvErr := k8stest.GetPV(pvc.Spec.VolumeName)
		if getPvErr != nil {
			return nil
		}
		return pv

	},
		defTimeoutSecs, // timeout
		"5s",           // polling interval
	).Should(Not(BeNil()))

	// Wait for the PV to be bound.
	Eventually(func() coreV1.PersistentVolumePhase {
		phase, err := k8stest.GetPvStatusPhase(pvc.Spec.VolumeName)
		Expect(err).ToNot(HaveOccurred(), "failed to get pv %s phase", pvc.Spec.VolumeName)
		return phase
	},
		defTimeoutSecs, // timeout
		"5s",           // polling interval
	).Should(Equal(coreV1.VolumeBound))

	Eventually(func() *common.MayastorVolume {
		msv, err := k8stest.GetMSV(string(pvc.ObjectMeta.UID))
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		return msv
	},
		defTimeoutSecs,
		"5s",
	).Should(Not(BeNil()))

	Eventually(func() string {
		msv, err := k8stest.GetMSV(string(pvc.ObjectMeta.UID))
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		return msv.Status.State
	},
		defTimeoutSecs,
		"5s",
	).Should(Equal(controlplane.VolStateDegraded()))

}

// use the e2e-agent run on each non-nexus node:
//    for each non-nexus replica node
//        nvme connect to its own target
//        checksum /dev/nvme0n1p2
//        disconnect
//    compare the checksum results, they should match
func (c *primitiveFaultInjectionConfig) dataIntegrityCheck() {

	// the first replica checksummed from the second node
	replicas, err := mayastorclient.ListReplicas([]string{c.replicaIPs[0]})
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(len(replicas)).To(Equal(1), "Expected to find 1 replica")
	uri := replicas[0].Uri
	logf.Log.Info("uri", "uri", uri)
	firstchecksum, err := k8stest.ChecksumReplica(c.replicaIPs[0], c.replicaIPs[0], uri, 60)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// the second replica checksummed from the first node
	replicas, err = mayastorclient.ListReplicas([]string{c.replicaIPs[1]})
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(len(replicas)).To(Equal(1), "Expected to find 1 replica")
	uri = replicas[0].Uri
	logf.Log.Info("uri", "uri", uri)
	secondchecksum, err := k8stest.ChecksumReplica(c.replicaIPs[1], c.replicaIPs[1], uri, 60)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// verify that they match
	logf.Log.Info("match", "first", firstchecksum, "this", secondchecksum)
	Expect(secondchecksum).To(Equal(firstchecksum), "checksums differ")
}

// patch msv with existing replication factor minus one
func (c *primitiveFaultInjectionConfig) waitForFioPodCompletion() {
	err := k8stest.WaitPodComplete(c.fioPodName, sleepTime, int(c.timeout))
	Expect(err).ToNot(HaveOccurred(), "Failed to check %s pod completion status", c.fioPodName)
}

// verify faulted replica
func (c *primitiveFaultInjectionConfig) verifyFaultedReplica() {
	var onlineCount, faultedCount, degradedCount, otherCount int
	t0 := time.Now()
	for ix := 0; ix < patchTimeout; ix += patchSleepTime {
		time.Sleep(time.Second * patchSleepTime)
		msv, err := k8stest.GetMSV(c.uuid)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		Expect(msv).ToNot(BeNil(), "got nil msv for %v", c.uuid)
		onlineCount = 0
		faultedCount = 0
		degradedCount = 0
		otherCount = 0

		for _, child := range msv.Status.Nexus.Children {
			if child.State == controlplane.ChildStateFaulted() {
				faultedCount++
			} else if child.State == controlplane.ChildStateOnline() {
				onlineCount++
			} else if child.State == controlplane.ChildStateDegraded() {
				degradedCount++
			} else {
				logf.Log.Info("Children state other then faulted and online", "child.State", child.State)
				otherCount++
			}
		}
		logf.Log.Info("Replica state", "Faulted", faultedCount, "Online", onlineCount, "Degraded", degradedCount, "other", otherCount)
		if faultedCount == 1 && otherCount == 0 && onlineCount != 0 {
			break
		} else if degradedCount == 1 && otherCount == 0 && onlineCount != 0 {
			break
		}
	}
	Expect(otherCount).To(BeZero(), "Got at least one children state other then faulted or online")
	logf.Log.Info("MSV sync waiting time", "seconds", time.Since(t0))
}
