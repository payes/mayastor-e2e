package v1alpha1

import (
	"mayastor-e2e/common/e2e_config"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var PoolSchemeGroupVersion = schema.GroupVersion{Group: e2e_config.GetConfig().Product.CrdGroupName,
	Version: e2e_config.GetConfig().Product.CrdPoolGroupVersion}

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

var NodeSchemeGroupVersion = schema.GroupVersion{Group: e2e_config.GetConfig().Product.CrdGroupName,
	Version: e2e_config.GetConfig().Product.CrdNodeGroupVersion}

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

var VolumeSchemeGroupVersion = schema.GroupVersion{Group: e2e_config.GetConfig().Product.CrdGroupName,
	Version: e2e_config.GetConfig().Product.CrdVolumeGroupVersion}
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
