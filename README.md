# Mayastor E2E
This repository contains the scripts and source files for end-to-end testing of Mayastor

On first checkout run `pre-commit install` this will enable running of pre-commit checks when committing changes.
You may need to install `pre-commit` on the host system.

## Directories
* src - source for go tests
* configurations - configuration for e2e platforms
* loki - deployment files for loki
* e2e-fio - files to build e2e-fio docker image

# Pre-requisites

The test doesn't yet manage the lifecycle of the cluster being tested,
therefore the test hosts' kubeconfig must point to a Kubernetes cluster.
You can verify that the kubeconfig is setup correctly simply with
`kubectl get nodes`.

The cluster under test must meet the following requirements:
* Have 3 nodes (4 nodes if running reassignment test)
* Each node must be configured per the quick start:
  * At least 1GiB of hugepages available (e.g. 512 2MiB hugepages)
  * Each node must be labelled for use by mayastor (ie "openebs.io/engine=mayastor")

The test host must have the following installed:
* go (>= v1.15)
* kubectl (tested with v1.18)
* helm

# Setting up the cluster

### Create a cluster in public cloud

Use terraform script in `aws-kubeadm/` from
[terraform repo](https://github.com/mayadata-io/mayastor-terraform-playground) to
set up a cluster in AWS suitable for running the tests. The configuration file for
terraform could look like this (replace docker registry value by yours - must
be reachable from the cluster):

```
cluster_name    = "<NICKNAME>-e2e-test-cluster"
deploy_mayastor = false
num_workers     = 3
ebs_volume_size = 5
mayastor_use_develop_images = true
aws_instance_root_size_gb   = 10
docker_insecure_registry    = "52.58.174.24:5000"
```

### Create local cluster

Many possibilities here. You could use libvirt based cluster created by
terraform (in terraform/ in this repo).

# Running the tests

If you'd like to run the tests as a whole (as they are run in our CI/CD
pipeline) then use the script `./scripts/e2e-test.sh`.

To run particular test cd to the directory with tests and type `go test .  -ginkgo.v -ginkgo.progress -timeout 0`
Most of the tests assume that mayastor is already installed. `install` test
can be run to do that.
Note some tests require deletion of pools and reconfiguration of pools, these tests will only work if
* The mayastor nodes are homogenous in terms of pool devices
  * single pool device
  * the pool device name is the same on all nodes
* The environment variable `e2e_pool_device` is set

List of these tests (not exhaustive):
* `pool_modify_test`
* `mayastorpool_schema`
* `ms_pool_delete`
* `ms_pod_disruption_rm_msv`
* `primitive_msp_stress`
* `nexus_location`
* `expand_msp_disk`

# Configuration
The e2e test suite support runtime configuration to set parameters and variables,
to suit the cluster under test.
When the test suites are run using `./scripts/e2e-test.sh`
* The configuration file can be specified using the `--config`
option
* The platform configuration file can be specified using the `--platform_config` option

Alternatively set the following environment variables
 * `e2e_config_file`
 * `e2e_platform_config_file`

The go package used to read (and write) configuration files supports the following formats
1. `yaml`
2. `json`
3. `toml`

The mayastor e2e project uses `yaml` exclusively.
If configurations are not specified the configuration is defaulted to values used for `CI/CD`.

The contents of the configuration files and the defaulted values  are described in
`./common/e2e_config/e2e_config.go`
Note this is subject to change, usually additions.

The test code "searches" for configurations specified if the environment variables are not absolute paths to configuration files.

Configuration files are typically loacted in `/configurations`.

Platform configuration files are typically located in `/configurations/platforms`

To use configuration files in these locations it is sufficient to set the environment variables just to the names of the files.

The test scripts will fail (panic) if
1. A configuration is specified and the specified configuration file cannot be found on the filesystem or the configuration directory
2. The contents of specified configuration are invalid

Once the configuration has been loaded and all fields resolved, the contents are written out to a file, typically in the `artifacts` directory.
The full path to the file will be printed on the console.

# Reports
Reports in the `junit/xml` format will be generate only if a reports directory is specified
 * by environment variable `e2e_reports_dir`
 * in the loaded configuration file

# Artefacts
Artefacts generated a part of the test will be generated in a subdirecrtory under `<artifacts>/sessions` directory

The `<artifacts>` directory is either
 * the `artifacts` subdirecrity under the repo root directory - if the go script successfully detected the repo root path
 * `/tmp/mayastor-e2e`

The name of the subdirectory under `<artifacts>/sessions` is one of
 * environment variable `e2e_session_dir`
 * loaded from the configuration file
 * `default`

 # Environment variables
 * `e2e_root_dir` Root directory of the `mayastor-e2e` repo. If not specified a sub path `mayastor-e2e/src` is searched for in the path of the `go` file running the code and set if found.
 * `e2e_config_file` name of configuration file in `configurations` or full path to configuration file
 * `e2e_platform_config_file` name of platform configuration file in `configurations/platforms` or full path to platform configuration file
 * `e2e_mayastor_root_dir` absolute path to `mayastor` repo, only required for `install` and `uninstall` tests
 * `e2e_session_dir` absolute path for session artifacts
 * `e2e_reports_dir` absolute path for generated reports
 * `e2e_docker_registry` Registry from where mayastor images are retrieved
 * `e2e_image_tag ` docker image tags for mayastor images
 * `e2e_pool_device` pool device used by mayastor, required for `install` and some disruptive tests which modify/delete/recreate pools on the test cluster. These tests are best run on a cluster where mayastor is installed using the `install` test
 * `e2eFioImage` docker image name of the mayastor e2e test pod
 * `e2e_default_replica_count` default replica count for volumes created by the tests
 * `e2e_defer_asserts` defer asserts until after cleanup, so that all `It` clauses can be excercised. Only one test uses this at the moment.
 * `e2e_uninstall_cleanup` flag for `uninstall` test
 * `e2e_policy_cleanup_before` experimental flag
