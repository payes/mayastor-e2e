package k8stest

import (
	errors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

type container struct {
	corev1.Container // kubernetes container type
}

// OptionFunc is a typed function that abstracts anykind of operation
// against the provided container instance
//
// This is the basic building block to create functional operations
// against the container instance
type OptionFunc func(*container)

// asContainer transforms this container instance into corresponding kubernetes
// container type
func (c *container) asContainer() corev1.Container {
	return corev1.Container{
		Name:                     c.Name,
		Image:                    c.Image,
		Command:                  c.Command,
		Args:                     c.Args,
		WorkingDir:               c.WorkingDir,
		Ports:                    c.Ports,
		EnvFrom:                  c.EnvFrom,
		Env:                      c.Env,
		Resources:                c.Resources,
		VolumeMounts:             c.VolumeMounts,
		VolumeDevices:            c.VolumeDevices,
		LivenessProbe:            c.LivenessProbe,
		ReadinessProbe:           c.ReadinessProbe,
		Lifecycle:                c.Lifecycle,
		TerminationMessagePath:   c.TerminationMessagePath,
		TerminationMessagePolicy: c.TerminationMessagePolicy,
		ImagePullPolicy:          c.ImagePullPolicy,
		SecurityContext:          c.SecurityContext,
		Stdin:                    c.Stdin,
		StdinOnce:                c.StdinOnce,
		TTY:                      c.TTY,
	}
}

// New returns a new kubernetes container
func NewContainer(opts ...OptionFunc) corev1.Container {
	c := &container{}
	for _, o := range opts {
		o(c)
	}
	return c.asContainer()
}

// Builder provides utilities required to build a kubernetes container type
type ContainerBuilder struct {
	con    *container // container instance
	errors []error    // errors found while building the container instance
}

// NewBuilder returns a new instance of builder
func NewContainerBuilder() *ContainerBuilder {
	return &ContainerBuilder{
		con: &container{},
	}
}

// Build returns the final kubernetes container
func (b *ContainerBuilder) Build() (corev1.Container, error) {

	err := b.validate()
	if err != nil {
		return corev1.Container{}, err
	}
	return b.con.asContainer(), nil
}

func (b *ContainerBuilder) validate() error {
	if len(b.errors) != 0 {
		return errors.Errorf(
			"failed to validate: build errors were found: %+v",
			b.errors,
		)
	}
	return nil
}

// WithName sets the name of the container
func (b *ContainerBuilder) WithName(name string) *ContainerBuilder {
	if len(name) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing name"),
		)
		return b
	}
	WithName(name)(b.con)
	return b
}

// WithName sets the name of the container
func WithName(name string) OptionFunc {
	return func(c *container) {
		c.Name = name
	}
}

// WithImage sets the image of the container
func (b *ContainerBuilder) WithImage(img string) *ContainerBuilder {
	if len(img) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing image"),
		)
		return b
	}
	WithImage(img)(b.con)
	return b
}

// WithImage sets the image of the container
func WithImage(img string) OptionFunc {
	return func(c *container) {
		c.Image = img
	}
}

// WithCommandNew sets the command of the container
func (b *ContainerBuilder) WithCommandNew(cmd []string) *ContainerBuilder {
	if cmd == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil command"),
		)
		return b
	}

	if len(cmd) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing command"),
		)
		return b
	}

	newcmd := []string{}
	newcmd = append(newcmd, cmd...)

	b.con.Command = newcmd
	return b
}

// WithArgumentsNew sets the command arguments of the container
func (b *ContainerBuilder) WithArgumentsNew(args []string) *ContainerBuilder {
	if args == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil arguments"),
		)
		return b
	}

	if len(args) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing arguments"),
		)
		return b
	}

	newargs := []string{}
	newargs = append(newargs, args...)

	b.con.Args = newargs
	return b
}

// WithVolumeMountsNew sets the command arguments of the container
func (b *ContainerBuilder) WithVolumeMountsNew(volumeMounts []corev1.VolumeMount) *ContainerBuilder {
	if volumeMounts == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil volumemounts"),
		)
		return b
	}

	if len(volumeMounts) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing volumemounts"),
		)
		return b
	}
	newvolumeMounts := []corev1.VolumeMount{}
	newvolumeMounts = append(newvolumeMounts, volumeMounts...)
	b.con.VolumeMounts = newvolumeMounts
	return b
}

