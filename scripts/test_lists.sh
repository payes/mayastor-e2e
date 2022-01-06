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
MQ-1783-fsx_ext4_stress
MQ-2330-ms_pod_disruption_rm_vol
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

# removed synchronous_replication, pvc_delete, maximum_vols_io, multiple_vols_pod_io, nexus_location, mayastorpool_schema
# control_plane_rescheduling, pvc_stress_fio, io_soak tests becuase tests are not stable in restul control plane
# removed check_mayastornode test as it's no longer valid in restful control plane
profiles[self_ci]="
basic_volume_io
csi
resource_check
ms_pod_restart
ms_pool_delete
volume_filesystem
dynamic_provisioning
expand_msp_disk
pvc_waitforfirstconsumer
pvc_readwriteonce
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
"

#hc1-nightly version of nightly-stable for CP2
# order is alphabetical, except for tests with long execution times
#   : primitive_msp_deletion, primitive_msp_state, node_failure, ms_pod_disruption,
#
profiles[hc1-nightly]="
primitive_msp_deletion
primitive_msp_state
node_failure
ms_pod_disruption
basic_volume_io
control_plane_rescheduling
csi
resource_check
dynamic_provisioning
expand_msp_disk
io_soak
mayastorpool_schema
MQ-1498-primitive_device_retirement
MQ-1783-fsx_ext4_stress
MQ-2219-rc-reconciliation
MQ-2330-ms_pod_disruption_rm_vol
MQ-2632-pvc_create_delete
maximum_vols_io
ms_pod_restart
ms_pod_disruption_no_io
ms_pool_delete
msv_rebuild
multiple_vols_pod_io
node_shutdown
nexus_location
pool_modify
primitive_data_integrity
primitive_fault_injection
primitive_fuzz_msv
primitive_msp_stress
primitive_replicas
primitive_volumes
pvc_delete
pvc_readwriteonce
pvc_stress_fio
pvc_waitforfirstconsumer
single_msn_shutdown
stale_msp_after_node_power_failure
synchronous_replication
volume_filesystem
"
#stale_msp_after_node_power_failure is moac specific

profiles[hc1-staging]="
"

profiles[experiment]="
volume_filesystem
dynamic_provisioning
"

