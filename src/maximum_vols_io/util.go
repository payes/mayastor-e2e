package maximum_vols_io

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	"time"

	coreV1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// createSC will create storageclass
func (c *maxVolConfig) createSC() {
	scObj, err := k8stest.NewScBuilder().
		WithName(c.scName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.protocol).
		WithReplicas(c.replicas).
		WithFileSystemType(c.fsType).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating storage class definition %s", c.scName)
	err = k8stest.CreateSc(scObj)
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", c.scName)
}

// deleteSC will delete storageclass
func (c *maxVolConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createPVC will create pvc
func (c *maxVolConfig) createPVC() *maxVolConfig {
	// Create the volumes
	for _, pvc := range c.pvcNames {
		uid := k8stest.MkPVC(c.pvcSize, pvc, c.scName, common.VolFileSystem, common.NSDefault)
		c.uuid = append(c.uuid, uid)
	}

	return c
}

// deletePVC will delete all pvc
func (c *maxVolConfig) deletePVC() {
	for _, pvc := range c.pvcNames {
		k8stest.RmPVC(pvc, c.scName, common.NSDefault)
	}
}

// createFioPods will create fio pods and run fio concurrently on all mounted volumes
func (c *maxVolConfig) createFioPods() {
	pvcCount := 0

	for _, podName := range c.fioPodNames {
		var volMounts []coreV1.VolumeMount
		var volDevices []coreV1.VolumeDevice
		var volFioArgs [][]string

		// fio pod container
		podContainer := coreV1.Container{
			Name:  podName,
			Image: common.GetFioImage(),
			// Image:           "mayadata/e2e-fio",
			ImagePullPolicy: coreV1.PullAlways,
			Args:            []string{"sleep", "1000000"},
		}
		var volumes []coreV1.Volume
		for ix := 0; ix < c.volCountPerPod; ix++ {
			// volume claim details
			volume := coreV1.Volume{
				Name: fmt.Sprintf("ms-volume-%s", c.pvcNames[pvcCount]),
				VolumeSource: coreV1.VolumeSource{
					PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
						ClaimName: c.pvcNames[pvcCount],
					},
				},
			}
			volumes = append(volumes, volume)
			// volume mount or device
			if c.volType == common.VolFileSystem {
				mount := coreV1.VolumeMount{
					Name:      fmt.Sprintf("ms-volume-%s", c.pvcNames[pvcCount]),
					MountPath: fmt.Sprintf("/volume-%s", c.pvcNames[pvcCount]),
				}
				volMounts = append(volMounts, mount)
				volFioArgs = append(volFioArgs, []string{
					fmt.Sprintf("--filename=/volume-%s/fio-test-file", c.pvcNames[pvcCount]),
					fmt.Sprintf("--size=%dm", common.DefaultFioSizeMb),
				})
			} else {
				device := coreV1.VolumeDevice{
					Name:       fmt.Sprintf("ms-volume-%s", c.pvcNames[pvcCount]),
					DevicePath: fmt.Sprintf("/dev/sdm-%s", c.pvcNames[pvcCount]),
				}
				volDevices = append(volDevices, device)
				volFioArgs = append(volFioArgs, []string{
					fmt.Sprintf("--filename=/dev/sdm-%s", c.pvcNames[pvcCount]),
				})
			}
			pvcCount += 1
		}

		podObj, err := k8stest.NewPodBuilder().
			WithName(podName).
			WithNamespace(common.NSDefault).
			WithRestartPolicy(coreV1.RestartPolicyNever).
			WithContainer(podContainer).
			WithVolumes(volumes).Build()
		Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", podName)
		Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

		switch c.volType {
		case common.VolFileSystem:
			podObj.Spec.Containers[0].VolumeMounts = volMounts
		case common.VolRawBlock:
			podObj.Spec.Containers[0].VolumeDevices = volDevices
		}

		var podArgs []string
		for _, v := range volFioArgs {
			podArgs = append(podArgs, "--")
			podArgs = append(podArgs, common.GetFioArgs()...)
			podArgs = append(podArgs, v...)
			podArgs = append(podArgs, "--time_based")
			podArgs = append(podArgs, fmt.Sprintf("--runtime=%d", int(c.duration.Seconds())))
			podArgs = append(podArgs, fmt.Sprintf("--thinktime=%d", int(c.thinkTime.Microseconds())))
			podArgs = append(podArgs, "&")
		}

		logf.Log.Info("fio", "arguments", podArgs)
		podObj.Spec.Containers[0].Args = podArgs

		// Create first fio pod
		_, err = k8stest.CreatePod(podObj, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", podName)
		// Wait for the fio Pod to transition to running
		Eventually(func() bool {
			return k8stest.IsPodRunning(podName, common.NSDefault)
		},
			defTimeoutSecs,
			"1s",
		).Should(Equal(true))
	}
}

// delete all fip pods
func (c *maxVolConfig) deleteFioPods() {
	for _, podName := range c.fioPodNames {
		// Delete the fio pod
		err := k8stest.DeletePod(podName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "failed to delete fio pod")
	}
}

// check fio pods completion status
func (c *maxVolConfig) checkFioPodsComplete() {
	for _, podName := range c.fioPodNames {
		logf.Log.Info("Waiting for run to complete", "duration", c.duration, "timeout", c.timeout)
		tSecs := 0
		var phase coreV1.PodPhase
		var err error
		for {
			if tSecs > int(c.timeout.Seconds()) {
				break
			}
			time.Sleep(1 * time.Second)
			tSecs += 1
			phase, err = k8stest.CheckPodCompleted(podName, common.NSDefault)
			Expect(err).To(BeNil(), "CheckPodComplete got error %s", err)
			if phase != coreV1.PodRunning {
				break
			}
		}
		Expect(phase == coreV1.PodSucceeded).To(BeTrue(), "fio pod phase is %s", phase)
		logf.Log.Info("fio completed", "duration", tSecs)
	}
}
