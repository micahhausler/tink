package api

import (
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	"github.com/tinkerbell/tink/k8s/api/v1alpha1"
)

func init() {
	v1alpha1.AddToScheme(scheme.Scheme)
}

type TinkerbellV1Alpha1Interface interface {
	RESTClient() rest.Interface
	Hardware() HardwareInterface
	Workflow() WorkflowInterface
	Template() TemplateInterface
}

type TinkerbellV1Alpha1Client struct {
	restClient rest.Interface
}

// compile time check
var _ TinkerbellV1Alpha1Interface = &TinkerbellV1Alpha1Client{}

func NewForConfig(c *rest.Config) (*TinkerbellV1Alpha1Client, error) {
	config := *c
	config.ContentConfig.GroupVersion = &v1alpha1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &TinkerbellV1Alpha1Client{restClient: client}, nil
}

func (c *TinkerbellV1Alpha1Client) Workflow() WorkflowInterface {
	return &workflowClient{
		restClient: c.restClient,
	}
}

func (c *TinkerbellV1Alpha1Client) Hardware() HardwareInterface {
	return &hardwareClient{
		restClient: c.restClient,
	}
}

func (c *TinkerbellV1Alpha1Client) Template() TemplateInterface {
	return &templateClient{
		restClient: c.restClient,
	}
}

func (c *TinkerbellV1Alpha1Client) RESTClient() rest.Interface {
	return c.restClient
}
