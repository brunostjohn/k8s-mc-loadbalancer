package v0alpha1

import "k8s.io/apimachinery/pkg/runtime"

func (in *MCIngress) DeepCopyInto(out *MCIngress) {
    out.TypeMeta = in.TypeMeta
    out.ObjectMeta = in.ObjectMeta
    out.Spec = MCIngressSpec{
        Service: in.Spec.Service,
		Port: in.Spec.Port,
		Hosts: in.Spec.Hosts,
		SendProxyProtocol: in.Spec.SendProxyProtocol,
    }
}

func (in *MCIngress) DeepCopyObject() runtime.Object {
    out := MCIngress{}
    in.DeepCopyInto(&out)

    return &out
}

func (in *MCIngressList) DeepCopyObject() runtime.Object {
    out := MCIngressList{}
    out.TypeMeta = in.TypeMeta
    out.ListMeta = in.ListMeta

    if in.Items != nil {
        out.Items = make([]MCIngress, len(in.Items))
        for i := range in.Items {
            in.Items[i].DeepCopyInto(&out.Items[i])
        }
    }

    return &out
}