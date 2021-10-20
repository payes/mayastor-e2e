package k8stest

// Utility functions for Persistent Volume Claims and Persistent Volumes
import (
	"context"
	"fmt"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/mayastorclient"
	"strings"
	"sync"

	"mayastor-e2e/common"

	. "github.com/onsi/gomega"

	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defTimeoutSecs = "90s"

// IsPVCDeleted Check for a deleted Persistent Volume Claim,
// either the object does not exist
// or the status phase is invalid.
func IsPVCDeleted(volName string, nameSpace string) bool {
	pvc, err := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	if err != nil {
		// Unfortunately there is no associated error code, so we resort to string comparison
		if strings.HasPrefix(err.Error(), "persistentvolumeclaims") &&
			strings.HasSuffix(err.Error(), " not found") {
			return true
		}
	}
	// After the PVC has been deleted it may still accessible, but status phase will be invalid
	Expect(err).To(BeNil())
	Expect(pvc).ToNot(BeNil())
	switch pvc.Status.Phase {
	case
		coreV1.ClaimBound,
		coreV1.ClaimPending,
		coreV1.ClaimLost:
		return false
	default:
		return true
	}
}

// IsPVDeleted Check for a deleted Persistent Volume,
// either the object does not exist
// or the status phase is invalid.
func IsPVDeleted(volName string) bool {
	pv, err := gTestEnv.KubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), volName, metaV1.GetOptions{})
	if err != nil {
		// Unfortunately there is no associated error code so we resort to string comparison
		if strings.HasPrefix(err.Error(), "persistentvolumes") &&
			strings.HasSuffix(err.Error(), " not found") {
			return true
		}
	}
	// After the PV has been deleted it may still accessible, but status phase will be invalid
	Expect(err).To(BeNil())
	Expect(pv).ToNot(BeNil())
	logf.Log.Info("IsPVDeleted", "volume", volName, "status.Phase", pv.Status.Phase)
	switch pv.Status.Phase {
	case
		coreV1.VolumeBound,
		coreV1.VolumeAvailable,
		coreV1.VolumeFailed,
		coreV1.VolumePending,
		coreV1.VolumeReleased:
		return false
	default:
		return true
	}
}

// IsPvcBound returns true if a PVC with the given name is bound otherwise false is returned.
func IsPvcBound(pvcName string, nameSpace string) bool {
	return GetPvcStatusPhase(pvcName, nameSpace) == coreV1.ClaimBound
}

// GetPvcStatusPhase Retrieve status phase of a Persistent Volume Claim
func GetPvcStatusPhase(volname string, nameSpace string) (phase coreV1.PersistentVolumeClaimPhase) {
	pvc, getPvcErr := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Get(context.TODO(), volname, metaV1.GetOptions{})
	Expect(getPvcErr).To(BeNil())
	Expect(pvc).ToNot(BeNil())
	return pvc.Status.Phase
}

// GetPvStatusPhase Retrieve status phase of a Persistent Volume
func GetPvStatusPhase(volname string) (phase coreV1.PersistentVolumePhase) {
	pv, getPvErr := gTestEnv.KubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), volname, metaV1.GetOptions{})
	Expect(getPvErr).To(BeNil())
	Expect(pv).ToNot(BeNil())
	return pv.Status.Phase
}

