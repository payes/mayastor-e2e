package crds

import (
	"context"
	"fmt"
	"sync"

	"mayastor-e2e/common"
	v1alpha1Api "mayastor-e2e/common/crds/api/types/v1alpha1"
	v1alpha1Client "mayastor-e2e/common/crds/clientset/v1alpha1"

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

// == Maystor Pool  ======================

func CreatePool(poolName string, node string, disks []string) (*v1alpha1Api.MayastorPool, error) {
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

func GetPool(poolName string) (v1alpha1Api.MayastorPool, error) {
	msp := v1alpha1Api.MayastorPool{}
	res, err := poolClientSet.MayastorPools().Get(context.TODO(), poolName, metaV1.GetOptions{})
	if res != nil && err == nil {
		msp = *res
	}
	return msp, err
}

func UpdatePool(mspIn v1alpha1Api.MayastorPool) (*v1alpha1Api.MayastorPool, error) {
	mspOut, err := poolClientSet.MayastorPools().Update(context.TODO(), &mspIn, metaV1.UpdateOptions{})
	return mspOut, err
}

func DeletePool(poolName string) error {
	err := poolClientSet.MayastorPools().Delete(context.TODO(), poolName, metaV1.DeleteOptions{})
	return err
}

func ListPools() ([]v1alpha1Api.MayastorPool, error) {
	poolList, err := poolClientSet.MayastorPools().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []v1alpha1Api.MayastorPool{}, err
	}
	return poolList.Items, nil
}

// == Maystor Nodes ======================

func GetNode(nodeName string) (v1alpha1Api.MayastorNode, error) {
	msn := v1alpha1Api.MayastorNode{}
	res, err := nodeClientSet.MayastorNodes().Get(context.TODO(), nodeName, metaV1.GetOptions{})
	if res != nil && err == nil {
		msn = *res
	}
	return msn, err
}

func ListNodes() ([]v1alpha1Api.MayastorNode, error) {
	nodeList, err := nodeClientSet.MayastorNodes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []v1alpha1Api.MayastorNode{}, err
	}
	return nodeList.Items, nil
}

func DeleteNode(nodeName string) error {
	err := nodeClientSet.MayastorNodes().Delete(context.TODO(), nodeName, metaV1.DeleteOptions{})
	return err
}

// == Maystor Volumes ======================

//  MOAC/Mayastor creates Mayastor Volumes, create use case?

func GetVolume(volName string) (v1alpha1Api.MayastorVolume, error) {
	msv := v1alpha1Api.MayastorVolume{}
	res, err := volClientSet.MayastorVolumes().Get(context.TODO(), volName, metaV1.GetOptions{})
	if res != nil && err == nil {
		msv = *res
	}
	return msv, err
}

func UpdateVolume(msvIn v1alpha1Api.MayastorVolume) (*v1alpha1Api.MayastorVolume, error) {
	msvOut, err := volClientSet.MayastorVolumes().Update(context.TODO(), &msvIn, metaV1.UpdateOptions{})
	return msvOut, err
}

func DeleteVolume(volName string) error {
	err := volClientSet.MayastorVolumes().Delete(context.TODO(), volName, metaV1.DeleteOptions{})
	return err
}

func ListVolumes() ([]v1alpha1Api.MayastorVolume, error) {
	volList, err := volClientSet.MayastorVolumes().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []v1alpha1Api.MayastorVolume{}, err
	}
	return volList.Items, nil
}
