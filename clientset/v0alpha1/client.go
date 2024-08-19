package v0alpha1

import (
	"github.com/brunostjohn/k8s-mc-loadbalancer/api/types/v0alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type MCIngressV0Alpha1Interface interface {
    Ingresses(namespace string) MCIngressInterface 
}

type MCIngressV0Alpha1Client struct {
    restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*MCIngressV0Alpha1Client, error) {
    config := *c
    config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v0alpha1.GroupName, Version: v0alpha1.GroupVersion}
    config.APIPath = "/apis"
    config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
    config.UserAgent = rest.DefaultKubernetesUserAgent()

    client, err := rest.RESTClientFor(&config)
    if err != nil {
        return nil, err
    }

    return &MCIngressV0Alpha1Client{restClient: client}, nil
}

func (c *MCIngressV0Alpha1Client) Ingresses() MCIngressInterface {
    return &mcIngressClient{
        restClient: c.restClient,
    }
}