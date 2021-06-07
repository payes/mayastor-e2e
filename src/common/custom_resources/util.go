package custom_resources

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sync"

	"mayastor-e2e/common"
	v1alpha1Api "mayastor-e2e/common/custom_resources/api/types/v1alpha1"
	v1alpha1Client "mayastor-e2e/common/custom_resources/clientset/v1alpha1"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var csOnce sync.Once
var poolClientSet *v1alpha1Client.MayastorPoolV1Alpha1Client
var nodeClientSet *v1alpha1Client.MayastorNodeV1Alpha1Client
var volClientSet *v1alpha1Client.MayastorVolumeV1Alpha1Client

func init() {
	csOnce.Do(func() {
		useCluster := true
		testEnv := &envtest.Environment{
			UseExistingCluster: &useCluster,
		}
		config, err := testEnv.Start()
		if err != nil {
			fmt.Printf("Error %v", err)
		}
		_ = v1alpha1Api.PoolAddToScheme(scheme.Scheme)
		_ = v1alpha1Api.NodeAddToScheme(scheme.Scheme)
		_ = v1alpha1Api.VolumeAddToScheme(scheme.Scheme)

		poolClientSet, err = v1alpha1Client.MspNewForConfig(config)
		if err != nil {
			fmt.Printf("Error %v", err)
		}
		nodeClientSet, err = v1alpha1Client.MsnNewForConfig(config)
		if err != nil {
			fmt.Printf("Error %v", err)
		}
		volClientSet, err = v1alpha1Client.MsvNewForConfig(config)
		if err != nil {
			fmt.Printf("Error %v", err)
		}
	},
	)
}

// == Mayastor Pool  ======================

func CreateMsPool(poolName string, node string, disks []string) (*v1alpha1Api.MayastorPool, error) {
	msp := v1alpha1Api.MayastorPool{
		TypeMeta: metaV1.TypeMeta{Kind: "MayastorPool"},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      poolName,
			Namespace: common.NSMayastor(),
		},
		Spec: v1alpha1Api.MayastorPoolSpec{
			Node:  node,
			Disks: disks,
		},
	}
	mspOut, err := poolClientSet.MayastorPools().Create(context.TODO(), &msp, metaV1.CreateOptions{})
	return mspOut, err
}

func GetMsPool(poolName string) (v1alpha1Api.MayastorPool, error) {
	msp := v1alpha1Api.MayastorPool{}
	res, err := poolClientSet.MayastorPools().Get(context.TODO(), poolName, metaV1.GetOptions{})
	if res != nil && err == nil {
		msp = *res
	}
	return msp, err
}

func UpdateMsPool(mspIn v1alpha1Api.MayastorPool) (*v1alpha1Api.MayastorPool, error) {
	mspOut, err := poolClientSet.MayastorPools().Update(context.TODO(), &mspIn, metaV1.UpdateOptions{})
	return mspOut, err
}

func DeleteMsPool(poolName string) error {
	err := poolClientSet.MayastorPools().Delete(context.TODO(), poolName, metaV1.DeleteOptions{})
	return err
}

func ListMsPools() ([]v1alpha1Api.MayastorPool, error) {
	poolList, err := poolClientSet.MayastorPools().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []v1alpha1Api.MayastorPool{}, err
	}
	return poolList.Items, nil
}

// CheckAllMsPoolsAreOnline checks if all mayastor pools are online
func CheckAllMsPoolsAreOnline() error {
	allHealthy := true
	pools, err := ListMsPools()
	if err == nil && pools != nil && len(pools) != 0 {
		for _, pool := range pools {
			poolName := pool.GetName()
			state := pool.Status.State
			if state != "online" {
				log.Log.Info("CheckAllMsPoolsAreOnline", "pool", poolName, "state", state)
				allHealthy = false
			}
		}
	}

	if !allHealthy {
		return fmt.Errorf("all pools were not healthy")
	}
	return err
}

// == Mayastor Nodes ======================

func GetMsNode(nodeName string) (v1alpha1Api.MayastorNode, error) {
	msn := v1alpha1Api.MayastorNode{}
	res, err := nodeClientSet.MayastorNodes().Get(context.TODO(), nodeName, metaV1.GetOptions{})
	if res != nil && err == nil {
		msn = *res
	}
	return msn, err
}

func ListMsNodes() ([]v1alpha1Api.MayastorNode, error) {
	nodeList, err := nodeClientSet.MayastorNodes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []v1alpha1Api.MayastorNode{}, err
	}
	return nodeList.Items, nil
}

func DeleteMsNode(nodeName string) error {
	err := nodeClientSet.MayastorNodes().Delete(context.TODO(), nodeName, metaV1.DeleteOptions{})
	return err
}

