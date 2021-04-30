package v1alpha1

import (
	"context"
	"time"

	"mayastor-e2e/common"
	v1alpha12 "mayastor-e2e/common/crds/api/types/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// MayastorNodeInterface has methods to work with Mayastor node resources.
type MayastorNodeInterface interface {
	Get(ctxt context.Context, name string, opts metav1.GetOptions) (*v1alpha12.MayastorNode, error)
	List(ctxt context.Context, opts metav1.ListOptions) (*v1alpha12.MayastorNodeList, error)
	Delete(ctxt context.Context, name string, opts metav1.DeleteOptions) error
	Watch(ctxt context.Context, opts metav1.ListOptions) (watch.Interface, error)
	// ...
}

// msnClient implements MayastorNodeInterface
type msnClient struct {
	restClient rest.Interface
}

// Get takes the name of the Mayastor node and returns the server's representation of it, and an error if one occurred.
func (c *msnClient) Get(ctxt context.Context, name string, opts metav1.GetOptions) (*v1alpha12.MayastorNode, error) {
	result := v1alpha12.MayastorNode{}
	err := c.restClient.
		Get().
		Namespace(common.NSMayastor).
		Resource(common.CRDNodesResourceName).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctxt).
		Into(&result)

	return &result, err
}

// List takes the label and field selectors, and returns a list matching Mayastor Node, and an error if one occurred.
func (c *msnClient) List(ctxt context.Context, opts metav1.ListOptions) (*v1alpha12.MayastorNodeList, error) {
	result := v1alpha12.MayastorNodeList{}
	err := c.restClient.
		Get().
		Namespace(common.NSMayastor).
		Resource(common.CRDNodesResourceName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctxt).
		Into(&result)

	return &result, err
}

// Delete takes the name of the Mayastor node and deletes it. Returns error if one occurred.
func (c *msnClient) Delete(ctxt context.Context, name string, opts metav1.DeleteOptions) error {
	return c.restClient.
		Delete().
		Namespace(common.NSMayastor).
		Resource(common.CRDNodesResourceName).
		Name(name).
		Body(&opts).
		Do(ctxt).
		Error()
}

// Watch takes the label and field selectors, and returns a watch.Interface the watches matching Mayastor nodes, and an error if one occurred.
func (c *msnClient) Watch(ctxt context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(common.NSMayastor).
		Resource(common.CRDNodesResourceName).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctxt)
}
