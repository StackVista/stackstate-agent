//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"k8s.io/apimachinery/pkg/version"
	"strconv"
)

// IngressCollector implements the ClusterTopologyCollector interface.
type IngressCollector struct {
	ComponentChan chan<- *topology.Component
	RelationChan  chan<- *topology.Relation
	k8sVersion    *version.Info
	ClusterTopologyCollector
}

// NewIngressCollector
func NewIngressCollector(componentChannel chan<- *topology.Component, relationChannel chan<- *topology.Relation,
	clusterTopologyCollector ClusterTopologyCollector, k8sVersion *version.Info) ClusterTopologyCollector {
	return &IngressCollector{
		ComponentChan:            componentChannel,
		RelationChan:             relationChannel,
		ClusterTopologyCollector: clusterTopologyCollector,
		k8sVersion:               k8sVersion,
	}
}

// GetName returns the name of the Collector
func (*IngressCollector) GetName() string {
	return "Ingress Collector"
}

// Collects and Published the Ingress Components
func (ic *IngressCollector) CollectorFunction() error {
	var ingresses []IngressInterface
	ingresses, err := ic.getExtV1Ingresses(ingresses)
	if err != nil {
		return err
	}
	ingresses, err = ic.getNetV1Ingresses(ingresses)
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

func (ic *IngressCollector) getExtV1Ingresses(ingresses []IngressInterface) ([]IngressInterface, error) {
	log.Debugf("Kubernetes version is %+v", ic.k8sVersion)
	if ic.k8sVersion != nil && ic.k8sVersion.Major == "1" {
		log.Debugf("Kubernetes version is Major=%s, Minor=%s", ic.k8sVersion.Major, ic.k8sVersion.Minor)
		minor, err := strconv.Atoi(ic.k8sVersion.Minor[:2])
		if err != nil {
			return ingresses, fmt.Errorf("cannot parse server minor version %q: %w", ic.k8sVersion.Minor[:2], err)
		}
		if minor >= 22 {
			log.Debugf("Kubernetes version is >= 1.22, the topology collector will NOT query ingresses from 'extensions/v1beta1' version")
			return ingresses, nil
		}
		log.Debugf("Kubernetes version is <= 1.21, the topology collector will query ingresses from 'extensions/v1beta1' version")
	}
	ingressesExt, err := ic.GetAPIClient().GetIngressesExtV1B1()
	if err != nil {
		return nil, err
	}
	for _, in := range ingressesExt {
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
		ingresses = append(ingresses, IngressNetV1{
			o: in,
		})
	}
	return ingresses, nil
}

// Creates a StackState ingress component from a Kubernetes / OpenShift Ingress
func (ic *IngressCollector) ingressToStackStateComponent(ingress IngressInterface) *topology.Component {
	log.Tracef("Mapping Ingress to StackState component: %s", ingress.GetString())

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
		object := ingress.GetKubernetesObject()
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
func (ic *IngressCollector) endpointStackStateComponentFromIngress(ingress IngressInterface, ingressPoint string) *topology.Component {
	log.Tracef("Mapping Ingress to StackState endpoint component: %s", ingressPoint)

	tags := ic.initTags(ingress.GetObjectMeta())
	identifiers := make([]string, 0)
	endpointExternalID := ic.buildEndpointExternalID(ingressPoint)

	component := &topology.Component{
		ExternalID: endpointExternalID,
		Type:       topology.Type{Name: "endpoint"},
		Data: map[string]interface{}{
			"name":              ingressPoint,
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
