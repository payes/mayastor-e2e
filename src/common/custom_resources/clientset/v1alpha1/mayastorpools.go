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

// MayastorPoolInterface has methods to work with Mayastor pool resources.
type MayastorPoolInterface interface {
	Create(ctxt context.Context, mayastorpool *v1alpha12.MayastorPool, opts metav1.CreateOptions) (*v1alpha12.MayastorPool, error)
	Get(ctxt context.Context, name string, opts metav1.GetOptions) (*v1alpha12.MayastorPool, error)
	List(ctxt context.Context, opts metav1.ListOptions) (*v1alpha12.MayastorPoolList, error)
	Update(ctxt context.Context, mayastorpool *v1alpha12.MayastorPool, opts metav1.UpdateOptions) (*v1alpha12.MayastorPool, error)
	Delete(ctxt context.Context, name string, opts metav1.DeleteOptions) error
	Watch(ctxt context.Context, opts metav1.ListOptions) (watch.Interface, error)
	// ...
}

// mspClient implements MayastorPoolInterface
type mspClient struct {
	restClient rest.Interface
}

// Create takes the representation of a Mayastor pool and creates it. Returns the server's representation of the pool, and an error if one occurred.
func (c *mspClient) Create(ctxt context.Context, mayastorpool *v1alpha12.MayastorPool, opts metav1.CreateOptions) (*v1alpha12.MayastorPool, error) {
	result := v1alpha12.MayastorPool{}
	err := c.restClient.
		Post().
		Namespace(common.NSMayastor()).
		Resource(common.CRDPoolsResourceName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mayastorpool).
		Do(ctxt).
		Into(&result)

	return &result, err
}

// Get takes the name of the Mayastor pool and returns the server's representation of it, and an error if one occurred.
func (c *mspClient) Get(ctxt context.Context, name string, opts metav1.GetOptions) (*v1alpha12.MayastorPool, error) {
	result := v1alpha12.MayastorPool{}
	err := c.restClient.
		Get().
		Namespace(common.NSMayastor()).
		Resource(common.CRDPoolsResourceName).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctxt).
		Into(&result)

	return &result, err
}

// List takes the label and field selectors, and returns a list matching Mayastor Pool, and an error if one occurred.
func (c *mspClient) List(ctxt context.Context, opts metav1.ListOptions) (*v1alpha12.MayastorPoolList, error) {
	result := v1alpha12.MayastorPoolList{}
	err := c.restClient.
		Get().
		Namespace(common.NSMayastor()).
		Resource(common.CRDPoolsResourceName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctxt).
		Into(&result)

	return &result, err
}

// Update takes the representation of a Mayastor pool and updates it. Returns the server's representation of the pool, and an error if one occurred.
func (c *mspClient) Update(ctxt context.Context, mayastorpool *v1alpha12.MayastorPool, opts metav1.UpdateOptions) (*v1alpha12.MayastorPool, error) {
	result := v1alpha12.MayastorPool{}
	err := c.restClient.
		Put().
		Namespace(common.NSMayastor()).
		Resource(common.CRDPoolsResourceName).
		Name(mayastorpool.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(mayastorpool).
		Do(ctxt).
		Into(&result)

	return &result, err
}

// Delete takes the name of the Mayastor pool and deletes it. Returns error if one occurred.
func (c *mspClient) Delete(ctxt context.Context, name string, opts metav1.DeleteOptions) error {
	return c.restClient.
		Delete().
		Namespace(common.NSMayastor()).
		Resource(common.CRDPoolsResourceName).
		Name(name).
		Body(&opts).
		Do(ctxt).
		Error()
}

// Watch takes the label and field selectors, and returns a watch.Interface the watches matching Mayastor pools, and an error if one occurred.
func (c *mspClient) Watch(ctxt context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(common.NSMayastor()).
		Resource(common.CRDPoolsResourceName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctxt)
}
