package v1alpha1

import (
	"mayastor-e2e/tools/extended-test-framework/common"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// Node API

type MayastorNodeV1Alpha1Interface interface {
	MayastorNodes() MayastorNodeInterface
}

type MayastorNodeV1Alpha1Client struct {
	restClient rest.Interface
}

func MsnNewForConfig(c *rest.Config) (*MayastorNodeV1Alpha1Client, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: common.CRDGroupName, Version: common.CRDNodeGroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &MayastorNodeV1Alpha1Client{restClient: client}, nil
}

func (c *MayastorNodeV1Alpha1Client) MayastorNodes() MayastorNodeInterface {
	return &msnClient{
		restClient: c.restClient,
	}
}
