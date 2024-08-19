package loadbalancer

import (
	"context"
	"time"

	"github.com/brunostjohn/k8s-mc-loadbalancer/api/types/v0alpha1"
	client_v0alpha1 "github.com/brunostjohn/k8s-mc-loadbalancer/clientset/v0alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func watchWithChannel(ctx context.Context, watchFrequency time.Duration, clientSet client_v0alpha1.MCIngressV0Alpha1Client) chan []v0alpha1.MCIngress {
	ingressChan := make(chan []v0alpha1.MCIngress)
	ingressStore := watchResources(ctx, clientSet)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ingressListGeneric := ingressStore.List()
				ingressList := make([]v0alpha1.MCIngress, len(ingressListGeneric))

				for i, ingress := range ingressListGeneric {
					ingressList[i] = *ingress.(*v0alpha1.MCIngress)
				}

				ingressChan <- ingressList

				time.Sleep(watchFrequency)
			}
		}
	}()

	return ingressChan
}

func watchResources(ctx context.Context, clientSet client_v0alpha1.MCIngressV0Alpha1Client) cache.Store {
	ingressStore, ingressController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return clientSet.Ingresses().List(ctx, lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return clientSet.Ingresses().Watch(ctx, lo)
			},
		},
		&v0alpha1.MCIngress{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{},
	)

	go ingressController.Run(ctx.Done())
	return ingressStore
}
