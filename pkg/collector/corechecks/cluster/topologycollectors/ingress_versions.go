//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"k8s.io/api/extensions/v1beta1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type IngressInterface interface {
	GetServiceName() string
	String() string
	GetCreationTimestamp() metav1.Time
	GetUID() types.UID
	GetGenerateName() string
	GetKind() string
	GetKubernetesObject() MarshalableKubernetesObject
	GetObjectMeta() metav1.ObjectMeta
	GetName() string
	GetNamespace() string
	GetServiceNames() []string
	GetIngressPoints() []string
}

type V1Beta1Ingress struct {
	o v1beta1.Ingress
	MarshalableKubernetesObject
}

type NetV1Ingress struct {
	o netv1.Ingress
	MarshalableKubernetesObject
}

/* V1Beta1Ingress */

func (in V1Beta1Ingress) GetServiceName() string {
	if in.o.Spec.Backend != nil && in.o.Spec.Backend.ServiceName != "" {
		return in.o.Spec.Backend.ServiceName
	}
	return ""
}

func (in V1Beta1Ingress) String() string {
	return in.o.String()
}

func (in V1Beta1Ingress) GetCreationTimestamp() metav1.Time {
	return in.o.CreationTimestamp
}

func (in V1Beta1Ingress) GetUID() types.UID {
	return in.o.UID
}

func (in V1Beta1Ingress) GetGenerateName() string {
	return in.o.GenerateName
}

func (in V1Beta1Ingress) GetObjectMeta() metav1.ObjectMeta {
	return in.o.ObjectMeta
}

func (in V1Beta1Ingress) GetKubernetesObject() MarshalableKubernetesObject {
	return &in.o
}

func (in V1Beta1Ingress) GetKind() string {
	return in.o.Kind
}

func (in V1Beta1Ingress) GetName() string {
	return in.o.Name
}

func (in V1Beta1Ingress) GetNamespace() string {
	return in.o.Namespace
}

func (in V1Beta1Ingress) GetServiceNames() []string {
	var result []string
	for _, rules := range in.o.Spec.Rules {
		if rules.HTTP != nil {
			for _, path := range rules.HTTP.Paths {
				result = append(result, path.Backend.ServiceName)
			}
		}
	}
	return result
}

func (in V1Beta1Ingress) GetIngressPoints() []string {
	var result []string
	lbIngresses := in.o.Status.LoadBalancer.Ingress
	for _, ingressPoints := range lbIngresses {
		if ingressPoints.Hostname != "" {
			result = append(result, ingressPoints.Hostname)
		}
		if ingressPoints.IP != "" {
			result = append(result, ingressPoints.IP)
		}
	}
	return result
}

/* NetV1Ingress */

func (in NetV1Ingress) GetServiceName() string {
	if in.o.Spec.DefaultBackend != nil && in.o.Spec.DefaultBackend.Service.Name != "" {
		return in.o.Spec.DefaultBackend.Service.Name
	}
	return ""
}

func (in NetV1Ingress) String() string {
	return in.o.String()
}

func (in NetV1Ingress) GetCreationTimestamp() metav1.Time {
	return in.o.CreationTimestamp
}

func (in NetV1Ingress) GetUID() types.UID {
	return in.o.UID
}

func (in NetV1Ingress) GetGenerateName() string {
	return in.o.GenerateName
}

func (in NetV1Ingress) GetObjectMeta() metav1.ObjectMeta {
	return in.o.ObjectMeta
}

func (in NetV1Ingress) GetKubernetesObject() MarshalableKubernetesObject {
	return &in.o
}

func (in NetV1Ingress) GetKind() string {
	return in.o.Kind
}

func (in NetV1Ingress) GetName() string {
	return in.o.Name
}

func (in NetV1Ingress) GetNamespace() string {
	return in.o.Namespace
}

func (in NetV1Ingress) GetServiceNames() []string {
	var result []string
	for _, rules := range in.o.Spec.Rules {
		if rules.HTTP != nil {
			for _, path := range rules.HTTP.Paths {
				result = append(result, path.Backend.Service.Name)
			}
		}
	}
	return result
}

func (in NetV1Ingress) GetIngressPoints() []string {
	var result []string
	for _, ingressPoints := range in.o.Status.LoadBalancer.Ingress {
		if ingressPoints.Hostname != "" {
			result = append(result, ingressPoints.Hostname)
		}
		if ingressPoints.IP != "" {
			result = append(result, ingressPoints.IP)
		}
	}
	return result
}
