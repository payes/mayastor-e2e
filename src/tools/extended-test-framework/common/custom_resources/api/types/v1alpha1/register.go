package v1alpha1

import (
	"mayastor-e2e/tools/extended-test-framework/common"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var PoolSchemeGroupVersion = schema.GroupVersion{Group: common.CRDGroupName, Version: common.CRDPoolGroupVersion}

var (
	PoolSchemeBuilder = runtime.NewSchemeBuilder(poolAddKnownTypes)
	PoolAddToScheme   = PoolSchemeBuilder.AddToScheme
)

func poolAddKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(PoolSchemeGroupVersion,
		&MayastorPool{},
		&MayastorPoolList{},
	)

	metaV1.AddToGroupVersion(scheme, PoolSchemeGroupVersion)
	return nil
}

var NodeSchemeGroupVersion = schema.GroupVersion{Group: common.CRDGroupName, Version: common.CRDNodeGroupVersion}

var (
	NodeSchemeBuilder = runtime.NewSchemeBuilder(nodeAddKnownTypes)
	NodeAddToScheme   = NodeSchemeBuilder.AddToScheme
)

func nodeAddKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(NodeSchemeGroupVersion,
		&MayastorNode{},
		&MayastorNodeList{},
	)

	metaV1.AddToGroupVersion(scheme, NodeSchemeGroupVersion)
	return nil
}

var VolumeSchemeGroupVersion = schema.GroupVersion{Group: common.CRDGroupName, Version: common.CRDVolumeGroupVersion}

var (
	VolumeSchemeBuilder = runtime.NewSchemeBuilder(volumeAddKnownTypes)
	VolumeAddToScheme   = VolumeSchemeBuilder.AddToScheme
)

func volumeAddKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(VolumeSchemeGroupVersion,
		&MayastorVolume{},
		&MayastorVolumeList{},
	)

	metaV1.AddToGroupVersion(scheme, VolumeSchemeGroupVersion)
	return nil
}
