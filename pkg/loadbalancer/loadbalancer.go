package loadbalancer

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/brunostjohn/k8s-mc-loadbalancer/api/types/v0alpha1"
	client_v0alpha1 "github.com/brunostjohn/k8s-mc-loadbalancer/clientset/v0alpha1"
	ir "github.com/haveachin/infrared"
	"github.com/rs/zerolog"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type MCLoadBalancerOptions struct {
	BindAddress                    *net.IP
	BindPort                       int
	ProxyProtocolReceive           bool
	ProxyProtocolTrustedCIDRs      []string
	KeepAliveTimeout               int
	FiltersRateLimiterRequestLimit int
	FiltersRateLimiterWindowLength int
	Kubeconfig string
	Log *zerolog.Logger
	WatchFrequency time.Duration
	EnablePrometheus bool
	PrometheusBindAddress *net.IP
	PrometheusBindPort int
}

type ProxyUUIDs struct {
	ingress v0alpha1.MCIngress
	uuids []*ir.Proxy
}

type MCLoadBalancer struct {
	options MCLoadBalancerOptions
	client  *client_v0alpha1.MCIngressV0Alpha1Client
	gateway *ir.Gateway
	uuids  *[]ProxyUUIDs
	log *zerolog.Logger
	bindAddress string
	bindPort string
	watchFrequency time.Duration
}

func NewMCLoadBalancer(ctx context.Context, options MCLoadBalancerOptions) (*MCLoadBalancer, error) {
	var config *rest.Config
    var err error

	if options.Kubeconfig == "" {
        options.Log.Info().Msg("using in-cluster configuration")
        config, err = rest.InClusterConfig()
    } else {
        options.Log.Info().Msgf("using configuration from file %s", options.Kubeconfig)
        config, err = clientcmd.BuildConfigFromFlags("", options.Kubeconfig)
    }

	if err != nil {
		options.Log.Error().Err(err).Msg("error getting config")
		return nil, err
	}

	v0alpha1.AddToScheme(scheme.Scheme)

	client, err := client_v0alpha1.NewForConfig(config)
	if err != nil {
		options.Log.Error().Err(err).Msg("error creating client")
		return nil, err
	}

	gateway := ir.Gateway{
		ReceiveProxyProtocol: options.ProxyProtocolReceive,
	}

	if options.EnablePrometheus {
		addr := options.PrometheusBindAddress.String() + ":" + fmt.Sprint(options.PrometheusBindPort)
		err := gateway.EnablePrometheus(addr)
		if err != nil {
			options.Log.Error().Err(err).Msg("Error enabling prometheus")
		}
	}

	var bindAddress string
	if options.BindAddress != nil {
		bindAddress = options.BindAddress.String()
	} else {
		bindAddress = "0.0.0.0"
	}

	var bindPort string
	if options.BindPort != 0 {
		bindPort = fmt.Sprint(options.BindPort)
	} else {
		bindPort = "25565"
	}

	return &MCLoadBalancer{
		client: client,
		options: options,
		gateway: &gateway,
		log: options.Log,
		bindAddress: bindAddress,
		bindPort: bindPort,
		uuids: &[]ProxyUUIDs{},
		watchFrequency: options.WatchFrequency,
	}, nil
}

func (lb *MCLoadBalancer) Start(ctx context.Context) error {
	lb.log.Info().Msg("Starting load balancer")

	proxiesChannel := watchWithChannel(ctx, lb.watchFrequency, *lb.client)

	list, err := lb.client.Ingresses().List(ctx, v1.ListOptions{})
	if err != nil {
		lb.log.Error().Err(err).Msg("Error listing ingresses")
		return err
	}
	lb.log.Debug().Msgf("List of ingresses: %v", ingressListToString(&list.Items))

	lb.uuids = convertLbListToProxy(&list.Items, lb.bindAddress, lb.bindPort)

	go lb.gatewayListen(ctx)
	go lb.handleProxyUpdates(ctx, proxiesChannel)

	<-ctx.Done()
	
	return nil
}

func (lb *MCLoadBalancer) handleProxyUpdates(ctx context.Context, proxiesChannel chan []v0alpha1.MCIngress) {
	for {
		select {
		case <-ctx.Done():
			return
		case proxies := <-proxiesChannel:
			missing, missingProxies := missingListElements(*lb.uuids, proxies)
			for _, proxy := range missing {
				lb.log.Info().Msgf("Removing deleted proxy: %s", proxy)
				lb.gateway.CloseProxy(proxy)
			}
			proxiesAfterRemoval := make([]ProxyUUIDs, 0)
			for _, proxyUUID := range *lb.uuids {
				found := false
				for _, proxy := range missingProxies {
					if proxy.ingress.ObjectMeta.Name == proxyUUID.ingress.ObjectMeta.Name {
						found = true
						break
					}
				}

				if !found {
					proxiesAfterRemoval = append(proxiesAfterRemoval, proxyUUID)
				}
			}
			lb.uuids = &proxiesAfterRemoval

			newProxies := newListEleemnts(*lb.uuids, proxies)
			lb.log.Debug().Msgf("Found %v new proxies", len(newProxies))
			newList := convertLbListToProxy(&newProxies, lb.bindAddress, lb.bindPort)
			for _, proxyUUID := range *newList {
				for _, proxy := range proxyUUID.uuids {
					lb.log.Info().Msgf("Registering new proxy: %s", proxy.Config.DomainName)
					err := lb.gateway.RegisterProxy(proxy)
					if err != nil {
						lb.log.Error().Err(err).Msg("Error registering proxy")
					}
				}
			}
			newUUIDs := append(*lb.uuids, *newList...)
			lb.uuids = &newUUIDs
		}
	}
}

func newListEleemnts(old []ProxyUUIDs, new []v0alpha1.MCIngress) []v0alpha1.MCIngress {
	newList := make([]v0alpha1.MCIngress, 0)
	for _, proxy := range new {
		found := false
		for _, proxyUUID := range old {
			if proxy.ObjectMeta.Name == proxyUUID.ingress.ObjectMeta.Name {
				found = true
				break
			}
		}

		if !found {
			newList = append(newList, proxy)
		}
	}

	return newList
}

func missingListElements(old []ProxyUUIDs, new []v0alpha1.MCIngress) ([]string, []ProxyUUIDs) {
	missing := make([]string, 0)
	missingProxies := make([]ProxyUUIDs, 0)

	for _, proxyUUID := range old {
		found := false
		for _, proxy := range new {
			if proxyUUID.ingress.ObjectMeta.Name == proxy.ObjectMeta.Name {
				found = true
				break
			}
		}

		if !found {
			missingProxies = append(missingProxies, proxyUUID)
			for _, proxy := range proxyUUID.uuids {
				missing = append(missing, proxy.UID())
			}
		}
	}

	return missing, missingProxies
}

func (lb *MCLoadBalancer) gatewayListen(ctx context.Context) {
	flattened := flattenProxyUUIDs(lb.uuids)

	go lb.gateway.ListenAndServe(flattened)

	<-ctx.Done()

	lb.gateway.Close()
}