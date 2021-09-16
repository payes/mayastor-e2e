package main

import (
	"fmt"
	"mayastor-e2e/common"
	"mayastor-e2e/lib"
	"time"

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
	fmt.Printf("created sc\n")

	// create PV
	pvcname, err := lib.MkPVC(testConductor.clientset, 64, pvc_name, sc_name, vol_type, common.NSDefault)
	if err != nil {
		return fmt.Errorf("failed to create pvc %v", err)
	}
	fmt.Printf("created pvc %s\n", pvcname)

	// deploy fio
	err = lib.DeployFio(testConductor.clientset, fio_name, pvc_name, vol_type, 64, 1)
	if err != nil {
		return fmt.Errorf("failed to deploy pod %s, error: %v", fio_name, err)
	}
	fmt.Printf("created pod %s\n", fio_name)

	// alert workload monitor
	time.Sleep(600 * time.Second)
	return err
}
