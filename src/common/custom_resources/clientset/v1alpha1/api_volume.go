package v1alpha1

import (
	"mayastor-e2e/common/e2e_config"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// Volume API

type MayastorVolumeV1Alpha1Interface interface {
	MayastorVolumes() MayastorVolumeInterface
}

type MayastorVolumeV1Alpha1Client struct {
	restClient rest.Interface
}

func MsvNewForConfig(c *rest.Config) (*MayastorVolumeV1Alpha1Client, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: e2e_config.GetConfig().Product.CrdGroupName,
		Version: e2e_config.GetConfig().Product.CrdVolumeGroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &MayastorVolumeV1Alpha1Client{restClient: client}, nil
}

func (c *MayastorVolumeV1Alpha1Client) MayastorVolumes() MayastorVolumeInterface {
	return &msvClient{
		restClient: c.restClient,
	}
}
