package v0alpha1

import (
	"context"

	"github.com/brunostjohn/k8s-mc-loadbalancer/api/types/v0alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type MCIngressInterface interface {
    List(ctx context.Context, opts metav1.ListOptions) (*v0alpha1.MCIngressList, error)
    Get(ctx context.Context, name string, options metav1.GetOptions) (*v0alpha1.MCIngress, error)
    Create(ctx context.Context, ingress *v0alpha1.MCIngress) (*v0alpha1.MCIngress, error)
    Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type mcIngressClient struct {
    restClient rest.Interface
}

func (c *mcIngressClient) List(ctx context.Context, opts metav1.ListOptions) (*v0alpha1.MCIngressList, error) {
    result := v0alpha1.MCIngressList{}
    err := c.restClient.
        Get().
        Resource("mc-ingresses").
        VersionedParams(&opts, scheme.ParameterCodec).
        Do(ctx).
        Into(&result)

    return &result, err
}

func (c *mcIngressClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v0alpha1.MCIngress, error) {
    result := v0alpha1.MCIngress{}
    err := c.restClient.
        Get().
        Resource("mc-ingresses").
        Name(name).
        VersionedParams(&opts, scheme.ParameterCodec).
        Do(ctx).
        Into(&result)

    return &result, err
}

func (c *mcIngressClient) Create(ctx context.Context, ingress *v0alpha1.MCIngress) (*v0alpha1.MCIngress, error) {
    result := v0alpha1.MCIngress{}
    err := c.restClient.
        Post().
        Resource("mc-ingresses").
        Body(ingress).
        Do(ctx).
        Into(&result)

    return &result, err
}

func (c *mcIngressClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
    opts.Watch = true
    return c.restClient.
        Get().
        Resource("mc-ingresses").
        VersionedParams(&opts, scheme.ParameterCodec).
        Watch(ctx)
}