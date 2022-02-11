package k8stest

// Utility functions for Persistent Volume Claims and Persistent Volumes
import (
	"context"
	"fmt"
	"mayastor-e2e/common/controlplane"
	"mayastor-e2e/common/mayastorclient"
	"strings"
	"sync"
	"time"

	"mayastor-e2e/common"

	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const defTimeoutSecs = 360

// IsPVCDeleted Check for a deleted Persistent Volume Claim,
// either the object does not exist
// or the status phase is invalid.
func IsPVCDeleted(volName string, nameSpace string) (bool, error) {
	pvc, err := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	if err != nil {
		// Unfortunately there is no associated error code, so we resort to string comparison
		if strings.HasPrefix(err.Error(), "persistentvolumeclaims") &&
			strings.HasSuffix(err.Error(), " not found") {
			return true, nil
		} else {
			return false, fmt.Errorf("failed to get pvc %s, namespace: %s, error: %v", volName, nameSpace, err)
		}
	}
	// After the PVC has been deleted it may still accessible, but status phase will be invalid
	if pvc != nil {
		switch pvc.Status.Phase {
		case
			coreV1.ClaimBound,
			coreV1.ClaimPending,
			coreV1.ClaimLost:
			return false, nil
		default:
			return true, nil
		}
	}
	return false, nil
}

// IsPVDeleted Check for a deleted Persistent Volume,
// either the object does not exist
// or the status phase is invalid.
func IsPVDeleted(volName string) (bool, error) {
	pv, err := gTestEnv.KubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), volName, metaV1.GetOptions{})
	if err != nil {
		// Unfortunately there is no associated error code so we resort to string comparison
		if strings.HasPrefix(err.Error(), "persistentvolumes") &&
			strings.HasSuffix(err.Error(), " not found") {
			return true, nil
		} else {
			return false, fmt.Errorf("failed to get pv %s, error: %v", volName, err)
		}
	}

	if pv != nil {
		switch pv.Status.Phase {
		case
			coreV1.VolumeBound,
			coreV1.VolumeAvailable,
			coreV1.VolumeFailed,
			coreV1.VolumePending,
			coreV1.VolumeReleased:
			return false, nil
		default:
			return true, nil
		}
	}
	// After the PV has been deleted it may still accessible, but status phase will be invalid
	logf.Log.Info("IsPVDeleted", "volume", volName, "status.Phase", pv.Status.Phase)
	return false, nil
}

// IsPvcBound returns true if a PVC with the given name is bound otherwise false is returned.
func IsPvcBound(pvcName string, nameSpace string) (bool, error) {
	phase, err := GetPvcStatusPhase(pvcName, nameSpace)
	if err != nil {
		return false, err
	}
	return phase == coreV1.ClaimBound, nil
}

// GetPvcStatusPhase Retrieve status phase of a Persistent Volume Claim
func GetPvcStatusPhase(volname string, nameSpace string) (phase coreV1.PersistentVolumeClaimPhase, err error) {
	pvc, getPvcErr := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Get(context.TODO(), volname, metaV1.GetOptions{})
	if getPvcErr != nil {
		return "", fmt.Errorf("failed to get pvc: %s, namespace: %s, error: %v",
			volname,
			nameSpace,
			getPvcErr,
		)
	}
	if pvc == nil {
		return "", fmt.Errorf("PVC %s not found, namespace: %s",
			volname,
			nameSpace,
		)
	}
	return pvc.Status.Phase, nil
}

// GetPvStatusPhase Retrieve status phase of a Persistent Volume
func GetPvStatusPhase(volname string) (phase coreV1.PersistentVolumePhase, err error) {
	pv, getPvErr := gTestEnv.KubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), volname, metaV1.GetOptions{})
	if getPvErr != nil {
		return "", fmt.Errorf("failed to get pv: %s, error: %v",
			volname,
			getPvErr,
		)
	}
	if pv == nil {
		return "", fmt.Errorf("PV not found: %s", volname)
	}
	return pv.Status.Phase, nil
}

