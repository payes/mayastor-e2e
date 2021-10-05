package k8stest

// Utility functions for cleaning up a cluster
import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"mayastor-e2e/common"
	"mayastor-e2e/common/custom_resources"
	"mayastor-e2e/common/mayastorclient"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var ZeroInt64 = int64(0)

func DeleteAllDeployments(nameSpace string) (int, error) {
	logf.Log.Info("DeleteAllDeployments")
	numDeployments := 0

	deployments, err := gTestEnv.KubeInt.AppsV1().Deployments(nameSpace).List(context.TODO(), metaV1.ListOptions{})
	if err == nil {
		numDeployments = len(deployments.Items)
		logf.Log.Info("DeleteAllDeployments: found", "deployments", numDeployments)
		for _, deployment := range deployments.Items {
			logf.Log.Info("DeleteAllDeployments: Deleting", "deployment", deployment.Name)
			delErr := gTestEnv.KubeInt.AppsV1().Deployments(nameSpace).Delete(context.TODO(), deployment.Name, metaV1.DeleteOptions{})
			if delErr != nil {
				logf.Log.Info("DeleteAllDeployments: failed to delete the pod", "deployment", deployment.Name, "error", delErr)
			}
		}
	}

	return numDeployments, err
}

/// Delete all pods in the default namespace
func DeleteAllPods(nameSpace string) (int, error) {
	logf.Log.Info("DeleteAllPods")
	numPods := 0

	pods, err := gTestEnv.KubeInt.CoreV1().Pods(nameSpace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		logf.Log.Info("DeleteAllPods: list pods failed.", "error", err)
	} else {
		numPods = len(pods.Items)
		logf.Log.Info("DeleteAllPods: found", "pods", numPods)
		for _, pod := range pods.Items {
			logf.Log.Info("DeleteAllPods: Deleting", "pod", pod.Name)
			delErr := gTestEnv.KubeInt.CoreV1().Pods(nameSpace).Delete(context.TODO(), pod.Name, metaV1.DeleteOptions{GracePeriodSeconds: &ZeroInt64})
			if delErr != nil {
				logf.Log.Info("DeleteAllPods: failed to delete the pod", "podName", pod.Name, "error", delErr)
			}
		}
	}
	return numPods, err
}

