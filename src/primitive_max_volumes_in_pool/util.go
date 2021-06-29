package primitive_max_volumes_in_pool

import (
	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/k8stest"

	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// createSC will create storageclass
func (c *primitiveMaxVolConfig) createSC() {
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
func (c *primitiveMaxVolConfig) deleteSC() {
	err := k8stest.RmStorageClass(c.scName)
	Expect(err).ToNot(HaveOccurred(), "Deleting storage class %s", c.scName)
}

// createPVC will create pvc
func (c *primitiveMaxVolConfig) createPVC() *primitiveMaxVolConfig {
	// Create the volumes
	for _, pvc := range c.pvcNames {
		uid := k8stest.MkPVC(c.pvcSize, pvc, c.scName, common.VolFileSystem, common.NSDefault)
		c.uuid = append(c.uuid, uid)
	}

	return c
}

// verify msp used size
func (c *primitiveMaxVolConfig) verifyMspUsedSize(size int64) {
	// List Pools by CRDs
	crdPools, err := custom_resources.ListMsPools()
	Expect(err).ToNot(HaveOccurred(), "List pools via CRD failed")
	for _, crdPool := range crdPools {
		logf.Log.Info("Pool If", "name", crdPool.Name, "Used", crdPool.Status.Used, "Expected", size)
		if crdPool.Status.Used != int64(size) {
			logf.Log.Info("Pool If", "name", crdPool.Name, "Used", crdPool.Status.Used)
			errors.Errorf("pool %s used size did not reconcile", crdPool.Name)
			return
		}

	}
}

// deletePVC will delete all pvc
func (c *primitiveMaxVolConfig) deletePVC() {
	for _, pvc := range c.pvcNames {
		k8stest.RmPVC(pvc, c.scName, common.NSDefault)
	}
}
