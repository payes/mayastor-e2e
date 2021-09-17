#!/usr/bin/env bash

profiles[default]="
basic_volume_io
csi
resource_check
ms_pod_disruption
"

# deprecated use nightly-stable instead
profiles[nightly]="
primitive_replicas
primitive_msp_deletion
"

# NOTE resource_check must follow csi
profiles[nightly-stable]="
basic_volume_io
check_mayastornode
control_plane_rescheduling
csi
resource_check
dynamic_provisioning
expand_msp_disk
mayastorpool_schema
ms_pod_disruption_rm_msv
ms_pod_restart
ms_pool_delete
msv_rebuild
multiple_vols_pod_io
nexus_location
pool_modify
primitive_data_integrity
primitive_fault_injection
primitive_msp_stress
primitive_replicas
primitive_volumes
pvc_readwriteonce
pvc_stress_fio
pvc_waitforfirstconsumer
single_msn_shutdown
synchronous_replication
volume_filesystem
"

profiles[c1]="
io_soak
pvc_delete
ms_pod_disruption
ms_pod_disruption_no_io
node_shutdown
primitive_msp_deletion
"

profiles[notrun]="
basic_volume_io_iscsi
"

profiles[ondemand]="
basic_volume_io
csi
resource_check
"

# removed synchronous_replication doesn't  with mayastor build 755c435fdb0a.
# add pvc_delete and control_plane_rescheduling passes with mayastor build 755c435fdb0a
# remove pvc_delete, fails because MOAC does not remove finalizer
profiles[self_ci]="
basic_volume_io
io_soak
maximum_vols_io
multiple_vols_pod_io
csi
resource_check
pvc_stress_fio
ms_pod_restart
check_mayastornode
ms_pool_delete
volume_filesystem
dynamic_provisioning
mayastorpool_schema
expand_msp_disk
pvc_waitforfirstconsumer
nexus_location
pvc_readwriteonce
control_plane_rescheduling
"

profiles[validation]="
validate_integrity_test
"

profiles[staging]="
primitive_msp_state
primitive_fuzz_msv
MQ-1498-primitive_device_retirement
maximum_vols_io
node_failure
stale_msp_after_node_power_failure
MQ-1783-fsx_ext4_stress
"
