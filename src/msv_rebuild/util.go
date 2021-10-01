package msv_rebuild

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"

	coreV1 "k8s.io/api/core/v1"

	. "github.com/onsi/gomega"
)

// createStorageClass will create storageclass
func createStorageClass(scName string, replicas int) {
	err := k8stest.NewScBuilder().
		WithName(scName).
		WithNamespace(common.NSDefault).
		WithProtocol(common.ShareProtoNvmf).
		WithReplicas(replicas).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "Creating storage class %s", scName)
}

// createFioPod created fio pod obj and create fio pod
func createFioPod(fioPodName string, pvcName string, durationSecs int, volSize int) error {
	args := []string{
		"--",
		"--time_based",
		fmt.Sprintf("--runtime=%d", durationSecs),
		fmt.Sprintf("--filename=%s", common.FioFsFilename),
		fmt.Sprintf("--size=%dm", volSize),
	}
	fioArgs := append(args, common.GetFioArgs()...)
	// fio pod container
	podContainer := coreV1.Container{
		Name:  fioPodName,
		Image: common.GetFioImage(),
		Args:  fioArgs,
	}
	// volume claim details
	volume := coreV1.Volume{
		Name: "ms-volume",
		VolumeSource: coreV1.VolumeSource{
			PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
	}
	podObj, err := k8stest.NewPodBuilder().
		WithName(fioPodName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(common.VolFileSystem).Build()
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("Generating fio pod definition %s", fioPodName))
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")
	// Create fio pod
	_, err = k8stest.CreatePod(podObj, common.NSDefault)

	if err != nil {
		return nil
	}
	return err
}

func verifyChildrenCount(uuid string, replicas int) bool {
	children, err := k8stest.GetMsvNexusChildren(uuid)
	if err != nil {
		panic("Failed to get children")
	}
	return len(children) == replicas
}

func getChildren(uuid string) []common.NexusChild {
	children, err := k8stest.GetMsvNexusChildren(uuid)
	if err != nil {
		panic("Failed to get children")
	}
	return children
}
