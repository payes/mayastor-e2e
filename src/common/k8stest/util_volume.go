package k8stest

import (
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// Volume is a wrapper over named volume api object, used
// within Pods. It provides build, validations and other common
// logic to be used by various feature specific callers.
type Volume struct {
	object *corev1.Volume
}

// Builder is the builder object for Volume
type VolumeBuilder struct {
	volume *Volume
	errs   []error
}

// NewBuilder returns new instance of Builder
func NewVolumeBuilder() *VolumeBuilder {
	return &VolumeBuilder{volume: &Volume{object: &corev1.Volume{}}}
}

// WithName sets the Name field of Volume with provided value.
func (b *VolumeBuilder) WithName(name string) *VolumeBuilder {
	if len(name) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build Volume object: missing Volume name"),
		)
		return b
	}
	b.volume.object.Name = name
	return b
}

// WithHostDirectory sets the VolumeSource field of Volume with provided hostpath
// as type directory.
func (b *VolumeBuilder) WithHostDirectory(path string) *VolumeBuilder {
	if len(path) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: missing volume path"),
		)
		return b
	}
	volumeSource := corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: path,
		},
	}

	b.volume.object.VolumeSource = volumeSource
	return b
}

//WithSecret sets the VolumeSource field of Volume with provided Secret
func (b *VolumeBuilder) WithSecret(secret *corev1.Secret, defaultMode int32) *VolumeBuilder {
	dM := defaultMode
	if secret == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil ConfigMap"),
		)
		return b
	}
	if defaultMode == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: missing defaultmode"),
		)
	}
	volumeSource := corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			DefaultMode: &dM,
			SecretName:  secret.Name,
		},
	}
	b.volume.object.VolumeSource = volumeSource
	b.volume.object.Name = secret.Name
	return b
}

//WithConfigMap sets the VolumeSource field of Volume with provided ConfigMap
func (b *VolumeBuilder) WithConfigMap(configMap *corev1.ConfigMap, defaultMode int32) *VolumeBuilder {
	dM := defaultMode
	if configMap == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil ConfigMap"),
		)
		return b
	}
	if defaultMode == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: missing defaultmode"),
		)
	}
	volumeSource := corev1.VolumeSource{
		ConfigMap: &corev1.ConfigMapVolumeSource{
			DefaultMode: &dM,
			LocalObjectReference: corev1.LocalObjectReference{
				Name: configMap.Name,
			},
		},
	}
	b.volume.object.VolumeSource = volumeSource
	b.volume.object.Name = configMap.Name
	return b
}

// WithHostPathAndType sets the VolumeSource field of Volume with provided
// hostpath as directory path and type as directory type
func (b *VolumeBuilder) WithHostPathAndType(
	dirpath string,
	dirtype *corev1.HostPathType,
) *VolumeBuilder {
	if dirtype == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil volume type"),
		)
		return b
	}
	if len(dirpath) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: missing volume path"),
		)
		return b
	}
	newdirtype := *dirtype
	volumeSource := corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{
			Path: dirpath,
			Type: &newdirtype,
		},
	}

	b.volume.object.VolumeSource = volumeSource
	return b
}

// WithPVCSource sets the Volume field of Volume with provided pvc
func (b *VolumeBuilder) WithPVCSource(pvcName string) *VolumeBuilder {
	if len(pvcName) == 0 {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: missing pvc name"),
		)
		return b
	}
	volumeSource := corev1.VolumeSource{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: pvcName,
		},
	}
	b.volume.object.VolumeSource = volumeSource
	return b
}

// WithEmptyDir sets the EmptyDir field of the Volume with provided dir
func (b *VolumeBuilder) WithEmptyDir(dir *corev1.EmptyDirVolumeSource) *VolumeBuilder {
	if dir == nil {
		b.errs = append(
			b.errs,
			errors.New("failed to build volume object: nil dir"),
		)
		return b
	}

	newdir := *dir
	b.volume.object.EmptyDir = &newdir
	return b
}

// Build returns the Volume API instance
func (b *VolumeBuilder) Build() (*corev1.Volume, error) {
	if len(b.errs) > 0 {
		return nil, errors.Errorf("%+v", b.errs)
	}
	return b.volume.object, nil
}
