package iscsi_sc_validation

import (
	"fmt"
	"mayastor-e2e/common"
	"strings"

	"mayastor-e2e/common/k8stest"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	. "github.com/onsi/gomega"
)

// ScIscsiValidation create/delete sc,pvc,fio-pod
func (c *ScIscsiValidationConfig) ScIscsiValidation() {
	c.createSC()
	c.createPvc()
	c.verifyVolumeCreationErrors()
	c.deletePVC()
	c.deleteSc()
}

// createSC will create storageclass
func (c *ScIscsiValidationConfig) createSC() {
	err := k8stest.NewScBuilder().
		WithName(c.ScName).
		WithNamespace(common.NSDefault).
		WithProtocol(c.Protocol).
		WithReplicas(c.Replicas).
		WithFileSystemType(c.FsType).
		BuildAndCreate()
	Expect(err).ToNot(HaveOccurred(), "failed to create storage class %s", c.ScName)
}

// deleteSc will delete storageclass
func (c *ScIscsiValidationConfig) deleteSc() {
	err := k8stest.RmStorageClass(c.ScName)
	Expect(err).ToNot(HaveOccurred(), "failed to delete storage class %s", c.ScName)
}

// createPvc will create pvc
func (c *ScIscsiValidationConfig) createPvc() *ScIscsiValidationConfig {
	logf.Log.Info("Creating", "volume", c.PvcName, "storageClass", c.ScName, "volume type", c.FsType)
	volSizeMbStr := fmt.Sprintf("%dMi", c.PvcSize)
	var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem
	// PVC create options
	createOpts := &coreV1.PersistentVolumeClaim{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      c.PvcName,
			Namespace: common.NSDefault,
		},
		Spec: coreV1.PersistentVolumeClaimSpec{
			StorageClassName: &c.ScName,
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
	Expect(err).To(BeNil(), "Failed to create pvc %s, error %v", c.PvcName, err)
	Expect(pvc).ToNot(BeNil())
	c.Uuid = string(pvc.UID)
	c.PvName = pvc.Spec.VolumeName
	return c
}

// deletePVC will delete all pvc
func (c *ScIscsiValidationConfig) deletePVC() {
	err := k8stest.RmPVC(c.PvcName, c.ScName, common.NSDefault)
	Expect(err).ToNot(HaveOccurred(), "failed to delete pvc %s", c.PvcName)
}

// Check that volume have been erred out
func (c *ScIscsiValidationConfig) verifyVolumeCreationErrors() {
	namespace := common.NSDefault
	volName := c.PvcName
	Eventually(func() bool {
		listOptions := metaV1.ListOptions{
			FieldSelector: "involvedObject.name=" + volName,
		}
		events, err := k8stest.GetEvents(namespace, listOptions)
		Expect(err).ToNot(HaveOccurred(), "failed to get events %v", err)
		for _, event := range events.Items {
			if strings.Contains(event.Message, "InvalidArgument") &&
				strings.Contains(event.Message, "Invalid protocol") {
				return true
			}
		}
		logf.Log.Info("Waiting for error event", "pvc", volName, "events", events.Items)
		return false
	},
		defTimeoutSecs, // timeout
		"5s",           // polling interval
	).Should(Equal(true), "failed to verify volume creation errors for %s", volName)
}
