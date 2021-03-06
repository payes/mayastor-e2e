package k8sclient

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"time"

	"mayastor-e2e/common/mayastorclient"
	//"mayastor-e2e/tools/extended-test-framework/common/controlplane"

	"mayastor-e2e/tools/extended-test-framework/common/custom_resources"
	v1alpha1Api "mayastor-e2e/tools/extended-test-framework/common/custom_resources/api/types/v1alpha1"
	"mayastor-e2e/tools/extended-test-framework/common/mini_mcp_client"

	coreV1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const defTimeoutSecs = 180

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*v1alpha1Api.MayastorVolume, error) {
	msv, err := custom_resources.GetMsVol(uuid)
	if err != nil {
		return nil, fmt.Errorf("GetMSV: %v", err)
	}

	// pending means still being created
	if msv.Status.State == "pending" {
		return nil, nil
	}

	// Note: msVol.Node can be unassigned here if the volume is not mounted
	if msv.Status.State == "" {
		return nil, fmt.Errorf("GetMSV: state not defined, got msv.Status=\"%v\"", msv.Status)
	}

	if len(msv.Status.Replicas) < 1 {
		return nil, fmt.Errorf("GetMSV: msv.Status.Replicas=\"%v\"", msv.Status.Replicas)
	}
	return msv, nil
}

// GetPvcStatusPhase Retrieve status phase of a Persistent Volume Claim
func GetPvcStatusPhase(volname string, nameSpace string) (coreV1.PersistentVolumeClaimPhase, error) {
	pvc, getPvcErr := gKubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Get(context.TODO(), volname, metaV1.GetOptions{})
	return pvc.Status.Phase, getPvcErr
}

// GetPvStatusPhase Retrieve status phase of a Persistent Volume
func GetPvStatusPhase(volname string) (coreV1.PersistentVolumePhase, error) {
	pv, getPvErr := gKubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), volname, metaV1.GetOptions{})
	return pv.Status.Phase, getPvErr
}

// MkPVC Create a PVC and verify that
//	1. The PVC status transitions to bound,
//	2. The associated PV is created and its status transitions bound
//	3. The associated MV is created and has a State "healthy"
func MkPVC(volSizeMb int, volName string, scName string, volType VolumeType, nameSpace string, moac bool) (string, error) {
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
	case VolFileSystem:
		var fileSystemVolumeMode = coreV1.PersistentVolumeFilesystem
		createOpts.Spec.VolumeMode = &fileSystemVolumeMode
	case VolRawBlock:
		var blockVolumeMode = coreV1.PersistentVolumeBlock
		createOpts.Spec.VolumeMode = &blockVolumeMode
	}

	// Create the PVC.
	PVCApi := gKubeInt.CoreV1().PersistentVolumeClaims
	_, createErr := PVCApi(nameSpace).Create(context.TODO(), createOpts, metaV1.CreateOptions{})
	if createErr != nil {
		return "", createErr
	}

	// Confirm the PVC has been created.
	pvc, getPvcErr := PVCApi(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	if getPvcErr != nil {
		return "", getPvcErr
	}
	if pvc == nil {
		return "", fmt.Errorf("PVC is nil")
	}

	ScApi := gKubeInt.StorageV1().StorageClasses
	sc, getScErr := ScApi().Get(context.TODO(), scName, metaV1.GetOptions{})
	if getScErr != nil {
		return "", getScErr
	}
	// no need to wait for it to be bound
	if *sc.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
		return string(pvc.ObjectMeta.UID), nil
	}

	// Wait for the PVC to be bound.
	for i := 0; ; i++ {
		if i >= defTimeoutSecs {
			return "", fmt.Errorf("timed out waiting for PVC to be bound")
		}
		phase, err := GetPvcStatusPhase(volName, nameSpace)
		if err != nil {
			return "", err
		}
		if phase == coreV1.ClaimBound {
			break
		}
		time.Sleep(time.Second)
	}

	// Refresh the PVC contents, so that we can get the PV name.
	pvc, getPvcErr = PVCApi(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
	if getPvcErr != nil {
		return "", getPvcErr
	}
	if pvc == nil {
		return "", fmt.Errorf("PVC is nil")
	}

	// Wait for the PV to be provisioned
	for i := 0; ; i++ {
		pv, getPvErr := gKubeInt.CoreV1().PersistentVolumes().Get(context.TODO(), pvc.Spec.VolumeName, metaV1.GetOptions{})
		if getPvErr == nil && pv != nil {
			break
		}
		if i >= defTimeoutSecs {
			return "", fmt.Errorf("timed out waiting for PV")
		}
		time.Sleep(time.Second)
	}

	// Wait for the PV to be bound.
	for i := 0; ; i++ {
		if i >= defTimeoutSecs {
			return "", fmt.Errorf("timed out waiting for PV to be bound")
		}
		phase, err := GetPvStatusPhase(pvc.Spec.VolumeName)
		if err != nil {
			return "", err
		}
		if phase == coreV1.VolumeBound {
			break
		}
		time.Sleep(time.Second)
	}
	if moac {
		for i := 0; ; i++ {
			if i >= defTimeoutSecs {
				return "", fmt.Errorf("timed out waiting for PVC to be bound")
			}
			msv, _ := GetMSV(string(pvc.ObjectMeta.UID))
			if msv != nil {
				break
			}
			time.Sleep(time.Second)
		}
	}
	logf.Log.Info("Created", "volume", volName, "uuid", pvc.ObjectMeta.UID, "storageClass", scName, "volume type", volType)
	return string(pvc.ObjectMeta.UID), nil
}

