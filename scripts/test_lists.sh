#!/usr/bin/env bash

# List and Sequence of tests.
# Restrictions:
#   1. resource_check MUST follow csi
#       resource_check is a follow up check for the 3rd party CSI test suite.
#   2. ms_pod_disruption SHOULD be the last test
#

DEFAULT_TEST_LIST="
basic_volume_io
csi
resource_check
replica
rebuild
ms_pod_disruption"
CONTINUOUS_TEST_LIST="
basic_volume_io
csi
resource_check
rebuild
io_soak
volume_filesystem
ms_pod_disruption"
NIGHTLY_TEST_LIST="
basic_volume_io
csi
resource_check
io_soak
multiple_vols_pod_io"
NIGHTLY_FULL_TEST_LIST="
basic_volume_io
csi
resource_check
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
dynamic_provisioning
check_mayastornode
pvc_waitforfirstconsumer"
ONDEMAND_TEST_LIST="
basic_volume_io
csi
resource_check"
SELF_CI_TEST_LIST="
basic_volume_io
csi
resource_check
pvc_stress_fio
io_soak
multiple_vols_pod_io
ms_pod_restart
check_mayastornode
pvc_waitforfirstconsumer"
SOAK_TEST_LIST="
io_soak"