// Make best attempt to delete PersistentVolumeClaims
// returns ok -> operations succeeded, resources undeleted, delete resources failed
func DeleteAllPvcs(nameSpace string) (int, error) {
	logf.Log.Info("DeleteAllPvcs")
	mayastorStorageClasses, err := getMayastorScMap()
	if err != nil {
		return -1, err
	}
	// Delete all PVCs found
	pvcs, err := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		logf.Log.Info("DeleteAllPvcs: list PersistentVolumeClaims failed.", "error", err)
	} else if len(pvcs.Items) != 0 {
		for _, pvc := range pvcs.Items {
			if !mayastorStorageClasses[*pvc.Spec.StorageClassName] {
				continue
			}
			logf.Log.Info("DeleteAllPvcs: deleting", "PersistentVolumeClaim", pvc.Name)
			delErr := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).Delete(context.TODO(), pvc.Name, metaV1.DeleteOptions{GracePeriodSeconds: &ZeroInt64})
			if delErr != nil {
				logf.Log.Info("DeleteAllPvcs: failed to delete", "PersistentVolumeClaim", pvc.Name, "error", delErr)
			}
		}
	}

	// Wait 2 minutes for PVCS to be deleted
	numPvcs := 0
	for attempts := 0; attempts < 120; attempts++ {
		numPvcs = 0
		pvcs, err := gTestEnv.KubeInt.CoreV1().PersistentVolumeClaims(nameSpace).List(context.TODO(), metaV1.ListOptions{})
		if err == nil {
			for _, pvc := range pvcs.Items {
				if !mayastorStorageClasses[*pvc.Spec.StorageClassName] {
					continue
				}
				numPvcs += 1
			}
			if numPvcs == 0 {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	logf.Log.Info("DeleteAllPvcs:", "number of PersistentVolumeClaims", numPvcs, "error", err)
	return numPvcs, err
}

// Make best attempt to delete PersistentVolumes
func DeleteAllPvs() (int, error) {
	mayastorStorageClasses, err := getMayastorScMap()
	if err != nil {
		return -1, err
	}
	// Delete all PVs found
	// First remove all finalizers
	pvs, err := gTestEnv.KubeInt.CoreV1().PersistentVolumes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		logf.Log.Info("DeleteAllPvs: list PersistentVolumes failed.", "error", err)
	} else if len(pvs.Items) != 0 {
		empty := make([]string, 0)
		for _, pv := range pvs.Items {
			if !mayastorStorageClasses[pv.Spec.StorageClassName] {
				continue
			}
			finalizers := pv.GetFinalizers()
			if len(finalizers) != 0 {
				logf.Log.Info("DeleteAllPvs: deleting finalizer for",
					"PersistentVolume", pv.Name, "finalizers", finalizers)
				pv.SetFinalizers(empty)
				_, _ = gTestEnv.KubeInt.CoreV1().PersistentVolumes().Update(context.TODO(), &pv, metaV1.UpdateOptions{})
			}
		}
	}

	// then wait for up to 2 minute for resources to be cleared
	numPvs := 0
	for attempts := 0; attempts < 120; attempts++ {
		numPvs = 0
		pvs, err := gTestEnv.KubeInt.CoreV1().PersistentVolumes().List(context.TODO(), metaV1.ListOptions{})
		if err == nil {
			for _, pv := range pvs.Items {
				if !mayastorStorageClasses[pv.Spec.StorageClassName] {
					continue
				}
				numPvs += 1
			}
			if numPvs == 0 {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	// Then delete the PVs
	pvs, err = gTestEnv.KubeInt.CoreV1().PersistentVolumes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		logf.Log.Info("DeleteAllPvs: list PersistentVolumes failed.", "error", err)
	} else if len(pvs.Items) != 0 {
		for _, pv := range pvs.Items {
			if !mayastorStorageClasses[pv.Spec.StorageClassName] {
				continue
			}
			logf.Log.Info("DeleteAllPvs: deleting PersistentVolume",
				"PersistentVolume", pv.Name)
			if delErr := gTestEnv.KubeInt.CoreV1().PersistentVolumes().Delete(context.TODO(), pv.Name, metaV1.DeleteOptions{GracePeriodSeconds: &ZeroInt64}); delErr != nil {
				logf.Log.Info("DeleteAllPvs: failed to delete PersistentVolume",
					"PersistentVolume", pv.Name, "error", delErr)
			}
		}
	}
	// Wait 2 minutes for resources to be deleted
	for attempts := 0; attempts < 120; attempts++ {
		numPvs = 0
		pvs, err := gTestEnv.KubeInt.CoreV1().PersistentVolumes().List(context.TODO(), metaV1.ListOptions{})
		if err == nil {
			for _, pv := range pvs.Items {
				if !mayastorStorageClasses[pv.Spec.StorageClassName] {
					continue
				}
				numPvs += 1
			}
			if numPvs == 0 {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	logf.Log.Info("DeleteAllPvs:", "number of PersistentVolumes", numPvs, "error", err)
	return numPvs, err
}

// DeleteAllMsvs Make best attempt to delete MayastorVolumes
func DeleteAllMsvs() (int, error) {
	// For now on control plane 2 we cannot/should not delete MSVs
	if common.IsControlPlaneMcp() {
		return 0, nil
	}

	// If after deleting PVCs and PVs Mayastor volumes are leftover
	// try cleaning them up explicitly

	msvs, err := ListMsvs()
	if err != nil {
		// This function may be called by AfterSuite by uninstall test so listing MSVs may fail correctly
		logf.Log.Info("DeleteAllMsvs: list MSVs failed.", "Error", err)
		return 0, err
	}
	if err == nil && msvs != nil && len(msvs) != 0 {
		for _, msv := range msvs {
			logf.Log.Info("DeleteAllMsvs: deleting MayastorVolume", "MayastorVolume", msv.Name)
			if delErr := DeleteMsv(msv.Name); delErr != nil {
				logf.Log.Info("DeleteAllMsvs: failed deleting MayastorVolume", "MayastorVolume", msv.Name, "error", delErr)
			}
		}
	}

	numMsvs := 0
	// Wait 2 minutes for resources to be deleted
	for attempts := 0; attempts < 120; attempts++ {
		msvs, err := ListMsvs()
		if err == nil && msvs != nil {
			numMsvs = len(msvs)
			if numMsvs == 0 {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	logf.Log.Info("DeleteAllMsvs:", "number of MayastorVolumes", numMsvs)

	return numMsvs, err
}

// deletePoolFinalizer, delete finalizers on a pool -if any.
// handle resource conflict errors by reloading the CR and retrying removal of finalizers.
// also handle concurrent removal of the pool gracefully
func deletePoolFinalizer(poolName string) (bool, error) {
	const sleepTime = 5
	for ix := 1; ix < 30; ix += sleepTime {
		msp, err := custom_resources.GetMsPool(poolName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				// The pool was deleted whilst trying to delete the pool finalizer
				// which means the pool finalizer was deleted by MOAC.
				// Nothing to do.
				return false, nil
			}
			// Failed to retrieve the MSP
			return false, err
		}
		if len(msp.Finalizers) == 0 {
			// No finalizers to remove
			return false, nil
		}
		msp.SetFinalizers(make([]string, 0))
		_, err = custom_resources.UpdateMsPool(msp)
		if err == nil {
			// Successfully removed finalizers
			return true, nil
		}
		// If the error is resource conflict try again
		if k8serrors.IsConflict(err) {
			logf.Log.Info("On pool update finalizer, got resource conflict error, retrying...")
		} else {
			if k8serrors.IsNotFound(err) {
				// The pool was deleted whilst trying to delete the pool finalizer
				// which means the pool finalizer was deleted by MOAC.
				// Nothing to do.
				return false, nil
			}
			logf.Log.Info("On pool update finalizer", "error", err)
			return false, err
		}
		time.Sleep(sleepTime * time.Second)
	}
	return false, fmt.Errorf("failed to remove pool finalizer (conflict)")
}

func DeleteAllPoolFinalizers() (bool, error) {
	deletedFinalizer := false
	var deleteErr error

	pools, err := custom_resources.ListMsPools()
	if err != nil {
		logf.Log.Info("DeleteAllPoolFinalizers: list MSPs failed.", "Error", err)
		return false, err
	}

	for _, pool := range pools {
		deleted, err := deletePoolFinalizer(pool.GetName())
		if err != nil {
			deleteErr = MakeAccumulatedError(deleteErr, err)
		}
		deletedFinalizer = deletedFinalizer || deleted
	}

	return deletedFinalizer, deleteErr
}

func DeleteAllPools() bool {
	pools, err := custom_resources.ListMsPools()
	if err != nil {
		// This function may be called by AfterSuite by uninstall test so listing MSVs may fail correctly
		logf.Log.Info("DeleteAllPools: list MSPs failed.", "Error", err)
	}
	if err == nil && pools != nil && len(pools) != 0 {
		logf.Log.Info("DeleteAllPools: deleting MayastorPools")
		for _, pool := range pools {
			logf.Log.Info("DeleteAllPools: deleting", "pool", pool.GetName())
			if !common.IsControlPlaneMcp() {
				finalizers := pool.GetFinalizers()
				if finalizers != nil {
					logf.Log.Info("DeleteAllPools: found finalizers on pool ", "pool", pool.GetName(), "finalizers", finalizers)
				}
			}
			err = custom_resources.DeleteMsPool(pool.GetName())
			if err != nil {
				logf.Log.Info("DeleteAllPools: failed to delete pool", "pool", pool.GetName(), "error", err)
			}
		}
	}

	numPools := 0
	// Wait 2 minutes for resources to be deleted
	for attempts := 0; attempts < 120; attempts++ {
		pools, err := custom_resources.ListMsPools()
		if err == nil && pools != nil {
			numPools = len(pools)
		}
		if numPools == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	logf.Log.Info("DeleteAllPools: ", "Pool count", numPools)
	if numPools != 0 {
		logf.Log.Info("DeleteAllPools: ", "Pools", pools)
	}
	return numPools == 0
}

//  >=0 definitive number of mayastor pods
// < 0 indeterminate
func MayastorUndeletedPodCount() int {
	ns, err := gTestEnv.KubeInt.CoreV1().Namespaces().Get(context.TODO(), common.NSMayastor(), metaV1.GetOptions{})
	if err != nil {
		logf.Log.Info("MayastorUndeletedPodCount: get namespace", "error", err)
		//FIXME: if the error is namespace not found return 0
		return -1
	}
	if ns == nil {
		// No namespace => no mayastor pods
		return 0
	}
	pods, err := gTestEnv.KubeInt.CoreV1().Pods(common.NSMayastor()).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		logf.Log.Info("MayastorUndeletedPodCount: list pods failed.", "error", err)
		return -1
	}
	return len(pods.Items)
}

// Force deletion of all existing mayastor pods
// returns  the number of pods still present, and error
func ForceDeleteMayastorPods() (bool, int, error) {
	var err error
	podsDeleted := false

	logf.Log.Info("EnsureMayastorDeleted")
	pods, err := gTestEnv.KubeInt.CoreV1().Pods(common.NSMayastor()).List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		logf.Log.Info("EnsureMayastorDeleted: list pods failed.", "error", err)
		return false, 0, err
	} else if len(pods.Items) == 0 {
		return false, 0, nil
	}

	logf.Log.Info("EnsureMayastorDeleted: MayastorPods found.", "Count", len(pods.Items))
	for _, pod := range pods.Items {
		logf.Log.Info("EnsureMayastorDeleted: Force deleting", "pod", pod.Name)
		cmd := exec.Command("kubectl", "-n", common.NSMayastor(), "delete", "pod", pod.Name, "--grace-period", "0", "--force")
		_, err := cmd.CombinedOutput()
		if err != nil {
			logf.Log.Info("EnsureMayastorDeleted", "podName", pod.Name, "error", err)
		} else {
			podsDeleted = true
		}
	}

	podCount := 0
	// We have made the best effort to cleanup, give things time to settle.
	for attempts := 0; attempts < 60 && MayastorUndeletedPodCount() != 0; attempts++ {
		pods, err = gTestEnv.KubeInt.CoreV1().Pods(common.NSMayastor()).List(context.TODO(), metaV1.ListOptions{})
		if err == nil {
			podCount = len(pods.Items)
			if podCount == 0 {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	return podsDeleted, podCount, err
}

// "Big" sweep, attempts to remove artefacts left over in the cluster
// that would prevent future successful test runs.
// returns true if cleanup was successful i.e. all resources were deleted
// and no errors were encountered.
func CleanUp() bool {
	var errs []error
	deploymentCount := 0
	podCount := 0
	pvcCount := 0

	nameSpaces, err := gTestEnv.KubeInt.CoreV1().Namespaces().List(context.TODO(), metaV1.ListOptions{})
	if err == nil {
		for _, ns := range nameSpaces.Items {
			if strings.HasPrefix(ns.Name, common.NSE2EPrefix) || ns.Name == common.NSDefault {
				tmp, err := DeleteAllDeployments(ns.Name)
				if err != nil {
					errs = append(errs, err)
				}
				deploymentCount += tmp
				tmp, err = DeleteAllPods(ns.Name)
				if err != nil {
					errs = append(errs, err)
				}
				podCount += tmp
				tmp, err = DeleteAllPvcs(ns.Name)
				if err != nil {
					errs = append(errs, err)
				}
				pvcCount += tmp
			}
		}
	} else {
		errs = append(errs, err)
	}

	pvCount, err := DeleteAllPvs()
	if err != nil {
		errs = append(errs, err)
	}
	msvCount, err := DeleteAllMsvs()
	if err != nil {
		errs = append(errs, err)
	}
	// Pools should not have finalizers if there are no associated volume resources.
	poolFinalizerDeleted, delPoolFinalizeErr := DeleteAllPoolFinalizers()

	logf.Log.Info("Resource cleanup",
		"deploymentCount", deploymentCount,
		"podCount", podCount,
		"pvcCount", pvcCount,
		"pvCount", pvCount,
		"msvCount", msvCount,
		"err", errs,
		"poolFinalizerDeleted", poolFinalizerDeleted,
		"delPoolFinalizeErr", delPoolFinalizeErr,
	)

	scList, err := gTestEnv.KubeInt.StorageV1().StorageClasses().List(context.TODO(), metaV1.ListOptions{})
	if err == nil {
		for _, sc := range scList.Items {
			if sc.Provisioner == "io.openebs.csi-mayastor" {
				logf.Log.Info("Deleting", "storageClass", sc.Name)
				_ = gTestEnv.KubeInt.StorageV1().StorageClasses().Delete(context.TODO(), sc.Name, metaV1.DeleteOptions{GracePeriodSeconds: &ZeroInt64})
			}
		}
	} else {
		errs = append(errs, err)
	}

	for _, ns := range nameSpaces.Items {
		if strings.HasPrefix(ns.Name, common.NSE2EPrefix) {
			err = RmNamespace(ns.Name)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	err = EnsureNodeLabels()
	if err != nil {
		errs = append(errs, err)
	}

	// log all the errors
	for _, err := range errs {
		logf.Log.Info("", "error", err)
	}

	if mayastorclient.CanConnect() {
		_ = RmReplicasInCluster()
	}

	return len(errs) == 0
}
