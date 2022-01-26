package last_replica

import (
	"time"

	"mayastor-e2e/common"
	e2e_agent "mayastor-e2e/common/e2e-agent"
	"mayastor-e2e/common/k8stest"
	"mayastor-e2e/common/mayastorclient"

	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "120s"

func setupTestVolAndPod(ctx *testContext) {
	err := k8stest.NewScBuilder().
		WithName(ctx.scName).
		WithReplicas(ctx.replicaCount).
		WithProtocol(ctx.protocol).
		WithNamespace(common.NSDefault).
		WithVolumeBindingMode(ctx.volBindingMode).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "failed to create storage class %s", ctx.scName)

	// Create the volume
	ctx.volUid = k8stest.MkPVC(ctx.pvcSize, ctx.pvcName, ctx.scName, ctx.volumeType, common.NSDefault)
	logf.Log.Info("Volume", "uid", ctx.volUid)

	// volume claim details
	volume := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: ctx.pvcName,
			},
		},
	}

	container := k8stest.MakeFioContainer(ctx.podName, ctx.fioArgs)
	podObj, err := k8stest.NewPodBuilder().
		WithName(ctx.podName).
		WithNamespace(common.NSDefault).
		WithContainer(container).
		WithVolume(volume).
		WithVolumeDeviceOrMount(ctx.volumeType).
		Build()
	Expect(err).ToNot(HaveOccurred(), "failed to build test pod")
	pod, err := k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to create test pod")
	Expect(pod).ToNot(BeNil(), "go nil pointer to test pod ")

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(ctx.podName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	logf.Log.Info("fio test pod is running.")
}

func faultReplica(ctx *testContext) {
	//verify if nexus is created or not
	msv, nexusErr := k8stest.GetMSV(ctx.volUid)
	Expect(nexusErr).ToNot(HaveOccurred(), "failed to retrieve nexus details")
	logf.Log.Info("msv", "Children", msv.Status.Nexus.Children)

	ipaddrs := k8stest.GetMayastorNodeIPAddresses()
	for _, ipaddr := range ipaddrs {
		nexuses, err := mayastorclient.ListNexuses([]string{ipaddr})
		if err == nil {
			for _, nexus := range nexuses {
				Expect(len(nexus.Children)).ToNot(BeZero())
				if nexus.DeviceUri == msv.Status.Nexus.DeviceUri {
					logf.Log.Info("Faulting",
						"ip", ipaddr,
						"nexusuuid", nexus.Uuid,
						"childuri", nexus.Children[0].Uri,
					)
					err = mayastorclient.FaultNexusChild(ipaddr, nexus.Uuid, nexus.Children[0].Uri)
					if err != nil {
						logf.Log.Info("Failed to fault the nexus replica")
					}
				}
			}
		}
	}
}

func checkChildren(ctx *testContext) (int, error) {
	nc := -1
	//verify that the volume has children
	msv, err := k8stest.GetMSV(ctx.volUid)
	if err == nil {
		logf.Log.Info("msv", "uid", ctx.volUid, "Children", msv.Status.Nexus.Children)
		nc = len(msv.Status.Nexus.Children)
	}
	return nc, err
}

func assertCheckChildren(ctx *testContext) {
	nc, err := checkChildren(ctx)
	Expect(err).ToNot(HaveOccurred(), "Failed to retrieve nexus children")
	Expect(nc > 0).To(BeTrue())
}

func waitPodComplete(ctx *testContext) (int, error) {
	logf.Log.Info("Waiting for run to complete", "timeout", ctx.fioTimeout)
	tSecs := 0
	var exitCodes []int

	completed := false
	var err error
	for {
		if tSecs > ctx.fioTimeout {
			break
		}
		time.Sleep(1 * time.Second)
		tSecs += 1
		completed, exitCodes, err = k8stest.GetPodCompletion(ctx.podName, common.NSDefault)
		if err != nil {
			logf.Log.Info("CheckPodComplete got error %s", err)
			break
		}
		if completed {
			break
		}
	}
	logf.Log.Info("fio pod", "completed", completed, "waited", tSecs, "exitCodes", exitCodes)
	for _, ec := range exitCodes {
		if ec != 0 {
			return ec, err
		}
	}
	return 0, err
}

func cleanup(ctx *testContext) {
	// Delete the fio pod
	err := k8stest.DeletePod(ctx.podName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred())

	// Delete the volume
	k8stest.RmPVC(ctx.pvcName, ctx.scName, common.NSDefault)

	err = k8stest.RmStorageClass(ctx.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", ctx.scName)
}

func disablePoolDeviceAtIp(ipAddress string, device string) {
	logf.Log.Info("disabling device", "address", ipAddress, "device", device)
	_, err := e2e_agent.ControlDevice(ipAddress, device, "offline")
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func enablePoolDeviceAtIp(ipAddress string, device string) {
	logf.Log.Info("enabling device", "address", ipAddress, "device", device)
	_, err := e2e_agent.ControlDevice(ipAddress, device, "running")
	Expect(err).ToNot(HaveOccurred(), "%v", err)
}

func blipPoolDevices(ctx *testContext) {
	offlinePoolDevices(ctx)
	if ctx.blipSeconds > 0 {
		logf.Log.Info("Sleeping", "secs", ctx.blipSeconds)
		time.Sleep(time.Duration(ctx.blipSeconds) * time.Second)
	}
	onlinePoolDevices(ctx)
}

func offlinePoolDevices(ctx *testContext) {
	ipaddrs := k8stest.GetMayastorNodeIPAddresses()
	for _, ipaddr := range ipaddrs {
		disablePoolDeviceAtIp(ipaddr, ctx.poolDev)
	}
}

func onlinePoolDevices(ctx *testContext) {
	ipaddrs := k8stest.GetMayastorNodeIPAddresses()
	for _, ipaddr := range ipaddrs {
		enablePoolDeviceAtIp(ipaddr, ctx.poolDev)
	}
}