// MkPVC Create a PVC and verify that
//	1. The PVC status transitions to bound,
//	2. The associated PV is created and its status transitions bound
//	3. The associated MV is created and has a State "healthy"
func MkPVC(volSizeMb int, volName string, scName string, volType common.VolumeType, nameSpace string) (string, error) {
	const timoSleepSecs = 1
	logf.Log.Info("Creating", "volume", volName, "storageClass", scName, "volume type", volType)
	volSizeMbStr := fmt.Sprintf("%dMi", volSizeMb)
	var err error

	t0 := time.Now()
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
	if createErr != nil {
		return "", fmt.Errorf("failed to create pvc: %s, error: %v", volName, createErr)
	}

	// Confirm the PVC has been created.
	pvc, getPvcErr := PVCApi(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	if getPvcErr != nil {
		return "", fmt.Errorf("failed to get pvc: %s, namespace: %s, error: %v", volName, nameSpace, getPvcErr)
	} else if pvc == nil {
		return "", fmt.Errorf("PVC %s not found, namespace: %s", volName, nameSpace)
	}

	ScApi := gTestEnv.KubeInt.StorageV1().StorageClasses
	sc, getScErr := ScApi().Get(context.TODO(), scName, metaV1.GetOptions{})
	if getScErr != nil {
		return "", fmt.Errorf("failed to get storageclass: %s, error: %v", scName, getScErr)
	}
	if *sc.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
		return string(pvc.ObjectMeta.UID), nil
	}

	// Wait for the PVC to be bound.
	for ix := 0; ix < defTimeoutSecs/timoSleepSecs; ix++ {
		var pvcPhase coreV1.PersistentVolumeClaimPhase
		pvcPhase, err = GetPvcStatusPhase(volName, nameSpace)
		if err == nil && pvcPhase == coreV1.ClaimBound {
			break
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get pvc status, pvc: %s, namespace:  %s, error: %v", volName, nameSpace, err)
	}

	// Refresh the PVC contents, so that we can get the PV name.
	pvc, getPvcErr = PVCApi(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	if getPvcErr != nil {
		return "", fmt.Errorf("failed to get pvc: %s, namespace: %s, error: %v", volName, nameSpace, getPvcErr)
	} else if pvc == nil {
		return "", fmt.Errorf("PVC %s not found, namespace: %s", volName, nameSpace)
	}

	// Wait for the PV to be provisioned
	var pv *coreV1.PersistentVolume
	for ix := 0; ix < defTimeoutSecs/timoSleepSecs; ix++ {
		pv, err = gTestEnv.KubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), pvc.Spec.VolumeName, metaV1.GetOptions{})
		if err == nil && pv != nil {
			break
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get pv, pv: %s, error: %v", pvc.Spec.VolumeName, err)
	}

	// Wait for the PV to be bound.
	for ix := 0; ix < defTimeoutSecs/timoSleepSecs; ix++ {
		var pvPhase coreV1.PersistentVolumePhase
		pvPhase, err = GetPvStatusPhase(pv.Name)
		if err == nil && pvPhase == coreV1.VolumeBound {
			break
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get pv status, pv: %s, error: %v", pvc.Spec.VolumeName, err)
	}

	// Wait for the PV to be provisioned
	for ix := 0; ix < defTimeoutSecs/timoSleepSecs; ix++ {
		var msv *common.MayastorVolume
		msv, err = GetMSV(string(pvc.ObjectMeta.UID))
		if err == nil && msv != nil {
			break
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get mayastor volume, uuid: %s, error: %v", pvc.ObjectMeta.UID, err)
	}

	err = MsvConsistencyCheck(string(pvc.ObjectMeta.UID))
	if err != nil {
		return "", fmt.Errorf("msv consistency check failed, msv uuid: %s, error: %v", string(pvc.ObjectMeta.UID), err)
	}

	logf.Log.Info("Created", "volume", volName, "uuid", pvc.ObjectMeta.UID, "storageClass", scName, "volume type", volType, "elapsed time", time.Since(t0))
	return string(pvc.ObjectMeta.UID), nil
}

// MsvConsistencyCheck check consistency of  MSV Spec, Status, and associated objects returned by gRPC
func MsvConsistencyCheck(uuid string) error {
	//FIXME: implement new MsvConsistencyCheck inline with mayastor control plane
	// JIRA: https://mayadata.atlassian.net/browse/MQ-2741
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
func RmPVC(volName string, scName string, nameSpace string) error {
	const timoSleepSecs = 1
	logf.Log.Info("Removing volume", "volume", volName, "storageClass", scName)
	var isDeleted bool

	PVCApi := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims

	// Confirm the PVC has been deleted.
	pvc, getPvcErr := PVCApi(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	if k8serrors.IsNotFound(getPvcErr) {
		return fmt.Errorf("PVC %s not found error, namespace: %s", volName, nameSpace)
	} else if getPvcErr != nil {
		return fmt.Errorf("failed to get pvc %s, namespace: %s, error: %v", volName, nameSpace, getPvcErr)
	} else if pvc == nil {
		return fmt.Errorf("PVC %s not found, namespace: %s", volName, nameSpace)
	}
	// Delete the PVC
	deleteErr := PVCApi(nameSpace).Delete(context.TODO(), volName, metaV1.DeleteOptions{})
	if deleteErr != nil {
		return fmt.Errorf("failed to delete PVC %s, namespace: %s, error: %v", volName, nameSpace, deleteErr)
	}

	logf.Log.Info("Waiting for PVC to be deleted", "volume", volName, "storageClass", scName)
	var err error
	// Wait for the PVC to be deleted.
	for ix := 0; ix < defTimeoutSecs/timoSleepSecs; ix++ {
		isDeleted, err = IsPVCDeleted(volName, nameSpace)
		if isDeleted {
			break
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	if err != nil {
		return err
	} else if !isDeleted {
		return fmt.Errorf("pvc not deleted, pvc: %s, namespace: %s", volName, nameSpace)
	}

	// Wait for the PV to be deleted.
	logf.Log.Info("Waiting for PV to be deleted", "volume", volName, "storageClass", scName)
	for ix := 0; ix < defTimeoutSecs/timoSleepSecs; ix++ {
		// This check is required here because it will check for pv name
		// when pvc is in pending state at that time we will not
		// get pv name inside pvc spec i.e pvc.Spec.VolumeName
		if pvc.Spec.VolumeName != "" {
			isDeleted, err = IsPVDeleted(pvc.Spec.VolumeName)
			if isDeleted {
				break
			}
		} else {
			isDeleted = true
			break
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	if err != nil {
		return err
	} else if !isDeleted {
		return fmt.Errorf("PV not deleted, pv: %s", pvc.Spec.VolumeName)
	}

	// Wait for the MSV to be deleted.
	for ix := 0; ix < defTimeoutSecs/timoSleepSecs; ix++ {
		isDeleted = IsMsvDeleted(string(pvc.ObjectMeta.UID))
		if isDeleted {
			break
		}
		time.Sleep(timoSleepSecs * time.Second)
	}
	if !isDeleted {
		return fmt.Errorf("mayastor volume not deleted, msv: %s", pvc.ObjectMeta.UID)
	}
	return nil
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
