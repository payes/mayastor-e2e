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

profiles[nightly-stable]="
basic_volume_io
check_mayastornode
control_plane_rescheduling
csi
resource_check
dynamic_provisioning
expand_msp_disk
mayastorpool_schema
ms_pod_restart
ms_pool_delete
pool_modify
primitive_data_integrity
primitive_msp_deletion
primitive_replicas
primitive_volumes
pvc_readwriteonce
"

profiles[c1]="
nexus_location
pvc_stress_fio
pvc_waitforfirstconsumer
volume_filesystem
synchronous_replication
io_soak
pvc_delete
ms_pod_disruption
ms_pod_disruption_no_io
ms_pod_disruption_rm_msv
maximum_vols_io
multiple_vols_pod_io
node_failure
node_shutdown
single_msn_shutdown
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
pvc_delete
control_plane_rescheduling
"

profiles[validation]="
validate_integrity_test
"

profiles[staging]="
primitive_msp_state
primitive_msp_stress
msv_rebuild
"
