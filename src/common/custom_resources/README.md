The following files
* `mayastorvolume.yaml`
* `mayastorpool.yaml`
* `mayastornode.yaml`
are copied from the mayastor repository and represent the current shape of
Mayastor custom resources implemented in the clientset

In particular the `struct` defines for `spec` and `status` 
must match the defintions in the aforementioned yaml files,
in these files
* `./api/types/v1alpha1/mayastorvolume.go`
* `./api/types/v1alpha1/mayastorpool.go`
* `./api/types/v1alpha1/mayastornode.go`


### Update procedure
1. Copy the appropriate yaml files across.
2. Edit the `struct` defines to match the yaml files
   * `mayastorvolume.yaml` -> `./api/types/v1alpha1/mayastorvolume.go`
   * `mayastorpool.yaml` -> `./api/types/v1alpha1/mayastorpool.go`
   * `mayastornode.yaml` -> `./api/types/v1alpha1/mayastornode.go`
