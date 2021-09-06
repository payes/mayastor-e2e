#!/bin/bash
set -e

PATH=/root/xfstests/bin:$PATH

test_dev=$1
scratch_dev=$2

mkfs -t xfs -f $test_dev
mkfs -t xfs -f $scratch_dev

mkdir ~/test_dir ~/scratch_dir
mount $test_dev ~/test_dir
mount $scratch_dev ~/scratch_dir

export TEST_DEV=$test_dev
export TEST_DIR=~/test_dir
export SCRATCH_DEV=$scratch_dev
export SCRATCH_MNT=~/scratch_dir

cd /root/xfstests; ./check -E /root/excluded_list -g xfs/quick

umount $test_dev
umount $scratch_dev
