package v1alpha1

import (
	"mayastor-e2e/tools/extended-test-framework/common"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// Pool API

type MayastorPoolV1Alpha1Interface interface {
	MayastorPools() MayastorPoolInterface
}

type MayastorPoolV1Alpha1Client struct {
	restClient rest.Interface
}

func MspNewForConfig(c *rest.Config) (*MayastorPoolV1Alpha1Client, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: common.CRDGroupName, Version: common.CRDPoolGroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &MayastorPoolV1Alpha1Client{restClient: client}, nil
}

func (c *MayastorPoolV1Alpha1Client) MayastorPools() MayastorPoolInterface {
	return &mspClient{
		restClient: c.restClient,
	}
}
