package maximum_vols_io

import (
	"fmt"
	"strings"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// createSC will create storageclass
func (c *maxVolConfig) createSC() {
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
func (c *maxVolConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createPVC will create pvc
func (c *maxVolConfig) createPVC() *maxVolConfig {
	// Create the volumes
	for _, pvc := range c.pvcNames {
		uid, err := k8stest.MkPVC(c.pvcSize, pvc, c.scName, common.VolFileSystem, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "failed to create pvc %s", pvc)
		c.uuid = append(c.uuid, uid)
	}

	return c
}

// deletePVC will delete all pvc
func (c *maxVolConfig) deletePVC() {
	for _, pvc := range c.pvcNames {
		err := k8stest.RmPVC(pvc, c.scName, common.NSDefault)
		Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", pvc)
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
			Name:            podName,
			Image:           common.GetFioImage(),
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
					fmt.Sprintf("--name=benchtest-%d", ix),
					fmt.Sprintf("--filename=/volume-%s/fio-test-file", c.pvcNames[pvcCount]),
				})
			} else {
				device := coreV1.VolumeDevice{
					Name:       fmt.Sprintf("ms-volume-%s", c.pvcNames[pvcCount]),
					DevicePath: fmt.Sprintf("/dev/sdm-%s", c.pvcNames[pvcCount]),
				}
				volDevices = append(volDevices, device)
				volFioArgs = append(volFioArgs, []string{
					fmt.Sprintf("--name=benchtest-%d", ix),
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

		// Construct argument list for fio to run a single instance of fio,
		// with multiple jobs, one for each volume.
		var podArgs []string

		// 1) directives for all fio jobs
		podArgs = append(podArgs, "---")
		podArgs = append(podArgs, "fio")
		podArgs = append(podArgs, common.GetDefaultFioArguments()...)
		podArgs = append(podArgs, []string{
			fmt.Sprintf("--size=%dm", common.DefaultFioSizeMb),
		}...,
		)

		// 2) per volume directives (filename, size, and testname)
		for _, v := range volFioArgs {
			podArgs = append(podArgs, v...)
		}
		logf.Log.Info("fio", "commandline", strings.Join(podArgs[1:], " "))

		podArgs = append(podArgs, "&")
		logf.Log.Info("pod", "arguments", podArgs)
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
