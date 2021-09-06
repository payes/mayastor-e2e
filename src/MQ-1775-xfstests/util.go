package xfstests

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// createSC will create storageclass
func (c *xfsTestConfig) createSC() {
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
func (c *xfsTestConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createPVCs will create pvc
func (c *xfsTestConfig) createPVCs() *xfsTestConfig {
	for _, pvcName := range c.pvcNames {
		c.uuids = append(c.uuids, k8stest.MkPVC(c.pvcSize, pvcName, c.scName, common.VolRawBlock, common.NSDefault))
	}
	return c
}

// deletePVC will delete pvc
func (c *xfsTestConfig) deletePVC() {
	for _, pvcName := range c.pvcNames {
		k8stest.RmPVC(pvcName, c.scName, common.NSDefault)
	}
}

// createXFSTest will create xfstestpod and run xfstest
func (c *xfsTestConfig) createXFSTestPod() {
	// Construct argument list for xfstest
	var podArgs []string

	podArgs = append(podArgs, c.devicePaths...)
	logf.Log.Info("xfstest", "arguments", podArgs)

	// xfstestpod container
	podContainer := k8stest.MakeXFSTestContainer(c.xfstestPodName, podArgs)

	// volume claim details
	testVol := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: c.pvcNames[0],
			},
		},
	}

	scratchVol := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: c.pvcNames[1],
			},
		},
	}

	podObj, err := k8stest.NewPodBuilder().
		WithName(c.xfstestPodName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(testVol).
		WithVolume(scratchVol).
		WithVolumeDeviceOrMount(common.VolRawBlock).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating xfstestpod definition %s", c.xfstestPodName)
	Expect(podObj).ToNot(BeNil(), "failed to generate xfstestpod definition")

	// Create xfstestpod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating xfstestpod %s", c.xfstestPodName)
	// Wait for the xfstestPod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(c.xfstestPodName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
}

// delete xfstestpod
func (c *xfsTestConfig) deleteXFSTest() {
	// Delete the xfstestpod
	err := k8stest.DeletePod(c.xfstestPodName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete xfstestpod")
}

// verify status of IO after fault injection
func (c *xfsTestConfig) verifyUninterruptedIO() {
	logf.Log.Info("Verify status", "pod", c.xfstestPodName)
	var xfstestPodPhase coreV1.PodPhase
	var err error
	var status bool
	Eventually(func() bool {
		status = k8stest.IsPodRunning(c.xfstestPodName, common.NSDefault)
		return status
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	if !status {
		// check pod phase
		xfstestPodPhase, err = k8stest.CheckPodContainerCompleted(c.xfstestPodName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "Failed to get %s pod phase ", c.xfstestPodName)
	}
	if xfstestPodPhase == coreV1.PodSucceeded {
		logf.Log.Info("pod", "name", c.xfstestPodName, "phase", xfstestPodPhase)
	} else {
		Expect(status).To(Equal(true), "xfstestpod %s phase is %v", c.xfstestPodName, xfstestPodPhase)
	}
}

// patch msv with existing replication factor minus one
func (c *xfsTestConfig) waitForXFSTestPodCompletion() {
	err := k8stest.WaitPodComplete(c.xfstestPodName, sleepTime, podCompletionTimeout)
	Expect(err).ToNot(HaveOccurred(), "Failed to check %s pod completion status", c.xfstestPodName)
}
