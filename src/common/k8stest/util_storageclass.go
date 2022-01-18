package k8stest

import (
	"context"
	"mayastor-e2e/common"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"

	"github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StorageClass struct {
	object *storagev1.StorageClass
}

// ScBuilder enables building an instance of StorageClass
type ScBuilder struct {
	sc   *StorageClass
	errs []error
}

// NewScBuilder returns new instance of ScBuilder
func NewScBuilder() *ScBuilder {
	obj := ScBuilder{sc: &StorageClass{object: &storagev1.StorageClass{}}}
	// set default mayastor csi provisioner
	scObject := obj.WithProvisioner(common.CSIProvisioner)

	// set default replicas value i.e 1
	scObject = scObject.WithReplicas(common.DefaultReplicaCount)
	return scObject
}

// WithName sets the Name field of storageclass with provided argument.
func (b *ScBuilder) WithName(name string) *ScBuilder {
	if len(name) == 0 {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing storageclass name"))
		return b
	}
	b.sc.object.Name = name
	return b
}

// WithNamespace sets the namespace field of storageclass with provided argument.
func (b *ScBuilder) WithNamespace(ns string) *ScBuilder {
	if len(ns) == 0 {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing storageclass namespace"))
		return b
	}
	b.sc.object.Namespace = ns
	return b
}

// WithGenerateName appends a random string after the name
func (b *ScBuilder) WithGenerateName(name string) *ScBuilder {
	b.sc.object.GenerateName = name + "-"
	return b
}

// WithAnnotations sets the Annotations field of storageclass with provided value.
func (b *ScBuilder) WithAnnotations(annotations map[string]string) *ScBuilder {
	if len(annotations) == 0 {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing annotations"))
	}
	b.sc.object.Annotations = annotations
	return b
}

// WithReplicas sets the replica parameter of storageclass with provided argument.
func (b *ScBuilder) WithReplicas(value int) *ScBuilder {
	if value == 0 {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing storageclass replicas"))
		return b
	}
	if b.sc.object.Parameters == nil {
		b.sc.object.Parameters = map[string]string{}
	}
	b.sc.object.Parameters[string(common.ScReplicas)] = strconv.Itoa(value)
	return b
}

// WithFileType sets the fsType parameter of storageclass with provided argument.
func (b *ScBuilder) WithFileSystemType(value common.FileSystemType) *ScBuilder {
	if value != common.NoneFsType {
		if b.sc.object.Parameters == nil {
			b.sc.object.Parameters = map[string]string{}
		}
		b.sc.object.Parameters[string(common.ScFsType)] = string(value)
	}
	return b
}

// WithProtocol sets the protocol parameter of storageclass with provided argument.
func (b *ScBuilder) WithProtocol(value common.ShareProto) *ScBuilder {
	if b.sc.object.Parameters == nil {
		b.sc.object.Parameters = map[string]string{}
	}
	b.sc.object.Parameters[string(common.ScProtocol)] = string(value)
	return b
}

// WithProtocol sets the protocol parameter of storageclass with provided argument.
func (b *ScBuilder) WithLocal(value bool) *ScBuilder {
	if b.sc.object.Parameters == nil {
		b.sc.object.Parameters = map[string]string{}
	}
	b.sc.object.Parameters[string(common.ScLocal)] = strconv.FormatBool(value)
	return b
}

// WithProtocol sets the protocol parameter of storageclass with provided argument.
func (b *ScBuilder) WithIOTimeout(value int) *ScBuilder {
	if b.sc.object.Parameters == nil {
		b.sc.object.Parameters = map[string]string{}
	}
	b.sc.object.Parameters[string(common.IOTimeout)] = strconv.Itoa(value)
	return b
}

// WithProvisioner sets the Provisioner field of storageclass with provided argument.
func (b *ScBuilder) WithProvisioner(provisioner string) *ScBuilder {
	if len(provisioner) == 0 {
		b.errs = append(b.errs, errors.New("failed to build storageclass: missing provisioner name"))
		return b
	}
	b.sc.object.Provisioner = provisioner
	return b
}

// Build returns the StorageClass API instance
func (b *ScBuilder) Build() (*storagev1.StorageClass, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	if b.sc.object.Parameters[common.ScProtocol] == string(common.ShareProtoNvmf) && b.sc.object.Parameters[common.IOTimeout] == "" {
		b.sc.object.Parameters[common.IOTimeout] = "30"
		logf.Log.Info("NewScBuilder: \"defaulting\" ioTimeout", "value", b.sc.object.Parameters[common.IOTimeout])
	}
	return b.sc.object, nil
}

// WithVolumeBindingMode sets the volume binding mode of storageclass with
// provided argument.
// VolumeBindingMode indicates how PersistentVolumeClaims should be bound.
// VolumeBindingImmediate indicates that PersistentVolumeClaims should be
// immediately provisioned and bound. This is the default mode.
// VolumeBindingWaitForFirstConsumer indicates that PersistentVolumeClaims
// should not be provisioned and bound until the first Pod is created that
// references the PeristentVolumeClaim.  The volume provisioning and
// binding will occur during Pod scheduing.
func (b *ScBuilder) WithVolumeBindingMode(bindingMode storagev1.VolumeBindingMode) *ScBuilder {
	if bindingMode != "" {
		b.sc.object.VolumeBindingMode = &bindingMode
	}
	return b
}

//Build and create the StorageClass
func (b *ScBuilder) BuildAndCreate() error {
	scObj, err := b.Build()
	if err == nil {
		err = CreateSc(scObj)
	}
	return err
}

// CreateSc creates storageclass with provided storageclass object
func CreateSc(obj *storagev1.StorageClass) error {
	logf.Log.Info("Creating", "StorageClass", obj)
	ScApi := gTestEnv.KubeInt.StorageV1().StorageClasses
	_, createErr := ScApi().Create(context.TODO(), obj, metaV1.CreateOptions{})
	return createErr
}
