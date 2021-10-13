package controlplane

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"mayastor-e2e/common"
	"mayastor-e2e/common/controlplane/v0"
	cpv1_kubectl_plugin "mayastor-e2e/common/controlplane/v1"
	cpv1_rest_api "mayastor-e2e/common/controlplane/v1/rest-api"
	"mayastor-e2e/common/e2e_config"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type cpInterface interface {
	// Version

	MajorVersion() int
	Version() string

	IsTimeoutError(error) bool

	// Resource state strings abstraction

	VolStateHealthy() string
	VolStateDegraded() string
	ChildStateUnknown() string
	ChildStateOnline() string
	ChildStateDegraded() string
	ChildStateFaulted() string
	NexusStateUnknown() string
	NexusStateOnline() string
	NexusStateDegraded() string
	NexusStateFaulted() string
	MspStateOnline() string

	MspGrpcStateToCrdState(int) string

	// MSV abstraction

	GetMSV(uuid string) (*common.MayastorVolume, error)
	GetMsvNodes(uuid string) (string, []string)
	DeleteMsv(volName string) error
	ListMsvs() ([]common.MayastorVolume, error)
	SetMsvReplicaCount(uuid string, replicaCount int) error
	GetMsvState(uuid string) (string, error)
	GetMsvReplicas(volName string) ([]common.Replica, error)
	GetMsvNexusChildren(volName string) ([]common.NexusChild, error)
	GetMsvNexusState(uuid string) (string, error)
	IsMsvPublished(uuid string) bool
	IsMsvDeleted(uuid string) bool
	CheckForMsvs() (bool, error)
	CheckAllMsvsAreHealthy() error

	// Mayastor Node abstraction

	GetMSN(node string) (*common.MayastorNode, error)
	ListMsns() ([]common.MayastorNode, error)
	GetMsNodeStatus(node string) (string, error)

	NodeStateOffline() string
	NodeStateOnline() string
	NodeStateUnknown() string
	NodeStateEmpty() string

	// Mayastor Pool abstraction

	GetMsPool(poolName string) (*common.MayastorPool, error)
	ListMsPools() ([]common.MayastorPool, error)
}

var ifc cpInterface

var once sync.Once

func getControlPlane() cpInterface {
	once.Do(func() {
		version := e2e_config.GetConfig().MayastorVersion
		verComponents := strings.Split(version, ".")
		major, err := strconv.Atoi(verComponents[0])
		if err == nil {
			switch major {
			case 0:
				ifc, err = v0.MakeCP(major)
			case 1:
				if e2e_config.GetConfig().UseRestApiForControlPlaneV1 {
					logf.Log.Info("Using go rest API client for control plane V1")
					ifc, err = cpv1_rest_api.MakeCP(major)
				} else {
					logf.Log.Info("Using kubectl plugin for control plane V1")
					ifc, err = cpv1_kubectl_plugin.MakeCP(major)
				}
			default:
				err = fmt.Errorf("unsupported control plane version %v", version)
			}
		}
		if err != nil {
			panic(err)
		}
		if ifc == nil {
			panic("failed to set control plane object")
		}
	})
	return ifc
}

func VolStateHealthy() string {
	return getControlPlane().VolStateHealthy()
}

func VolStateDegraded() string {
	return getControlPlane().VolStateDegraded()
}

func ChildStateUnknown() string {
	return getControlPlane().ChildStateUnknown()
}

func ChildStateOnline() string {
	return getControlPlane().ChildStateOnline()
}

func ChildStateDegraded() string {
	return getControlPlane().ChildStateDegraded()
}

func ChildStateFaulted() string {
	return getControlPlane().ChildStateFaulted()
}

func NexusStateUnknown() string {
	return getControlPlane().NexusStateUnknown()
}

func NexusStateOnline() string {
	return getControlPlane().NexusStateOnline()
}

func NexusStateDegraded() string {
	return getControlPlane().NexusStateDegraded()
}

func NexusStateFaulted() string {
	return getControlPlane().NexusStateFaulted()
}

func MspStateOnline() string {
	return getControlPlane().MspStateOnline()
}

