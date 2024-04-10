//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IngressCollector implements the ClusterTopologyCollector interface.
type IngressCollector struct {
	ClusterTopologyCollector
}

// NewIngressCollector creates a new Ingress collector
func NewIngressCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &IngressCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*IngressCollector) GetName() string {
	return "Ingress Collector"
}

// CollectorFunction Collects and Published the Ingress Components
func (ic *IngressCollector) CollectorFunction() error {
	var ingresses []IngressInterface
	var err error
	if supported := ic.ClusterTopologyCollector.minimumMinorVersion(19); supported {
		ingresses, err = ic.getNetV1Ingresses(ingresses)
	} else {
		ingresses, err = ic.getExtV1Ingresses(ingresses)
	}
	if err != nil {
		return err
	}

	for _, in := range ingresses {
		component := ic.ingressToStackStateComponent(in)
		ic.SubmitComponent(component)
		// submit relation to service name for correlation
		if in.GetServiceName() != "" {
			serviceExternalID := ic.buildServiceExternalID(in.GetNamespace(), in.GetServiceName())

			// publish the ingress -> service relation
			relation := ic.ingressToServiceStackStateRelation(component.ExternalID, serviceExternalID)
			ic.SubmitRelation(relation)
		}

		// submit relation to service name in the ingress rules for correlation
		for _, serviceName := range in.GetServiceNames() {
			serviceExternalID := ic.buildServiceExternalID(in.GetNamespace(), serviceName)

			// publish the ingress -> service relation
			relation := ic.ingressToServiceStackStateRelation(component.ExternalID, serviceExternalID)
			ic.SubmitRelation(relation)
		}

		// submit relation to loadbalancer

		for _, ingressPoint := range in.GetIngressPoints() {
			endpoint := ic.endpointStackStateComponentFromIngress(in, ingressPoint)

			ic.SubmitComponent(endpoint)
			ic.SubmitRelation(ic.endpointToIngressStackStateRelation(endpoint.ExternalID, component.ExternalID))
		}
	}

	return nil
}

func (ic *IngressCollector) getExtV1Ingresses(ingresses []IngressInterface) ([]IngressInterface, error) {
	ingressesExt, err := ic.GetAPIClient().GetIngressesExtV1B1()
	if err != nil {
		return nil, err
	}
	for _, in := range ingressesExt {
		log.Debugf("Got Ingress '%s' from extensions/v1beta1", in.Name)
		ingresses = append(ingresses, IngressV1B1{
			o: in,
		})
	}
	return ingresses, nil
}

func (ic *IngressCollector) getNetV1Ingresses(ingresses []IngressInterface) ([]IngressInterface, error) {
	ingressesNetV1, err := ic.GetAPIClient().GetIngressesNetV1()
	if err != nil {
		return nil, err
	}
	for _, in := range ingressesNetV1 {
		log.Debugf("Got Ingress '%s' from networking.k8s.io/v1", in.Name)
		ingresses = append(ingresses, IngressNetV1{
			o: in,
		})
	}
	return ingresses, nil
}

// Creates a StackState ingress component from a Kubernetes / OpenShift Ingress
func (ic *IngressCollector) ingressToStackStateComponent(ingress IngressInterface) *topology.Component {
	log.Tracef("Mapping Ingress to StackState component: %s", ingress.GetString())

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := ic.initTags(ingress.GetObjectMeta(), metaV1.TypeMeta{Kind: "Ingress"})

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
		object := ingress.GetKubernetesObject()
		var sourceProperties map[string]interface{}
		if ic.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(object)
		} else {
			sourceProperties = makeSourceProperties(object)
		}
		component.SourceProperties = sourceProperties
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
func (ic *IngressCollector) endpointStackStateComponentFromIngress(ingress IngressInterface, ingressPoint string) *topology.Component {
	log.Tracef("Mapping Ingress to StackState endpoint component: %s", ingressPoint)

	tags := ic.initTags(ingress.GetObjectMeta(), metaV1.TypeMeta{Kind: "Endpoint"})
	identifiers := make([]string, 0)
	endpointExternalID := ic.buildEndpointExternalID(ingressPoint)

	component := &topology.Component{
		ExternalID: endpointExternalID,
		Type:       topology.Type{Name: "endpoint"},
		Data: map[string]interface{}{
			"name":              ingressPoint,
			"kind":              "Endpoint",
			"creationTimestamp": ingress.GetCreationTimestamp(),
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
