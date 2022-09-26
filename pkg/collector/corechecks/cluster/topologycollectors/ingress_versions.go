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
	GetCreationTimestamp() metav1.Time
	GetGenerateName() string
	GetIngressPoints() []string
	GetKind() string
	GetKubernetesObject() MarshalableKubernetesObject
	GetName() string
	GetNamespace() string
	GetObjectMeta() metav1.ObjectMeta
	GetServiceName() string
	GetServiceNames() []string
	GetString() string
	GetUID() types.UID
}

type IngressV1B1 struct {
	o v1beta1.Ingress
	MarshalableKubernetesObject
}

type IngressNetV1 struct {
	o netv1.Ingress
	MarshalableKubernetesObject
}

/* IngressV1B1 */

func (in IngressV1B1) GetServiceName() string {
	if in.o.Spec.Backend != nil && in.o.Spec.Backend.ServiceName != "" {
		return in.o.Spec.Backend.ServiceName
	}
	return ""
}

func (in IngressV1B1) GetString() string {
	return in.o.String()
}

func (in IngressV1B1) GetCreationTimestamp() metav1.Time {
	return in.o.CreationTimestamp
}

func (in IngressV1B1) GetUID() types.UID {
	return in.o.UID
}

func (in IngressV1B1) GetGenerateName() string {
	return in.o.GenerateName
}

func (in IngressV1B1) GetObjectMeta() metav1.ObjectMeta {
	return in.o.ObjectMeta
}

func (in IngressV1B1) GetKubernetesObject() MarshalableKubernetesObject {
	return &in.o
}

func (in IngressV1B1) GetKind() string {
	return in.o.Kind
}

func (in IngressV1B1) GetName() string {
	return in.o.Name
}

func (in IngressV1B1) GetNamespace() string {
	return in.o.Namespace
}

func (in IngressV1B1) GetServiceNames() []string {
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

func (in IngressV1B1) GetIngressPoints() []string {
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

/* IngressNetV1 */

func (in IngressNetV1) GetServiceName() string {
	if in.o.Spec.DefaultBackend != nil && in.o.Spec.DefaultBackend.Service.Name != "" {
		return in.o.Spec.DefaultBackend.Service.Name
	}
	return ""
}

func (in IngressNetV1) GetString() string {
	return in.o.String()
}

func (in IngressNetV1) GetCreationTimestamp() metav1.Time {
	return in.o.CreationTimestamp
}

func (in IngressNetV1) GetUID() types.UID {
	return in.o.UID
}

func (in IngressNetV1) GetGenerateName() string {
	return in.o.GenerateName
}

func (in IngressNetV1) GetObjectMeta() metav1.ObjectMeta {
	return in.o.ObjectMeta
}

func (in IngressNetV1) GetKubernetesObject() MarshalableKubernetesObject {
	return &in.o
}

func (in IngressNetV1) GetKind() string {
	return in.o.Kind
}

func (in IngressNetV1) GetName() string {
	return in.o.Name
}

func (in IngressNetV1) GetNamespace() string {
	return in.o.Namespace
}

func (in IngressNetV1) GetServiceNames() []string {
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

func (in IngressNetV1) GetIngressPoints() []string {
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