func DeletePVC(volName string, nameSpace string) error {
	logf.Log.Info("Deleting", "PVC", volName)
	err := gKubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Delete(context.TODO(), volName, metaV1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete PVC %s", volName)
	}

	// Wait for the PVC to be removed
	for i := 0; ; i++ {
		_, err := gKubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Get(context.TODO(), volName, metaV1.GetOptions{})
		if err != nil && errors.IsNotFound(err) {
			break
		}
		if i >= defTimeoutSecs {
			return fmt.Errorf("timed out waiting for PVC %s to be deleted", volName)
		}
		time.Sleep(time.Second)
	}
	return nil
}

func ListPVCs(nameSpace string) (*coreV1.PersistentVolumeClaimList, error) {
	pvcs, err := gKubeInt.CoreV1().PersistentVolumeClaims(nameSpace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list PVCs, error %v", err)
	}
	return pvcs, err
}

// gather the set of MSV states using grpc calls to the mayastor instances
func GetNexuses(nodeIPs []string) ([]mayastorclient.MayastorNexus, error) {
	var nexuses []mayastorclient.MayastorNexus

	nexuses, err := mayastorclient.ListNexuses(nodeIPs)
	if err != nil {
		return nexuses, fmt.Errorf("failed to list nexuses via gRPC, %v", err)
	}
	return nexuses, err
}

func WaitForMCPPath(timeout string) error {
	var err error
	timeoutSec, err := time.ParseDuration(timeout)
	const sleepTimeSec = 10 // sleep time in second
	if err != nil {
		return fmt.Errorf("failed to parse timeout %s string , error: %v", timeout, err)
	}
	ms_ips, err := GetMayastorNodeIPs()
	if err != nil {
		return fmt.Errorf("failed to get mayastor nodes IP, err: %s", err)
	} else if len(ms_ips) == 0 {
		return fmt.Errorf("No mayastor nodes found")
	}
	for ix := 0; ix < int(timeoutSec.Seconds())/sleepTimeSec; ix++ {
		// If this call goes through implies
		// REST, Core Agent and etcd pods are up and running
		_, err = mini_mcp_client.GetVolumes(ms_ips[0])
		if err != nil {
			return fmt.Errorf("failed to get volumes, error: %v", err)
		}
		if err != nil {
			logf.Log.Info("Failed to list msv", "error", err)
		} else {
			break
		}
		time.Sleep(sleepTimeSec * time.Second)
	}
	if err != nil {
		return fmt.Errorf("one of the rest, core agent or etcd pods are not in running state, error: %v", err)
	}
	return nil
}
