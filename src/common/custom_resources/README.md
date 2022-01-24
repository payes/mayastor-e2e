# Mayastor Custom Resources and E2E testing
## Introduction
Mayastor defines and uses 3 custom resources.
The definitions are cotained in following files
* `mayastorvolume.yaml`
* `mayastorpool.yaml`
* `mayastornode.yaml`

These files are copied from the moac repository and represent the current shape
 of Mayastor custom resources

To use these resources we have implemented a clientset,
the sources of which are in `api` and `clientset` subdirectories,
and `util.go`.

A point worth noting  is that for E2E testing these CRDs should match the
 definitions contained in the yaml files within the install bundle.

This is achieved by extracting the install bundle,
and regenerating the type definitions to match.

The python script `genGoCrdTypes.py` generates the appropriate definitions from
the yaml files.

To generate the files run the following command lines (with
working directory set to the mayastor e2e repo root directory).

```
./scripts/genGoCrdTypes.py artifacts/install/XXXX/csi/moac/crds/mayastor*.yaml
```
In case of mayastor control plane , use below command to generate pool custom resource client code

```
./scripts/genGoCrdTypes.py artifacts/install-bundle/XXXX/mcp/chart/templates/mayastorpoolcrd.yaml
```

where `XXXX` is the mayastor build tag

These commands will update the following files
* `./api/types/v1alpha1/mayastorvolume.go`
* `./api/types/v1alpha1/mayastorpool.go`
* `./api/types/v1alpha1/mayastornode.go`

## Update procedure
* Copy the appropriate yaml files across from the moac repository to `src/common/custom_resources`
* The execute
  * ``` ./genGoCrdTypes.py ./mayastor*.yaml```
* Commit the changed files and raise a Pull Request
