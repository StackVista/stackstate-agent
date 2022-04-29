//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"errors"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type StsIngress struct {
	v1beta1 *v1beta1.Ingress
	netv1   *netv1.Ingress
	MarshalableKubernetesObject
}

// IngressCollector implements the ClusterTopologyCollector interface.
type IngressCollector struct {
	ComponentChan chan<- *topology.Component
	RelationChan  chan<- *topology.Relation
	ClusterTopologyCollector
}

// NewIngressCollector
func NewIngressCollector(componentChannel chan<- *topology.Component, relationChannel chan<- *topology.Relation,
	clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &IngressCollector{
		ComponentChan:            componentChannel,
		RelationChan:             relationChannel,
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*IngressCollector) GetName() string {
	return "Ingress Collector"
}

// Collects and Published the Ingress Components
func (ic *IngressCollector) CollectorFunction() error {
	var ingresses []StsIngress
	ingresses, err := getExtV1Ingresses(ic, ingresses)
	if err != nil {
		return err
	}
	ingresses, err = getNetV1Ingresses(ic, ingresses)
	if err != nil {
		return err
	}

	for _, in := range ingresses {
		component := ic.ingressToStackStateComponent(in)
		ic.ComponentChan <- component
		// submit relation to service name for correlation
		if in.GetServiceName() != "" {
			serviceExternalID := ic.buildServiceExternalID(in.GetNamespace(), in.GetServiceName())

			// publish the ingress -> service relation
			relation := ic.ingressToServiceStackStateRelation(component.ExternalID, serviceExternalID)
			ic.RelationChan <- relation
		}

		// submit relation to service name in the ingress rules for correlation
		for _, serviceName := range in.GetServiceNames() {
			serviceExternalID := ic.buildServiceExternalID(in.GetNamespace(), serviceName)

			// publish the ingress -> service relation
			relation := ic.ingressToServiceStackStateRelation(component.ExternalID, serviceExternalID)
			ic.RelationChan <- relation
		}

		// submit relation to loadbalancer

		for _, ingressPoint := range in.GetIngressPoints() {
			endpoint := ic.endpointStackStateComponentFromIngress(in, ingressPoint)

			ic.ComponentChan <- endpoint
			ic.RelationChan <- ic.endpointToIngressStackStateRelation(endpoint.ExternalID, component.ExternalID)
		}
	}

	return nil
}

func getExtV1Ingresses(ic *IngressCollector, ingresses []StsIngress) ([]StsIngress, error) {
	ingressesExt, err := ic.GetAPIClient().GetIngressesExtV1()
	if err != nil {
		return nil, err
	}
	for _, in := range ingressesExt {
		ingresses = append(ingresses, StsIngress{
			v1beta1: &in,
		})
	}
	return ingresses, nil
}

func getNetV1Ingresses(ic *IngressCollector, ingresses []StsIngress) ([]StsIngress, error) {
	ingressesExt, err := ic.GetAPIClient().GetIngressesNetV1()
	if err != nil {
		return nil, err
	}
	for _, in := range ingressesExt {
		ingresses = append(ingresses, StsIngress{
			netv1: &in,
		})
	}
	return ingresses, nil
}

// Creates a StackState ingress component from a Kubernetes / OpenShift Ingress
func (ic *IngressCollector) ingressToStackStateComponent(ingress StsIngress) *topology.Component {
	log.Tracef("Mapping Ingress to StackState component: %s", ingress.String())

	tags := ic.initTags(ingress.GetObjectMeta())

	identifiers := make([]string, 0)

	name := ingress.GetName()
	ingressExternalID := ic.buildIngressExternalID(ingress.GetNamespace(), name)
	component := &topology.Component{
		ExternalID: ingressExternalID,
		Type:       topology.Type{Name: "ingress"},
		Data: map[string]interface{}{
			"name":        name,
			"tags":        tags,
			"identifiers": identifiers,
		},
	}

	if ic.IsSourcePropertiesFeatureEnabled() {
		object, err := ingress.GetKubernetesObject()
		log.Errorf("Could not get kubernetes object from ingress: %v", err)
		component.SourceProperties = makeSourceProperties(object)
	} else {
		component.Data.PutNonEmpty("creationTimestamp", ingress.GetCreationTimestamp())
		component.Data.PutNonEmpty("uid", ingress.GetUID())
		component.Data.PutNonEmpty("generateName", ingress.GetGenerateName())
		component.Data.PutNonEmpty("kind", ingress.GetKind())
	}

	log.Tracef("Created StackState Ingress component %s: %v", ingressExternalID, component.JSONString())

	return component
}

// Creates a StackState loadbalancer component from a Kubernetes / OpenShift Ingress
func (ic *IngressCollector) endpointStackStateComponentFromIngress(ingress StsIngress, ingressPoint string) *topology.Component {
	log.Tracef("Mapping Ingress to StackState endpoint component: %s", ingressPoint)

	tags := ic.initTags(ingress.GetObjectMeta())
	identifiers := make([]string, 0)
	endpointExternalID := ic.buildEndpointExternalID(ingressPoint)

	component := &topology.Component{
		ExternalID: endpointExternalID,
		Type:       topology.Type{Name: "endpoint"},
		Data: map[string]interface{}{
			"name":              ingressPoint,
			"creationTimestamp": ingress.GetCreationTimestamp,
			"tags":              tags,
			"identifiers":       identifiers,
		},
	}

	log.Tracef("Created StackState endpoint component %s: %v", endpointExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes / OpenShift Ingress to Service
func (ic *IngressCollector) ingressToServiceStackStateRelation(ingressExternalID, serviceExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes ingress to service relation: %s -> %s", ingressExternalID, serviceExternalID)

	relation := ic.CreateRelation(ingressExternalID, serviceExternalID, "routes")

	log.Tracef("Created StackState ingress -> service relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}

// Creates a StackState relation from an Endpoint to a Kubernetes / OpenShift Ingress
func (ic *IngressCollector) endpointToIngressStackStateRelation(endpointExternalID, ingressExternalID string) *topology.Relation {
	log.Tracef("Mapping endpoint to kubernetes ingress relation: %s -> %s", endpointExternalID, ingressExternalID)

	relation := ic.CreateRelation(endpointExternalID, ingressExternalID, "routes")

	log.Tracef("Created endpoint -> StackState ingress relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}

func (in StsIngress) GetServiceName() string {
	if in.v1beta1 != nil {
		if in.v1beta1.Spec.Backend != nil && in.v1beta1.Spec.Backend.ServiceName != "" {
			return in.v1beta1.Spec.Backend.ServiceName
		}
	} else if in.netv1 != nil {
		if in.netv1.Spec.DefaultBackend != nil && in.netv1.Spec.DefaultBackend.Service.Name != "" {
			return in.netv1.Spec.DefaultBackend.Service.Name
		}
	}
	return ""
}

func (in StsIngress) String() string {
	if in.v1beta1 != nil {
		return in.v1beta1.String()
	} else if in.netv1 != nil {
		return in.netv1.String()
	}
	return "invalid"
}

func (in StsIngress) GetCreationTimestamp() metav1.Time {
	if in.v1beta1 != nil {
		return in.v1beta1.CreationTimestamp
	} else {
		return in.netv1.CreationTimestamp
	}
}

func (in StsIngress) GetUID() types.UID {
	if in.v1beta1 != nil {
		return in.v1beta1.UID
	} else {
		return in.netv1.UID
	}
}

func (in StsIngress) GetGenerateName() string {
	if in.v1beta1 != nil {
		return in.v1beta1.GenerateName
	} else {
		return in.netv1.GenerateName
	}
}

func (in StsIngress) GetKind() string {
	if in.v1beta1 != nil {
		return in.v1beta1.Kind
	} else {
		return in.netv1.Kind
	}
}

func (in StsIngress) GetKubernetesObject() (MarshalableKubernetesObject, error) {
	if in.v1beta1 != nil {
		return in.v1beta1, nil
	} else if in.netv1 != nil {
		return in.netv1, nil
	}
	return nil, errors.New("Ingress not instance from extensionv1beta1 or networking.v1")
}

func (in StsIngress) GetObjectMeta() metav1.ObjectMeta {
	if in.v1beta1 != nil {
		return in.v1beta1.ObjectMeta
	} else if in.netv1 != nil {
		return in.netv1.ObjectMeta
	}
	return metav1.ObjectMeta{}
}

func (in StsIngress) GetName() string {
	if in.v1beta1 != nil {
		return in.v1beta1.Name
	} else if in.netv1 != nil {
		return in.netv1.Name
	}
	return "invalid"
}

func (in StsIngress) GetNamespace() string {
	if in.v1beta1 != nil {
		return in.v1beta1.Namespace
	} else if in.netv1 != nil {
		return in.netv1.Namespace
	}
	return ""
}

func (in StsIngress) GetServiceNames() []string {
	var result []string
	if in.v1beta1 != nil {
		for _, rules := range in.v1beta1.Spec.Rules {
			for _, path := range rules.HTTP.Paths {
				result = append(result, path.Backend.ServiceName)
			}
		}
	} else if in.netv1 != nil {
		for _, rules := range in.netv1.Spec.Rules {
			for _, path := range rules.HTTP.Paths {
				result = append(result, path.Backend.Service.Name)
			}
		}
	}
	return result
}

func (in StsIngress) GetIngressPoints() []string {
	var result []string
	var lbIngresses []v1.LoadBalancerIngress
	if in.v1beta1 != nil {
		lbIngresses = in.v1beta1.Status.LoadBalancer.Ingress
	} else if in.netv1 != nil {
		lbIngresses = in.netv1.Status.LoadBalancer.Ingress
	}
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