// MkPVC Create a PVC and verify that
//	1. The PVC status transitions to bound,
//	2. The associated PV is created and its status transitions bound
//	3. The associated MV is created and has a State "healthy"
func MkPVC(volSizeMb int, volName string, scName string, volType common.VolumeType, nameSpace string) string {
	logf.Log.Info("Creating", "volume", volName, "storageClass", scName, "volume type", volType)
	volSizeMbStr := fmt.Sprintf("%dMi", volSizeMb)
	// PVC create options
	createOpts := &coreV1.PersistentVolumeClaim{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      volName,
			Namespace: nameSpace,
		},
		Spec: coreV1.PersistentVolumeClaimSpec{
			StorageClassName: &scName,
			AccessModes:      []coreV1.PersistentVolumeAccessMode{coreV1.ReadWriteOnce},
			Resources: coreV1.ResourceRequirements{
				Requests: coreV1.ResourceList{
					coreV1.ResourceStorage: resource.MustParse(volSizeMbStr),
				},
			},
		},
	}

	switch volType {
	case common.VolFileSystem:
		var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem
		createOpts.Spec.VolumeMode = &fileSystemVolumeMode
	case common.VolRawBlock:
		var blockVolumeMode = coreV1.PersistentVolumeBlock
		createOpts.Spec.VolumeMode = &blockVolumeMode
	}

	// Create the PVC.
	PVCApi := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims
	_, createErr := PVCApi(nameSpace).Create(context.TODO(), createOpts, metaV1.CreateOptions{})
	Expect(createErr).To(BeNil())

	// Confirm the PVC has been created.
	pvc, getPvcErr := PVCApi(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	Expect(getPvcErr).To(BeNil())
	Expect(pvc).ToNot(BeNil())

	ScApi := gTestEnv.KubeInt.StorageV1().StorageClasses
	sc, getScErr := ScApi().Get(context.TODO(), scName, metaV1.GetOptions{})
	Expect(getScErr).To(BeNil())
	if *sc.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
		return string(pvc.ObjectMeta.UID)
	}

	// Wait for the PVC to be bound.
	Eventually(func() coreV1.PersistentVolumeClaimPhase {
		return GetPvcStatusPhase(volName, nameSpace)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(coreV1.ClaimBound))

	// Refresh the PVC contents, so that we can get the PV name.
	pvc, getPvcErr = PVCApi(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	Expect(getPvcErr).To(BeNil())
	Expect(pvc).ToNot(BeNil())

	// Wait for the PV to be provisioned
	Eventually(func() *coreV1.PersistentVolume {
		pv, getPvErr := gTestEnv.KubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), pvc.Spec.VolumeName, metaV1.GetOptions{})
		if getPvErr != nil {
			return nil
		}
		return pv

	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Not(BeNil()))

	// Wait for the PV to be bound.
	Eventually(func() coreV1.PersistentVolumePhase {
		return GetPvStatusPhase(pvc.Spec.VolumeName)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(coreV1.VolumeBound))

	Eventually(func() *common.MayastorVolume {
		msv, _ := GetMSV(string(pvc.ObjectMeta.UID))
		return msv
	},
		defTimeoutSecs,
		"1s",
	).Should(Not(BeNil()))

	err := MsvConsistencyCheck(string(pvc.ObjectMeta.UID))
	Expect(err).ToNot(HaveOccurred(), "%v", err)

	logf.Log.Info("Created", "volume", volName, "uuid", pvc.ObjectMeta.UID, "storageClass", scName, "volume type", volType)
	return string(pvc.ObjectMeta.UID)
}

// MsvConsistencyCheck check consistency of  MSV Spec, Status, and associated objects returned by gRPC
func MsvConsistencyCheck(uuid string) error {
	//FIXME: implement new MsvConsistencyCheck inline with mayastor control plane
	if controlplane.MajorVersion() != 0 {
		return nil
	}
	msv, err := GetMSV(uuid)
	if msv == nil {
		return fmt.Errorf("MsvConsistencyCheck: GetMsv: %v, got nil pointer to msv", uuid)
	}
	if err != nil {
		return fmt.Errorf("MsvConsistencyCheck: GetMsv: %v", err)
	}
	if int64(msv.Spec.RequiredBytes) != msv.Status.Size {
		return fmt.Errorf("MsvConsistencyCheck: msv spec required bytes %d != msv status size %d", msv.Spec.RequiredBytes, msv.Status.Size)
	}
	if msv.Spec.ReplicaCount != len(msv.Status.Replicas) {
		return fmt.Errorf("MsvConsistencyCheck: msv spec replica count %d != msv status replicas %d", msv.Spec.ReplicaCount, len(msv.Status.Replicas))
	}

	if mayastorclient.CanConnect() {

		gReplicas, err := mayastorclient.FindReplicas(uuid, GetMayastorNodeIPAddresses())
		if err != nil {
			return fmt.Errorf("failed to find replicas using gRPC %v", err)
		}
		for _, gReplica := range gReplicas {
			if gReplica.Size != uint64(msv.Status.Size) {
				return fmt.Errorf("MsvConsistencyCheck: replica size  %d != msv status size %d", gReplica.Size, msv.Status.Size)
			}
		}

		if msv.Spec.ReplicaCount != len(gReplicas) {
			return fmt.Errorf("MsvConsistencyCheck: msv spec replica count %d != list matching replicas found using gRPC %d", msv.Spec.ReplicaCount, len(gReplicas))
		}
		nexus := msv.Status.Nexus
		// The nexus is only present when a volume is mounted by a pod.
		if nexus.Node != "" {
			if msv.Spec.ReplicaCount != len(msv.Status.Nexus.Children) {
				return fmt.Errorf("MsvConsistencyCheck: msv spec replica count %d != msv status nexus children %d", msv.Spec.ReplicaCount, len(msv.Status.Nexus.Children))
			}
			nexusNodeIp, err := GetNodeIPAddress(nexus.Node)
			if err != nil {
				return fmt.Errorf("MsvConsistencyCheck: failed to resolve nexus node IP address, %v", err)
			}
			grpcNexus, err := mayastorclient.FindNexus(uuid, []string{*nexusNodeIp})
			if err != nil {
				return fmt.Errorf("MsvConsistencyCheck: failed to list nexuses gRPC, %v", err)
			}
			if grpcNexus == nil {
				return fmt.Errorf("MsvConsistencyCheck: failed to find nexus gRPC")
			}
			if grpcNexus.Size != uint64(msv.Status.Size) {
				return fmt.Errorf("MsvConsistencyCheck: nexus size mismatch msv and grpc")
			}
			if len(grpcNexus.Children) != msv.Spec.ReplicaCount {
				return fmt.Errorf("MsvConsistencyCheck: msv replica count != grpc nexus children")
			}
			if grpcNexus.State.String() != msv.Status.Nexus.State {
				return fmt.Errorf("MsvConsistencyCheck: msv nexus state != grpc nexus state")
			}
		} else {
			logf.Log.Info("MsvConsistencyCheck nexus unavailable")
		}
	} else {
		logf.Log.Info("MsvConsistencyCheck,  gRPC calls to mayastor are not enabled, not checking MSVs using gRPC calls")
	}

	logf.Log.Info("MsvConsistencyCheck OK")
	return nil
}

// RmPVC Delete a PVC in the default namespace and verify that
//	1. The PVC is deleted
//	2. The associated PV is deleted
//  3. The associated MV is deleted
func RmPVC(volName string, scName string, nameSpace string) {
	logf.Log.Info("Removing volume", "volume", volName, "storageClass", scName)

	PVCApi := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims

	// Confirm the PVC has been deleted.
	pvc, getPvcErr := PVCApi(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	if k8serrors.IsNotFound(getPvcErr) {
		return
	} else {
		Expect(getPvcErr).To(BeNil())
		Expect(pvc).ToNot(BeNil())
	}
	// Delete the PVC
	deleteErr := PVCApi(nameSpace).Delete(context.TODO(), volName, metaV1.DeleteOptions{})
	Expect(deleteErr).To(BeNil())

	logf.Log.Info("Waiting for PVC to be deleted", "volume", volName, "storageClass", scName)
	// Wait for the PVC to be deleted.
	Eventually(func() bool {
		return IsPVCDeleted(volName, nameSpace)
	},
		defTimeoutSecs, // timeout
		"1s",           // polling interval
	).Should(Equal(true), "pvc not deleted %s", volName)
	// Wait for the PV to be deleted.
	logf.Log.Info("Waiting for PV to be deleted", "volume", volName, "storageClass", scName)
	Eventually(func() bool {
		// This check is required here because it will check for pv name
		// when pvc is in pending state at that time we will not
		// get pv name inside pvc spec i.e pvc.Spec.VolumeName
		if pvc.Spec.VolumeName != "" {
			return IsPVDeleted(pvc.Spec.VolumeName)
		}
		return true
	},
		"360s", // timeout
		"1s",   // polling interval
	).Should(Equal(true), "PV (%s) for %s not deleted", pvc.Spec.VolumeName, volName)

	// Wait for the MSV to be deleted.
	Eventually(func() bool {
		return IsMsvDeleted(string(pvc.ObjectMeta.UID))
	},
		"360s", // timeout
		"1s",   // polling interval
	).Should(Equal(true))
}

// CreatePVC Create a PVC in default namespace, no options and no context
func CreatePVC(pvc *v1.PersistentVolumeClaim, nameSpace string) (*v1.PersistentVolumeClaim, error) {
	return gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Create(context.TODO(), pvc, metaV1.CreateOptions{})
}

// GetPVC Retrieve a PVC in default namespace, no options and no context
func GetPVC(volName string, nameSpace string) (*v1.PersistentVolumeClaim, error) {
	return gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
}

// DeletePVC Delete a PVC in default namespace, no options and no context
func DeletePVC(volName string, nameSpace string) error {
	return gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Delete(context.TODO(), volName, metaV1.DeleteOptions{})
}

// GetPV Retrieve a PV in default namespace, no options and no context
func GetPV(volName string) (*v1.PersistentVolume, error) {
	return gTestEnv.KubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), volName, metaV1.GetOptions{})
}

