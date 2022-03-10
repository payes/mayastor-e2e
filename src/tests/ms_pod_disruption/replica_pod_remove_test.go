package ms_pod_disruption

import (
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/e2e_config"

	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Write data to a volume while replicas are being removed and added.
// After each transition, verify every block in the volume is correct
// 1) Write pattern-1 to every block in the volume once and simultaneously remove one replica
// 2) Verify that the volume becomes degraded and the data is correct
// 3) Write pattern-2 while adding a new replica to the volume while
// 4) Verify that the volume becomes healthy and the data is correct
// 5) Write pattern-3 while faulting the replica on the nexus node
// 6) Verify that the volume becomes degraded and the data is correct
// 7) Re enable the first replica, wait for the volume to become healthy
//    Verify that the data is still correct
func (env *DisruptionEnv) PodLossTestWriteContinuously() {
	e2eCfg := e2e_config.GetConfig()

	// 1) Write pattern-1 to every block in the volume once and simultaneously remove one replica
	// Running fio with --do_verify=0, --verify=crc32 and --rw=randwrite means that only writes will occur
	// and no verification reads happen, verification can be done in the next step "off-line"
	logf.Log.Info("about to suppress mayastor on one replica")
	go env.suppressMayastorPodOn(env.replicaToRemove, e2eCfg.MsPodDisruption.UnscheduleDelay)

	logf.Log.Info("writing to the volume")
	err := env.fioWriteOnly(env.fioPodName, "crc32", env.removeThinkTime)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 2) Verify that the volume has become degraded and the data is correct
	// We make the assumption that the volume has had enough time to become faulted
	Expect(getMsvState(env.uuid)).To(Equal(controlplane.VolStateDegraded()), "Unexpected MSV state")
	logf.Log.Info("volume condition", "state", getMsvState(env.uuid))

	// Running fio with --verify=crc32 and --rw=randread means that only reads will occur
	// and verification is performed
	logf.Log.Info("verifying the degraded volume")
	err = fioVerify(env.fioPodName, "crc32")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 3) Write pattern-2 while adding a new replica to the volume while
	// re-enable mayastor on one unused node
	logf.Log.Info("replacing the original replica")
	go env.unsuppressMayastorPodOn(env.unusedNodes[0], e2eCfg.MsPodDisruption.RescheduleDelay)

	// Random writes only. Note the checksum is now md5 and is stored in the block
	// Any blocks that do not get modified will fail the verification run (below)
	// because the stored checksum will still be crc32
	logf.Log.Info("writing to the volume")
	err = env.fioWriteOnly(env.fioPodName, "md5", env.repairThinkTime)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 4) Verify that the volume becomes healthy and the data is correct
	// We make the assumption that the volume has had enough time to be repaired
	Expect(getMsvState(env.uuid)).To(Equal(controlplane.VolStateHealthy()), "Unexpected MSV state")
	logf.Log.Info("volume condition", "state", getMsvState(env.uuid))

	// Verify the data just written.
	logf.Log.Info("verifying the repaired volume")
	err = fioVerify(env.fioPodName, "md5")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 5) Write pattern-3 while faulting the replica on the nexus node
	// remove the replica from the nexus
	go env.faultNexusChild(e2eCfg.MsPodDisruption.UnscheduleDelay)

	// Running fio with --do_verify=0, --verify=sha1 and --rw=randwrite means that only writes will occur
	// and no verification is performed at this point
	logf.Log.Info("writing to the volume")
	err = env.fioWriteOnly(env.fioPodName, "sha1", env.removeThinkTime)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 6) Verify that the volume becomes degraded and the data is correct
	// We make the assumption that the volume has had enough time to become faulted
	Expect(getMsvState(env.uuid)).To(Equal(controlplane.VolStateDegraded()), "Unexpected MSV state")

	// Running fio with --verify=sha1 and --rw=randread means that only reads will occur
	// and verification happens
	// This step reads each block once
	logf.Log.Info("verifying the degraded volume")
	err = fioVerify(env.fioPodName, "sha1")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	logf.Log.Info("restoring the original replica")
	env.unsuppressMayastorPodOn(env.replicaToRemove, 0)

	// 7) verify that the volume becomes healthy again
	Eventually(func() string {
		return getMsvState(env.uuid)
	},
		env.rebuildTimeoutSecs, // timeout
		"1s",                   // polling interval
	).Should(Equal(controlplane.VolStateHealthy()))
	logf.Log.Info("volume condition", "state", getMsvState(env.uuid))

	// Re-verify with the original replica on-line, It it gets any IO the
	// verification will fail because it contains the wrong data.
	logf.Log.Info("verifying the repaired volume")
	err = fioVerify(env.fioPodName, "sha1")
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}