// WithVolumeDevices builds the containers with the appropriate volumeDevices
func (b *ContainerBuilder) WithVolumeDevices(volumeDevices []corev1.VolumeDevice) *ContainerBuilder {
	if volumeDevices == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil volumedevices"),
		)
		return b
	}
	if len(volumeDevices) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing volumedevices"),
		)
		return b
	}
	newVolumeDevices := []corev1.VolumeDevice{}
	newVolumeDevices = append(newVolumeDevices, volumeDevices...)
	b.con.VolumeDevices = newVolumeDevices
	return b
}

// WithImagePullPolicy sets the image pull policy of the container
func (b *ContainerBuilder) WithImagePullPolicy(policy corev1.PullPolicy) *ContainerBuilder {
	if len(policy) == 0 {
		b.errors = append(
			b.errors,
			errors.New(
				"failed to build container object: missing imagepullpolicy",
			),
		)
		return b
	}

	b.con.ImagePullPolicy = policy
	return b
}

// WithPrivilegedSecurityContext sets securitycontext of the container
func (b *ContainerBuilder) WithPrivilegedSecurityContext(privileged *bool) *ContainerBuilder {
	if privileged == nil {
		b.errors = append(
			b.errors,
			errors.New(
				"failed to build container object: missing securitycontext",
			),
		)
		return b
	}

	newprivileged := *privileged
	newsecuritycontext := &corev1.SecurityContext{
		Privileged: &newprivileged,
	}

	b.con.SecurityContext = newsecuritycontext
	return b
}

// WithResources sets resources of the container
func (b *ContainerBuilder) WithResources(
	resources *corev1.ResourceRequirements,
) *ContainerBuilder {
	if resources == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing resources"),
		)
		return b
	}

	newresources := *resources
	b.con.Resources = newresources
	return b
}

// WithResourcesByValue sets resources of the container
func (b *ContainerBuilder) WithResourcesByValue(resources corev1.ResourceRequirements) *ContainerBuilder {
	b.con.Resources = resources
	return b
}

// WithPortsNew sets ports of the container
func (b *ContainerBuilder) WithPortsNew(ports []corev1.ContainerPort) *ContainerBuilder {
	if ports == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil ports"),
		)
		return b
	}

	if len(ports) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing ports"),
		)
		return b
	}

	newports := []corev1.ContainerPort{}
	newports = append(newports, ports...)

	b.con.Ports = newports
	return b
}

// WithEnvsNew sets the envs of the container
func (b *ContainerBuilder) WithEnvsNew(envs []corev1.EnvVar) *ContainerBuilder {
	if envs == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil envs"),
		)
		return b
	}

	if len(envs) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing envs"),
		)
		return b
	}

	newenvs := []corev1.EnvVar{}
	newenvs = append(newenvs, envs...)

	b.con.Env = newenvs
	return b
}

// WithEnvs sets the envs of the container
func (b *ContainerBuilder) WithEnvs(envs []corev1.EnvVar) *ContainerBuilder {
	if envs == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil envs"),
		)
		return b
	}

	if len(envs) == 0 {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: missing envs"),
		)
		return b
	}

	if b.con.Env == nil {
		b.WithEnvsNew(envs)
		return b
	}

	b.con.Env = append(b.con.Env, envs...)
	return b
}

// WithLivenessProbe sets the liveness probe of the container
func (b *ContainerBuilder) WithLivenessProbe(liveness *corev1.Probe) *ContainerBuilder {
	if liveness == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil liveness probe"),
		)
		return b
	}

	b.con.LivenessProbe = liveness
	return b
}

// WithLifeCycle sets the life cycle of the container
func (b *ContainerBuilder) WithLifeCycle(lc *corev1.Lifecycle) *ContainerBuilder {
	if lc == nil {
		b.errors = append(
			b.errors,
			errors.New("failed to build container object: nil lifecycle"),
		)
		return b
	}

	b.con.Lifecycle = lc
	return b
}
