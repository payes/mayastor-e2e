package v1alpha1

import (
	"context"
	"time"

	"mayastor-e2e/common"
	v1alpha12 "mayastor-e2e/common/custom_resources/api/types/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// MayastorVolumeInterface has methods to work with Mayastor volume resources.
type MayastorVolumeInterface interface {
	//	Create(ctxt context.Context, mayastorvolume *v1alpha12.MayastorVolume, opts metav1.CreateOptions) (*v1alpha12.MayastorVolume, error)
	Get(ctxt context.Context, name string, opts metav1.GetOptions) (*v1alpha12.MayastorVolume, error)
	List(ctxt context.Context, opts metav1.ListOptions) (*v1alpha12.MayastorVolumeList, error)
	Update(ctxt context.Context, mayastorvolume *v1alpha12.MayastorVolume, opts metav1.UpdateOptions) (*v1alpha12.MayastorVolume, error)
	Delete(ctxt context.Context, name string, opts metav1.DeleteOptions) error
	Watch(ctxt context.Context, opts metav1.ListOptions) (watch.Interface, error)
	// ...
}

// msvClient implements MayastorVolumeInterface
type msvClient struct {
	restClient rest.Interface
}

/*
// Create takes the representation of a Mayastor volume and creates it. Returns the server's representation of the volume, and an error if one occurred.
func (c *msvClient) Create(ctxt context.Context, mayastorvolume *v1alpha12.MayastorVolume, opts metav1.CreateOptions) (*v1alpha12.MayastorVolume, error) {
	result := v1alpha12.MayastorVolume{}
	err := c.restClient.
		Post().
		Namespace(common.NSMayastor()).
		Resource(common.CRDVolumesResourceName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mayastorvolume).
		Do(ctxt).
		Into(&result)

	return &result, err
}
*/

// Get takes the name of the Mayastor volume and returns the server's representation of it, and an error if one occurred.
func (c *msvClient) Get(ctxt context.Context, name string, opts metav1.GetOptions) (*v1alpha12.MayastorVolume, error) {
	result := v1alpha12.MayastorVolume{}
	err := c.restClient.
		Get().
		Namespace(common.NSMayastor()).
		Resource(common.CRDVolumesResourceName).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctxt).
		Into(&result)

	return &result, err
}

// List takes the label and field selectors, and returns a list matching Mayastor Volume, and an error if one occurred.
func (c *msvClient) List(ctxt context.Context, opts metav1.ListOptions) (*v1alpha12.MayastorVolumeList, error) {
	result := v1alpha12.MayastorVolumeList{}
	err := c.restClient.
		Get().
		Namespace(common.NSMayastor()).
		Resource(common.CRDVolumesResourceName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctxt).
		Into(&result)

	return &result, err
}

// Update takes the representation of a Mayastor volume and updates it. Returns the server's representation of the volume, and an error if one occurred.
func (c *msvClient) Update(ctxt context.Context, mayastorvolume *v1alpha12.MayastorVolume, opts metav1.UpdateOptions) (*v1alpha12.MayastorVolume, error) {
	result := v1alpha12.MayastorVolume{}
	err := c.restClient.
		Put().
		Namespace(common.NSMayastor()).
		Resource(common.CRDVolumesResourceName).
		Name(mayastorvolume.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mayastorvolume).
		Do(ctxt).
		Into(&result)

	return &result, err
}

// Delete takes the name of the Mayastor volume and deletes it. Returns error if one occurred.
func (c *msvClient) Delete(ctxt context.Context, name string, opts metav1.DeleteOptions) error {
	return c.restClient.
		Delete().
		Namespace(common.NSMayastor()).
		Resource(common.CRDVolumesResourceName).
		Name(name).
		Body(&opts).
		Do(ctxt).
		Error()
}

// Watch takes the label and field selectors, and returns a watch.Interface the watches matching Mayastor volumes, and an error if one occurred.
func (c *msvClient) Watch(ctxt context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(common.NSMayastor()).
		Resource(common.CRDVolumesResourceName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctxt)
}
