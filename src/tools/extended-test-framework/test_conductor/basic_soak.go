package main

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/lib"
	"time"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	storageV1 "k8s.io/api/storage/v1"
)

func (testConductor TestConductor) BasicSoakTest() error {
	if err := InstallMayastor(testConductor.clientset, testConductor.config.PoolDevice); err != nil {
		return fmt.Errorf("failed to install mayastor %v", err)
	}
	var protocol common.ShareProto = common.ShareProtoNvmf
	var mode storageV1.VolumeBindingMode = storageV1.VolumeBindingImmediate
	var sc_name = "basic-soak-sc"
	var pvc_name = "basic-soak-pvc"
	var fio_name = "basic-soak-fio"
	var vol_type = common.VolRawBlock

	// create storage class
	err := lib.NewScBuilder().
		WithName(sc_name).
		WithReplicas(testConductor.config.DefaultReplicaCount).
		WithProtocol(protocol).
		WithNamespace(common.NSDefault).
		WithVolumeBindingMode(mode).
		BuildAndCreate(testConductor.clientset)
	if err != nil {
		return fmt.Errorf("failed to create sc %v", err)
	}
	logf.Log.Info("Created storage class", "sc", sc_name)

	// create PV
	pvcname, err := lib.MkPVC(testConductor.clientset, 64, pvc_name, sc_name, vol_type, common.NSDefault)
	if err != nil {
		return fmt.Errorf("failed to create pvc %v", err)
	}
	logf.Log.Info("Created pvc", "pvc", pvcname)

	// deploy fio
	err = lib.DeployFio(testConductor.clientset, fio_name, pvc_name, vol_type, 64, 1)
	if err != nil {
		return fmt.Errorf("failed to deploy pod %s, error: %v", fio_name, err)
	}
	logf.Log.Info("Created pod", "pod", fio_name)

	err = sendWorkload(testConductor.clientset, testConductor.pWorkloadMonitorClient, fio_name, common.NSDefault)
	if err != nil {
		return fmt.Errorf("failed to inform workload monitor of %s, error: %v", fio_name, err)
	}
	// alert workload monitor
	time.Sleep(600 * time.Second)
	return err
}