func getMayastorScMap() (map[string]bool, error) {
	mayastorStorageClasses := make(map[string]bool)
	ScApi := gTestEnv.KubeInt.StorageV1().StorageClasses
	scs, err := ScApi().List(context.TODO(), metaV1.ListOptions{})
	if err == nil {
		for _, sc := range scs.Items {
			if sc.Provisioner == common.CSIProvisioner {
				mayastorStorageClasses[sc.Name] = true
			}
		}
	}
	return mayastorStorageClasses, err
}

func CheckForPVCs() (bool, error) {
	logf.Log.Info("CheckForPVCs")
	foundResources := false

	mayastorStorageClasses, err := getMayastorScMap()
	if err != nil {
		return false, err
	}

	nameSpaces, err := gTestEnv.KubeInt.CoreV1().Namespaces().List(context.TODO(), metaV1.ListOptions{})
	if err == nil {
		for _, ns := range nameSpaces.Items {
			if strings.HasPrefix(ns.Name, common.NSE2EPrefix) || ns.Name == common.NSDefault {
				pvcs, err := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(ns.Name).List(context.TODO(), metaV1.ListOptions{})
				if err == nil && pvcs != nil && len(pvcs.Items) != 0 {
					for _, pvc := range pvcs.Items {
						if !mayastorStorageClasses[*pvc.Spec.StorageClassName] {
							continue
						}
						logf.Log.Info("CheckForVolumeResources: found PersistentVolumeClaims",
							"PersistentVolumeClaim", pvc)
						foundResources = true
					}
				}
			}
		}
	}

	return foundResources, err
}

