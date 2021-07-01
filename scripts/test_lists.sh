#!/usr/bin/env bash

profiles[default]="
basic_volume_io
csi
resource_check
replica
rebuild
ms_pod_disruption
"

profiles[continuous]="
basic_volume_io
csi
resource_check
rebuild
io_soak
volume_filesystem
ms_pod_disruption
ms_pod_disruption_no_io
ms_pod_disruption_rm_msv
"

profiles[nightly]="
csi
resource_check
primitive_msp_state
primitive_replicas
"

profiles[c1]="
basic_volume_io
check_mayastornode
control_plane_rescheduling
expand_msp_disk
mayastorpool_schema
ms_pod_restart
ms_pool_delete
nexus_location
pvc_readwriteonce
pvc_stress_fio
pvc_waitforfirstconsumer
volume_filesystem
synchronous_replication
io_soak
ms_pod_disruption_rm_msv
pool_modify
pvc_delete
dynamic_provisioning
msv_rebuild
ms_pod_disruption_no_io
ms_pod_disruption
maximum_vols_io
node_failure
node_shutdown
single_msn_shutdown
"

profiles[notrun]="
multiple_vols_pod_io
rebuild
replica
"

profiles[nightly_full]="
basic_volume_io
csi
resource_check
dynamic_provisioning
pvc_stress_fio
replica
rebuild
io_soak
multiple_vols_pod_io
mayastorpool_schema
ms_pod_restart
nexus_location
pool_modify
pvc_readwriteonce
volume_filesystem
ms_pod_disruption
ms_pod_disruption_no_io
ms_pod_disruption_rm_msv
ms_pool_delete
check_mayastornode
control_plane_rescheduling
expand_msp_disk
pvc_waitforfirstconsumer
synchronous_replication
msv_rebuild
pvc_delete
maximum_vols_io
single_msn_shutdown
node_shutdown
node_failure
primitive_replicas
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

profiles[soak]="
io_soak
"

profiles[validation]="
validate_integrity_test
"
