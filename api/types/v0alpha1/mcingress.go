package v0alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type MCIngressSpec struct {
	Service string `json:"service"`
	Port int `json:"port"`
	Hosts []string `json:"hosts"`
	SendProxyProtocol bool `json:"sendProxyProtocol"`
}

type MCIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec MCIngressSpec `json:"spec"`
}

type MCIngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MCIngress `json:"items"`
}