func MspGrpcStateToCrdState(mspState int) string {
	return getControlPlane().MspGrpcStateToCrdState(mspState)
}

//FIXME: MSV These functions are only guaranteed to
// work correctly if invoked from k8stest/msv.go
// which ensures that necessary setup functions
// have been called
// The issue is that for control plane v1 we need
// node addresses and the k8stest pkg provides that.

// GetMSV Get pointer to a mayastor volume custom resource
// returns nil and no error if the msv is in pending state.
func GetMSV(uuid string) (*common.MayastorVolume, error) {
	return getControlPlane().GetMSV(uuid)
}

// GetMsvNodes Retrieve the nexus node hosting the Mayastor Volume,
// and the names of the replica nodes
// function asserts if the volume CR is not found.
func GetMsvNodes(uuid string) (string, []string) {
	return getControlPlane().GetMsvNodes(uuid)
}

func DeleteMsv(volName string) error {
	return getControlPlane().DeleteMsv(volName)
}

func ListMsvs() ([]common.MayastorVolume, error) {
	return getControlPlane().ListMsvs()
}

func SetMsvReplicaCount(uuid string, replicaCount int) error {
	return getControlPlane().SetMsvReplicaCount(uuid, replicaCount)
}

func GetMsvState(uuid string) (string, error) {
	return getControlPlane().GetMsvState(uuid)
}

func GetMsvReplicas(volName string) ([]common.Replica, error) {
	return getControlPlane().GetMsvReplicas(volName)
}

func GetMsvNexusChildren(volName string) ([]common.NexusChild, error) {
	return getControlPlane().GetMsvNexusChildren(volName)
}

func GetMsvNexusState(uuid string) (string, error) {
	return getControlPlane().GetMsvNexusState(uuid)
}

func IsMsvPublished(uuid string) bool {
	return getControlPlane().IsMsvPublished(uuid)
}

func IsMsvDeleted(uuid string) bool {
	return getControlPlane().IsMsvDeleted(uuid)
}

func CheckForMsvs() (bool, error) {
	return getControlPlane().CheckForMsvs()
}

func CheckAllMsvsAreHealthy() error {
	return getControlPlane().CheckAllMsvsAreHealthy()
}

//FIXME: MSN These functions are only guaranteed to
// work correctly if invoked from k8stest/msn.go
// which ensures that necessary setup functions
// have been called
// The issue is that for control plane v1 we need
// node addresses and the k8stest pkg provides that.

// GetMSN Get pointer to a mayastor node custom resource
// returns nil and no error if the msn is in pending state.
func GetMSN(nodeName string) (*common.MayastorNode, error) {
	return getControlPlane().GetMSN(nodeName)
}

func ListMsns() ([]common.MayastorNode, error) {
	return getControlPlane().ListMsns()
}
func GetMsNodeStatus(nodeName string) (string, error) {
	return getControlPlane().GetMsNodeStatus(nodeName)
}

func NodeStateOnline() string {
	return getControlPlane().NodeStateOnline()
}

// NodeStateOffline is set when the node misses its watchdog deadline
func NodeStateOffline() string {
	return getControlPlane().NodeStateOffline()
}

// NodeStateUnknown is set if the mayastor instance deregisters itself (when the pod goes down gracefully),
// or if there's an error when we're issuing issuing "list" requests
func NodeStateUnknown() string {
	return getControlPlane().NodeStateUnknown()
}

// NodeStateEmpty i.e. no state at all if the control plane restarts and the node is not available at that time
func NodeStateEmpty() string {
	return getControlPlane().NodeStateEmpty()
}

func Version() string {
	return getControlPlane().Version()
}

func MajorVersion() int {
	return getControlPlane().MajorVersion()
}

func IsTimeoutError(err error) bool {
	return getControlPlane().IsTimeoutError(err)
}

func GetMsPool(poolName string) (*common.MayastorPool, error) {
	return getControlPlane().GetMsPool(poolName)
}

func ListMsPools() ([]common.MayastorPool, error) {
	return getControlPlane().ListMsPools()
}
