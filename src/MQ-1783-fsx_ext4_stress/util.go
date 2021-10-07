package fsx_ext4_stress

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/ctlpln"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// createSC will create storageclass
func (c *fsxExt4StressConfig) createSC() {
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
func (c *fsxExt4StressConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createPVCs will create pvc
func (c *fsxExt4StressConfig) createPVC() *fsxExt4StressConfig {
	c.uuid = k8stest.MkPVC(c.pvcSize, c.pvcName, c.scName, common.VolRawBlock, common.NSDefault)
	return c
}

// deletePVC will delete pvc
func (c *fsxExt4StressConfig) deletePVC() {
	k8stest.RmPVC(c.pvcName, c.scName, common.NSDefault)
}

// createFsx will create fsx pod and run fsx
func (c *fsxExt4StressConfig) createFsx() {
	// Construct argument list for fsx
	var podArgs []string

	podArgs = append(podArgs, c.devicePath, c.fileSystemType, strconv.Itoa(c.numberOfOperation))
	logf.Log.Info("fsx", "arguments", podArgs)

	// fsx pod container
	podContainer := k8stest.MakeFsxContainer(c.fsxPodName, podArgs)

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
		WithName(c.fsxPodName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(common.VolRawBlock).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fsx pod definition %s", c.fsxPodName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fsx pod definition")

	// Create fsx pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fsx pod %s", c.fsxPodName)
	// Wait for the fsx Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(c.fsxPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
}

// delete fsx pod
func (c *fsxExt4StressConfig) deleteFsx() {
	// Delete the fsx pod
	err := k8stest.DeletePod(c.fsxPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete fsx pod")
}

// Identify the nexus IP address,
// the uri of the replica ,
func (c *fsxExt4StressConfig) getNexusDetail() {
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

	logf.Log.Info("identified", "nexus", c.nexusNodeIP, "replica1", c.replicaIPs[0], "replica2", c.replicaIPs[1])
}

// Fault the replica hosted on the nexus node
func (c *fsxExt4StressConfig) faultNexusChild() {
	logf.Log.Info("faulting the nexus replica")
	err := mayastorclient.FaultNexusChild(c.nexusNodeIP, c.uuid, c.nexusRep)
	Expect(err).ToNot(HaveOccurred(), "failed to fault local replica")
}

// Validate that all state representations have converged in the expected state (gRPC and CRD)
func (c *fsxExt4StressConfig) verifyVolumeStateOverGrpcAndCrd() {
	logf.Log.Info("Verify crd and grpc status", "msv", c.uuid)
	msv, err := k8stest.GetMSV(c.uuid)
	Expect(err).ToNot(HaveOccurred(), "%v", err)
	Expect(msv).ToNot(BeNil(), "got nil msv for %v", c.uuid)
	nexusChildren := msv.Status.Nexus.Children
	for _, nxChild := range nexusChildren {
		Expect(nxChild.State).Should(Equal(ctlpln.ChildStateOnline()), "Nexus child  is not online")
	}

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
func (c *fsxExt4StressConfig) verifyUninterruptedIO() {
	logf.Log.Info("Verify status", "pod", c.fsxPodName)
	var fsxPodPhase coreV1.PodPhase
	var err error
	var status bool
	Eventually(func() bool {
		status = k8stest.IsPodRunning(c.fsxPodName, common.NSDefault)
		return status
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	if !status {
		// check pod phase
		fsxPodPhase, err = k8stest.CheckPodContainerCompleted(c.fsxPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "Failed to get %s pod phase ", c.fsxPodName)
	}
	if fsxPodPhase == coreV1.PodSucceeded {
		logf.Log.Info("pod", "name", c.fsxPodName, "phase", fsxPodPhase)
	} else {
		Expect(status).To(Equal(true), "fsx pod %s phase is %v", c.fsxPodName, fsxPodPhase)
	}
}

// patch msv with existing replication factor minus one
func (c *fsxExt4StressConfig) patchMsvReplica() {
	err := k8stest.SetMsvReplicaCount(c.uuid, c.replicaCount-1)
	Expect(err).ToNot(HaveOccurred(), "Failed to patch Mayastor volume %s", c.uuid)
}

// check volume status
func (c *fsxExt4StressConfig) verifyMsvStatus() {
	logf.Log.Info("Verify msv", "uuid", c.uuid)
	namespace := common.NSDefault
	volName := c.pvcName
	pvc, getPvcErr := k8stest.GetPVC(volName, namespace)
	Expect(getPvcErr).To(BeNil(), "Failed to get PVC %s", volName)
	Expect(pvc).ToNot(BeNil())

	// Wait for the PVC to be bound.
	Eventually(func() coreV1.PersistentVolumeClaimPhase {
		return k8stest.GetPvcStatusPhase(volName, namespace)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
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
		"1s",           // polling interval
	).Should(Not(BeNil()))

	// Wait for the PV to be bound.
	Eventually(func() coreV1.PersistentVolumePhase {
		return k8stest.GetPvStatusPhase(pvc.Spec.VolumeName)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(coreV1.VolumeBound))

	Eventually(func() *common.MayastorVolume {
		msv, err := k8stest.GetMSV(string(pvc.ObjectMeta.UID))
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		return msv
	},
		defTimeoutSecs,
		"1s",
	).Should(Not(BeNil()))

}

// patch msv with existing replication factor minus one
func (c *fsxExt4StressConfig) waitForFsxPodCompletion() {
	err := k8stest.WaitPodComplete(c.fsxPodName, sleepTime, podCompletionTimeout)
	Expect(err).ToNot(HaveOccurred(), "Failed to check %s pod completion status", c.fsxPodName)
}

// verify faulted replica
func (c *fsxExt4StressConfig) verifyFaultedReplica() {
	var onlineCount, faultedCount, otherCount int
	t0 := time.Now()
	for ix := 0; ix < patchTimeout; ix += patchSleepTime {
		time.Sleep(time.Second * patchSleepTime)
		msv, err := k8stest.GetMSV(c.uuid)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		Expect(msv).ToNot(BeNil(), "got nil msv for %v", c.uuid)
		onlineCount = 0
		faultedCount = 0
		otherCount = 0
		for _, child := range msv.Status.Nexus.Children {
			if child.State == ctlpln.ChildStateFaulted() {
				faultedCount++
			} else if child.State == ctlpln.ChildStateOnline() {
				onlineCount++
			} else {
				logf.Log.Info("Children state other then faulted and online", "child.State", child.State)
				otherCount++
			}
		}
		logf.Log.Info("Replica state", "faulted", faultedCount, "online", onlineCount, "other", otherCount)
		if faultedCount == 1 && otherCount == 0 && onlineCount != 0 {
			break
		}
	}
	Expect(otherCount).To(BeZero(), "Got at least one children state other then faulted or online")
	logf.Log.Info("MSV sync waiting time", "seconds", time.Since(t0))
}

// verify updated replica state
func (c *fsxExt4StressConfig) verifyUpdatedReplica() {
	var onlineCount, faultedCount, otherCount int
	t0 := time.Now()
	for ix := 0; ix < patchTimeout; ix += patchSleepTime {
		time.Sleep(time.Second * patchSleepTime)
		msv, err := k8stest.GetMSV(c.uuid)
		Expect(err).ToNot(HaveOccurred(), "%v", err)
		Expect(msv).ToNot(BeNil(), "got nil msv for %v", c.uuid)
		onlineCount = 0
		faultedCount = 0
		otherCount = 0
		for _, child := range msv.Status.Nexus.Children {
			if child.State == ctlpln.ChildStateFaulted() {
				faultedCount++
			} else if child.State == ctlpln.ChildStateOnline() {
				onlineCount++
			} else {
				logf.Log.Info("Children state other then faulted and online", "child.State", child.State)
				otherCount++
			}
		}
		logf.Log.Info("Replica state", "faulted", faultedCount, "online", onlineCount, "other", otherCount)
		if faultedCount == 0 && otherCount == 0 && onlineCount == msv.Spec.ReplicaCount {
			break
		}
	}
	Expect(otherCount).To(BeZero(), "Got at least one children state other then faulted or online")
	Expect(faultedCount).To(BeZero(), "Got at least one children state as faulted")
	logf.Log.Info("MSV sync waiting time", "seconds", time.Since(t0))
}
