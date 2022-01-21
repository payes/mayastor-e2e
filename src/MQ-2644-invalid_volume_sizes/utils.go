package MQ_2644_invalid_volume_sizes

import (
	"fmt"
	. "github.com/onsi/gomega"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"mayastor-e2e/common"
	"mayastor-e2e/common/k8stest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"time"
)

var fioWriteParams = []string{
	"--name=benchtest",
	"--numjobs=1",
	"--direct=1",
	"--rw=randwrite",
	"--ioengine=libaio",
	"--bs=4k",
	"--iodepth=16",
	"--verify=crc32",
	"--size=10M",
}

func (c *pvcConfig) createStorageClass() {
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
func deleteSC(scName string) {
	err := k8stest.RmStorageClass(scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", scName)
}

// createPVC will create pvc in serial
func (c *pvcConfig) createPVC(pvcName string, errExpected bool) {
	logf.Log.Info("Creating", "volume", pvcName, "storageClass", c.scName, "volume type", common.VolFileSystem)
	volSizeMbStr := fmt.Sprintf("%dMi", c.pvcSizeMB)
	var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem
	// PVC create options
	createOpts := &coreV1.PersistentVolumeClaim{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      pvcName,
			Namespace: common.NSDefault,
		},
		Spec: coreV1.PersistentVolumeClaimSpec{
			StorageClassName: &c.scName,
			AccessModes:      []coreV1.PersistentVolumeAccessMode{coreV1.ReadWriteOnce},
			Resources: coreV1.ResourceRequirements{
				Requests: coreV1.ResourceList{
					coreV1.ResourceStorage: resource.MustParse(volSizeMbStr),
				},
			},
			VolumeMode: &fileSystemVolumeMode,
		},
	}
	// Create the volumes
	pvc, err := k8stest.CreatePVC(createOpts, common.NSDefault)
	if errExpected {
		Expect(err).ToNot(BeNil(), "Can not create PVC with zero or negative value")
	} else {
		Expect(err).To(BeNil(), "Failed to create pvc, error %v", err)
		Expect(pvc).ToNot(BeNil())
	}
}

func (c *pvcConfig) pvcInvalidSizeTest() {
	c.createStorageClass()

	c.createPVC(c.pvcName, false)

	logf.Log.Info("Checking for PVC status", "Duration", c.delayTime, "Expected status", "Pending")
	Consistently(func() bool {
		pvc, err := k8stest.GetPVC(c.pvcName, common.NSDefault)
		return err != nil || pvc.Status.Phase != "Pending"
	}, c.delayTime, time.Second*30).Should(BeFalse())
	Expect(checkFor507Event()).To(BeTrue())
}

func (c *pvcConfig) pvcNormalFioTest() {
	c.createStorageClass()

	c.createPVC(c.pvcName, false)
	logf.Log.Info("Creating PVC", "name", c.pvcName, "volume size", c.pvcSizeMB)
	//Wait for more than 5 min mins. This step makes sure that all the volume creation requests have been sent to csi controller pod
	time.Sleep(time.Duration(c.delayTime) * time.Minute)
}

func (c *pvcConfig) pvcZeroOrNegativeSizeTest() {
	c.createStorageClass()
	c.createPVC(c.pvcName, true)
}

func checkFor507Event() bool {
	options := metaV1.ListOptions{
		TypeMeta: metaV1.TypeMeta{Kind: "PersistentVolumeClaim"},
	}
	e, err := k8stest.GetEvents(common.NSDefault, options)
	if err != nil {
		logf.Log.Error(err, "Failed to list PVC events")
		return false
	}
	for _, ev := range e.Items {
		if ev.Type == "Warning" && strings.Contains(ev.Message, "507 Insufficient Storage") {
			return true
		}
	}
	return false
}

func (c *pvcConfig) createFioPod(podName string) {

	var args = []string{
		"--",
	}
	args = append(args, fmt.Sprintf("--filename=%s", common.FioFsFilename))

	args = append(args, fioWriteParams...)

	logf.Log.Info("fio", "arguments", args)

	// fio pod container
	podContainer := coreV1.Container{
		Name:            podName,
		Image:           common.GetFioImage(),
		ImagePullPolicy: coreV1.PullAlways,
		Args:            args,
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
		WithName(podName).
		WithNamespace(common.NSDefault).
		WithRestartPolicy(coreV1.RestartPolicyNever).
		WithContainer(podContainer).
		WithVolume(volume).
		WithVolumeDeviceOrMount(c.volType).Build()
	Expect(err).ToNot(HaveOccurred(), "Generating fio pod definition %s", podName)
	Expect(podObj).ToNot(BeNil(), "failed to generate fio pod definition")

	_, err = k8stest.CreatePod(podObj, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "Creating fio pod %s", podName)

	// Wait for the fio Pod to transition to running
	Eventually(func() bool {
		return k8stest.IsPodRunning(podName, common.NSDefault)
	},
		defTimeoutSecs,
		"1s",
	).Should(Equal(true))
	logf.Log.Info("fio test pod is running.")
}

func (c *pvcConfig) runFio() {
	name := c.pvcName + "-fio"
	c.createFioPod(name)
	Eventually(func() coreV1.PodPhase {
		phase, err := k8stest.CheckPodCompleted(name, common.NSDefault)
		logf.Log.Info("CheckPodCompleted phase", "actual", phase, "desired", coreV1.PodSucceeded)
		if err != nil {
			return coreV1.PodUnknown
		}
		return phase
	},
		defTimeoutSecs,
		"5s",
	).Should(Equal(coreV1.PodSucceeded))
}

func cleanUp() {
	for _, name := range testNames {
		k8stest.RmPVC(name+"-pvc", name+"-sc", common.NSDefault)
		deleteSC(name + "-sc")
	}
}
