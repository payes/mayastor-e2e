# Extended Test Framework

This part of the mayastor-e2e repository contains source code to support testing in the extended test framework.
See https://mayadata.atlassian.net/wiki/spaces/MS/pages/edit-v2/848494593?draftShareId=0afa1b85-002c-40ea-a1c7-f7f0f5e19ac7

There are a number of components that control the test, each corresponding to a top-level directory.


## Test conductor

This carries out the test, typically deploying pods and PVCs, and monitoring the health of CRs.
It informs the test director of a test start and end, and sends workload-monitoring requests to the
workload monitor.


## Workload monitor

Monitors the state of pods requested from the test conductor. If a pod is in an unacceptable state,
it will send a failure event to the test director


## Test director

Gathers events from the other components and pushes test outcomes to e.g. XRay and Slack.


# To build

Build scripts are at <component>/scripts/build.sh
The script will create an image tagged "latest" and will push to the CI registry.

To regenerate the swagger code, call the corresponding gen_*_code.sh script in the scripts directory.


# To deploy


## Create a cluster

The cluster should have at least 4 worker nodes.

One node should be reserved for the ETFW actors, by applying the node label:

	openebs.io/role=mayastor-e2e

and by removing the label

	openebs.io/engine=mayastor

to prevent mayastor from using than node.


## Install mayastor

deploy mayastor on the cluster with an MSP defined for each node.


## Deploy the ETFW

Use the script at ./scripts/deploy.sh

The script needs three parameters:

The name of the test (e.g. "steady_state")

	-t <test name>

The Jira key of the associated Test Plan:

Note, the Test Plan must already contain the Test object of the test to be run.
The Jira key of the test is defined in the test_conductor config file defined
in ./deploy/test_conductor/\<test\>/config.yaml

	-p <plan ID>

How long to run the test with units (e.g. 60s / 5m50s / 36h):

	-d <duration>

## Remove the ETFW

    ./scripts/deploy.sh -r
