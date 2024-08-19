package loadbalancer

import (
	"encoding/json"

	types "github.com/brunostjohn/k8s-mc-loadbalancer/api/types/v0alpha1"
	ir "github.com/haveachin/infrared"
)

func ingressListToString(balancers *[]types.MCIngress) string {
	str := ""
	for _, lb := range *balancers {
		newStr, err := json.MarshalIndent(lb, "", "  ")
		if err != nil {
			return ""
		}
		str += string(newStr) + "\n"
	}

	return str
}

func flattenProxyUUIDs(proxies *[]ProxyUUIDs) []*ir.Proxy {
	flatProxies := make([]*ir.Proxy, 0)
	for _, proxyUUID := range *proxies {
		flatProxies = append(flatProxies, proxyUUID.uuids...)
	}

	return flatProxies
}

func convertLbListToProxy(lbs *[]types.MCIngress, bind string, bindPort string, clusterDomain string) *[]ProxyUUIDs {
	proxies := make([]ProxyUUIDs, len(*lbs))
	for i, lb := range *lbs {
		proxies[i] = *convertLbToProxy(&lb, bind, bindPort, clusterDomain)
	}

	return &proxies
}

func convertLbToProxy(lb *types.MCIngress, bind string, bindPort string, clusterDomain string) *ProxyUUIDs {
	proxies := make([]*ir.Proxy, len(lb.Spec.Hosts))
	for i, host := range lb.Spec.Hosts {
		proxies[i] = &ir.Proxy{
			Config: &ir.ProxyConfig{
				DomainName: host,
				ProxyTo: lb.Spec.Service+"."+lb.ObjectMeta.Namespace+".svc."+clusterDomain,
				ProxyBind: bind,
				ListenTo: ":"+bindPort,
			},
		}
	}

	return &ProxyUUIDs{
		ingress: *lb,
		uuids: proxies,
	}
}
