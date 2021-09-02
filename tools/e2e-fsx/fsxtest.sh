#!/bin/sh

set -e

# usage:  fsxtest $1 $2 $3
# $1 scratch device to use for testing
# $2 optional file system type
# $3 number of operations to perform

#Uncomment line below for debugging
#set -x
if [ $2 = "jfs" ]; then
  mkfs -t $2 -q $1
else
  mkfs -t $2 $1
fi
mkdir -p /testmount
mount -t $2 $1 /testmount
touch /testmount/testfile
fsx-linux -N $3 /testmount/testfile
RESULT=$?
# report the results
if [ $RESULT -eq "0" ]; then
  echo "PASS: fsxtest $1 $2 $3"
else
  echo "FAIL: fsxtest $1 $2 $3"
fi
umount /testmount
rm -rf /testmount
fsck -a -t $2 $1  # report the results
# e2fsck -f -y -C0 $1  # report the results
# while using above e2fsck command, it's throwing below error
# e2fsck: need terminal for interactive repairs
exit $RESULT