func CheckForPVs() (bool, error) {
	logf.Log.Info("CheckForPVs")
	foundResources := false

	mayastorStorageClasses, err := getMayastorScMap()
	if err != nil {
		return false, err
	}

	pvs, err := gTestEnv.KubeInt.CoreV1().PersistentVolumes().List(context.TODO(), metaV1.ListOptions{})
	if err == nil && pvs != nil && len(pvs.Items) != 0 {
		for _, pv := range pvs.Items {
			if !mayastorStorageClasses[pv.Spec.StorageClassName] {
				continue
			}
			logf.Log.Info("CheckForVolumeResources: found PersistentVolumes",
				"PersistentVolume", pv)
			foundResources = true
		}
	}
	return foundResources, err
}

func CreatePvc(createOpts *coreV1.PersistentVolumeClaim, errBuf *error, uuid *string, wg *sync.WaitGroup) {
	// Create the PVC.
	pvc, err := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(createOpts.ObjectMeta.Namespace).Create(context.TODO(), createOpts, metaV1.CreateOptions{})
	*errBuf = err
	if pvc != nil {
		*uuid = string(pvc.UID)
	}
	wg.Done()
}

func DeletePvc(volName string, namespace string, errBuf *error, wg *sync.WaitGroup) {
	// Delete the PVC.
	err := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(namespace).Delete(context.TODO(), volName, metaV1.DeleteOptions{})
	*errBuf = err
	wg.Done()
}