// CheckMsNodesAreOnline checks all the mayastor nodes are
// in online state or not if any of the msn is not in
// online state then it returns error.
func CheckMsNodesAreOnline() error {
	msns, err := ListMsNodes()
	if err == nil && msns != nil && len(msns) != 0 {
		for _, msn := range msns {
			if msn.Status != "online" {
				return fmt.Errorf("mayastornodes were not online")
			}
		}
	}
	return err
}

// == Mayastor Volumes ======================

//  MOAC/Mayastor creates Mayastor Volumes, create use case?

func GetMsVol(volName string) (*v1alpha1Api.MayastorVolume, error) {
	msv := v1alpha1Api.MayastorVolume{}
	res, err := volClientSet.MayastorVolumes().Get(context.TODO(), volName, metaV1.GetOptions{})
	if res != nil && err == nil {
		msv = *res
	}
	return &msv, err
}

func UpdateMsVol(msvIn *v1alpha1Api.MayastorVolume) (*v1alpha1Api.MayastorVolume, error) {
	msvOut, err := volClientSet.MayastorVolumes().Update(context.TODO(), msvIn, metaV1.UpdateOptions{})
	return msvOut, err
}

func DeleteMsVol(volName string) error {
	err := volClientSet.MayastorVolumes().Delete(context.TODO(), volName, metaV1.DeleteOptions{})
	return err
}

func ListMsVols() ([]v1alpha1Api.MayastorVolume, error) {
	volList, err := volClientSet.MayastorVolumes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []v1alpha1Api.MayastorVolume{}, err
	}
	return volList.Items, nil
}

// Helper functions
// GetMsVolState convenience function to retrieve the volume state.
func GetMsVolState(volName string) (string, error) {
	msv, err := GetMsVol(volName)
	if err == nil {
		return msv.Status.State, nil
	}
	return "", err
}

// GetMsVolReplicas convenience function to retrieve the list of replicas comprising the volume.
func GetMsVolReplicas(volName string) ([]v1alpha1Api.Replica, error) {
	msv, err := GetMsVol(volName)
	if err != nil {
		return nil, err
	}
	return msv.Status.Replicas, nil
}

// UpdateMsVolReplicaCount update the number of replicas in a volume.
func UpdateMsVolReplicaCount(volName string, replicaCount int) error {
	msv, err := GetMsVol(volName)
	if err != nil {
		return err
	}
	msv.Spec.ReplicaCount = replicaCount
	_, err = UpdateMsVol(msv)
	return err
}

// GetMsVolNexusChildren convenience function to retrieve the nexus children of a volume.
func GetMsVolNexusChildren(volName string) ([]v1alpha1Api.NexusChild, error) {
	msv, err := GetMsVol(volName)
	if err != nil {
		return nil, err
	}
	return msv.Status.Nexus.Children, nil
}

// GetMsVolNexusState returns the nexus state from the MSV.
// An error is returned if the nexus state cannot be retrieved.
func GetMsVolNexusState(uuid string) (string, error) {
	msv, err := GetMsVol(uuid)
	if err != nil {
		return "", err
	}
	return msv.Status.State, nil
}

// IsMsVolPublished returns true if the volume is published.
// A volume is published if the "targetNodes" field exists in the MSV.
func IsMsVolPublished(uuid string) bool {
	msv, err := GetMsVol(uuid)
	if err == nil {
		return 0 == len(msv.Status.TargetNodes)
	}
	return false
}

// IsMsVolDeleted check for a deleted Mayastor Volume, the object does not exist if deleted
func IsMsVolDeleted(uuid string) bool {
	_, err := GetMsVol(uuid)
	if err != nil && errors.IsNotFound(err) {
		return true
	}
	return false
}

// CheckForMsVols checks if any mayastor volumes exists
func CheckForMsVols() (bool, error) {
	log.Log.Info("CheckForMsVols")
	foundResources := false

	msvs, err := ListMsVols()
	if err == nil && msvs != nil && len(msvs) != 0 {
		log.Log.Info("CheckForVolumeResources: found MayastorVolumes",
			"MayastorVolumes", msvs)
		foundResources = true
	}
	return foundResources, err
}

// CheckAllMsVolsAreHealthy checks if all existing mayastor volumes are in healthy state
func CheckAllMsVolsAreHealthy() error {
	allHealthy := true
	msvs, err := ListMsVols()
	if err == nil && msvs != nil && len(msvs) != 0 {
		for _, msv := range msvs {
			if msv.Status.State != "healthy" {
				log.Log.Info("CheckAllMsVolsAreHealthy", "msv", msv)
				allHealthy = false
			}
		}
	}

	if !allHealthy {
		return fmt.Errorf("all MSVs were not healthy")
	}
	return err
}
