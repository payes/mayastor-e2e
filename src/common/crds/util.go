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
var clientSet *v1alpha1Client.MayastorPoolV1Alpha1Client

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
		_ = v1alpha1Api.AddToScheme(scheme.Scheme)

		clientSet, err = v1alpha1Client.NewForConfig(config)
		if err != nil {
			fmt.Printf("Error %v", err)
		}
	},
	)
}

func CreatePool(poolName string, node string, disks []string) error {
	msp := v1alpha1Api.MayastorPool{
		TypeMeta: metaV1.TypeMeta{Kind: "MayastorPool"},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      poolName,
			Namespace: common.NSMayastor,
		},
		Spec: v1alpha1Api.MayastorPoolSpec{
			Node:  node,
			Disks: disks,
		},
	}
	_, err := clientSet.MayastorPools().Create(context.TODO(), &msp, metaV1.CreateOptions{})
	return err
}

func GetPool(poolName string) (v1alpha1Api.MayastorPool, error) {
	msp := v1alpha1Api.MayastorPool{}
	res, err := clientSet.MayastorPools().Get(context.TODO(), poolName, metaV1.GetOptions{})
	if res != nil && err == nil {
		msp = *res
	}
	return msp, err
}

func UpdatePool(msp v1alpha1Api.MayastorPool) error {
	_, err := clientSet.MayastorPools().Update(context.TODO(), &msp, metaV1.UpdateOptions{})
	return err
}

func DeletePool(poolName string) error {
	err := clientSet.MayastorPools().Delete(context.TODO(), poolName, metaV1.DeleteOptions{})
	return err
}

func ListPools() ([]v1alpha1Api.MayastorPool, error) {
	poolList, err := clientSet.MayastorPools().List(context.TODO(), metaV1.ListOptions{})
	if err != nil {
		return []v1alpha1Api.MayastorPool{}, err
	}
	return poolList.Items, nil
}
