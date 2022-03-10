package ms_pod_disruption

import (
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/e2e_config"
	"strings"

	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Write data to a 2-replica volume while replicas are being removed and added.
// After each transition, verify every block in the volume is correct
// 1) Write pattern-1 to every block in the volume once and simultaneously offline one replica device
// 2) Verify that the volume becomes degraded and the data is correct (now only on nexus)
// 3) Write pattern-2 while adding a new replica to the volume.
// 4) Verify that the volume becomes healthy and the data is correct (nexus and new replica)
// 5) Write pattern-3 while offlining the backing device on the nexus node
// 6) Verify that the volume becomes degraded and the data is correct (now only on new replica)
// 7) Online the first replica's device, wait for the volume to become healthy
//    Verify that the data is still correct (new replica and first replica, not nexus)
func (env *DisruptionEnv) DeviceLossTest() {
	e2eCfg := e2e_config.GetConfig()

	poolDevice := e2eCfg.PoolDevice
	Expect(strings.HasPrefix(poolDevice, "/dev/")).To(BeTrue(), "unexpected pool spec %s", poolDevice)
	poolDevice = poolDevice[5:]

	// 1) Write pattern-1 to every block in the volume once and simultaneously offline
	// the backing device on one non-nexus node.
	// Running fio with --do_verify=0, --verify=crc32 and --rw=randwrite means that only writes will occur
	// and no verification reads happen, verification can be done in the next step "off-line".
	logf.Log.Info("about to disable the backing device on one replica")
	env.disablePoolDeviceAtNode(env.replicaToRemove, poolDevice, e2eCfg.MsPodDisruption.UnscheduleDelay)

	logf.Log.Info("writing to the volume")
	err := env.fioWriteOnly(env.fioPodName, "crc32", env.removeThinkTime)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 2) Verify that the volume has become degraded and the data is correct
	// We make the assumption that the volume has had enough time to become faulted
	Expect(getMsvState(env.uuid)).To(Equal(controlplane.VolStateDegraded()), "Unexpected MSV state")
	logf.Log.Info("volume condition", "state", getMsvState(env.uuid))

	// Running fio with --verify=crc32 and --rw=randread means that only reads will occur
	// and verification is performed.
	logf.Log.Info("verifying the degraded volume")
	err = fioVerify(env.fioPodName, "crc32")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 3) Write pattern-2 while adding a replacement replica to the volume by
	// re-enabling mayastor on one unused node.
	logf.Log.Info("making a new replica available")
	go env.unsuppressMayastorPodOn(env.unusedNodes[0], e2eCfg.MsPodDisruption.RescheduleDelay)

	// Random writes only. Note the checksum is now md5 and is stored in the block
	// Any blocks that do not get modified will fail the verification run (below)
	// because the stored checksum will still be crc32.
	logf.Log.Info("writing to the volume")
	err = env.fioWriteOnly(env.fioPodName, "md5", env.repairThinkTime)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 4) Verify that the volume becomes healthy and the data is correct
	// We make the assumption that the volume has had enough time to be repaired
	Expect(getMsvState(env.uuid)).To(Equal(controlplane.VolStateHealthy()), "Unexpected MSV state")
	logf.Log.Info("volume condition", "state", getMsvState(env.uuid))

	// Verify the data just written (nexus and new replica).
	logf.Log.Info("verifying the repaired volume")
	err = fioVerify(env.fioPodName, "md5")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 5) Write pattern-3 and offline the nexus node's backing device
	env.disablePoolDeviceAtIp(env.nexusIP, poolDevice, 0)

	// Running fio with --do_verify=0, --verify=sha1 and --rw=randwrite means that only writes will occur
	// and no verification is performed at this point.
	logf.Log.Info("writing to the volume")
	err = env.fioWriteOnly(env.fioPodName, "sha1", env.removeThinkTime)
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// 6) Verify that the volume has become degraded and the data is correct (new replica only).
	// We make the assumption that the volume has had enough time to become faulted.
	Expect(getMsvState(env.uuid)).To(Equal(controlplane.VolStateDegraded()), "Unexpected MSV state")

	// Running fio with --verify=sha1 and --rw=randread means that only reads will occur
	// and verification is carried out.
	// This step reads each block once.
	logf.Log.Info("verifying the degraded volume")
	err = fioVerify(env.fioPodName, "sha1")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	logf.Log.Info("online the first replica's backing device")
	env.enablePoolDeviceAtNode(env.replicaToRemove, poolDevice, 0)

	// 7) verify that the volume becomes healthy again (first replica and new replica, not the nexus)
	Eventually(func() string {
		return getMsvState(env.uuid)
	},
		env.rebuildTimeoutSecs, // timeout
		"1s",                   // polling interval
	).Should(Equal(controlplane.VolStateHealthy()))
	logf.Log.Info("volume condition", "state", getMsvState(env.uuid))

	// Re-verify with the original replica on-line, If it gets any IO the
	// verification will fail because it contains old data.
	logf.Log.Info("verifying the repaired volume")
	err = fioVerify(env.fioPodName, "sha1")
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	// Tidy up, re-enable the nexus node's backing device.
	env.enablePoolDeviceAtIp(env.nexusIP, poolDevice, 0)
}
