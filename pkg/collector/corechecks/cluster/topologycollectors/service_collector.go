//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/dns"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
)

// ServiceCollector implements the ClusterTopologyCollector interface.
type ServiceCollector struct {
	SelectorCorrChan chan<- *ServiceSelectorCorrelation
	ClusterTopologyCollector
	DNS dns.Resolver
}

// EndpointID contains the definition of a cluster ip
type EndpointID struct {
	URL           string
	RefExternalID string
}

// NewServiceCollector
func NewServiceCollector(
	serviceCorrChannel chan *ServiceSelectorCorrelation,
	clusterTopologyCollector ClusterTopologyCollector,
) ClusterTopologyCollector {
	return &ServiceCollector{
		SelectorCorrChan:         serviceCorrChannel,
		ClusterTopologyCollector: clusterTopologyCollector,
		DNS:                      dns.StandardResolver,
	}
}

// GetName returns the name of the Collector
func (*ServiceCollector) GetName() string {
	return "Service Collector"
}

// Collects and Published the Service Components
func (sc *ServiceCollector) CollectorFunction() error {
	// close seelector correlation channel
	// it will signal service2pod correlator to proceed
	defer close(sc.SelectorCorrChan)

	services, err := sc.GetAPIClient().GetServices()
	if err != nil {
		return err
	}

	serviceMap := make(map[string][]string)

	for _, service := range services {
		// creates and publishes StackState service component with relations
		serviceID := buildServiceID(service.Namespace, service.Name)
		component := sc.serviceToStackStateComponent(service)

		sc.SubmitComponent(component)

		// Check whether we have an ExternalName service, which will result in an extra component+relation
		if service.Spec.Type == v1.ServiceTypeExternalName {
			externalService := sc.serviceToExternalServiceComponent(service)

			sc.SubmitComponent(externalService)
			sc.SubmitRelation(sc.serviceToExternalServiceStackStateRelation(component.ExternalID, externalService.ExternalID))
		}

		// First ensure we publish all components, else the test becomes complex
		sc.SubmitRelation(sc.namespaceToServiceStackStateRelation(sc.buildNamespaceExternalID(service.Namespace), component.ExternalID))

		sc.SelectorCorrChan <- &ServiceSelectorCorrelation{
			ServiceExternalID: component.ExternalID,
			Namespace:         service.Namespace,
			LabelSelector:     service.Spec.Selector,
		}

		serviceMap[serviceID] = append(serviceMap[serviceID], component.ExternalID)
	}

	return nil
}

// Creates a StackState component from a Kubernetes / OpenShift Service
func (sc *ServiceCollector) serviceToStackStateComponent(service v1.Service) *topology.Component {
	log.Tracef("Mapping kubernetes pod service to StackState component: %s", service.String())
	identifiers := sc.identifiers(service)

	log.Tracef("Created identifiers for %s: %v", service.Name, identifiers)

	serviceExternalID := sc.buildServiceExternalID(service.Namespace, service.Name)

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := sc.initTags(service.ObjectMeta, metav1.TypeMeta{Kind: "Service"})
	tags["service-type"] = string(service.Spec.Type)

	if service.Spec.ClusterIP == "None" {
		tags["service"] = "headless"
	}

	component := &topology.Component{
		ExternalID: serviceExternalID,
		Type:       topology.Type{Name: "service"},
		Data: map[string]interface{}{
			"name":        service.Name,
			"tags":        tags,
			"identifiers": identifiers,
		},
	}

	if sc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if sc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&service)
		} else {
			sourceProperties = makeSourceProperties(&service)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("creationTimestamp", service.CreationTimestamp)
		component.Data.PutNonEmpty("uid", service.UID)
		component.Data.PutNonEmpty("kind", service.Kind)
		component.Data.PutNonEmpty("generateName", service.GenerateName)
	}

	log.Tracef("Created StackState service component %s: %v", serviceExternalID, component.JSONString())

	return component
}

func (sc *ServiceCollector) identifiers(service v1.Service) []string {
	// create identifier list to merge with StackState components
	identifiers := make([]string, 0)

	// all external ip's which are associated with this service, but are not managed by kubernetes
	for _, ip := range service.Spec.ExternalIPs {
		// verify that the ip is not empty
		if ip == "" {
			continue
		}
		// map all of the ports for the ip
		for _, port := range service.Spec.Ports {
			identifiers = append(identifiers, fmt.Sprintf("urn:endpoint:/%s:%d", ip, port.Port))

			if service.Spec.Type == v1.ServiceTypeNodePort && port.NodePort != 0 {
				identifiers = append(identifiers, fmt.Sprintf("urn:endpoint:/%s:%d", ip, port.NodePort))
			}
		}
	}

	switch service.Spec.Type {
	// identifier for
	case v1.ServiceTypeClusterIP:
		// verify that the cluster ip is not empty
		if service.Spec.ClusterIP != "None" && service.Spec.ClusterIP != "" {
			identifiers = append(identifiers, sc.buildEndpointExternalID(service.Spec.ClusterIP))
		}
	case v1.ServiceTypeNodePort:
		// verify that the node port is not empty
		if service.Spec.ClusterIP != "None" && service.Spec.ClusterIP != "" {
			identifiers = append(identifiers, sc.buildEndpointExternalID(service.Spec.ClusterIP))
			// map all of the node ports for the ip
			for _, port := range service.Spec.Ports {
				// map all the node ports
				if port.NodePort != 0 {
					identifiers = append(identifiers, sc.buildEndpointExternalID(fmt.Sprintf("%s:%d", service.Spec.ClusterIP, port.NodePort)))
				}
			}
		}
	case v1.ServiceTypeLoadBalancer:
		// verify that the load balance ip is not empty
		if service.Spec.LoadBalancerIP != "" {
			identifiers = append(identifiers, sc.buildEndpointExternalID(service.Spec.LoadBalancerIP))
		}
		// verify that the cluster ip is not empty
		if service.Spec.ClusterIP != "None" && service.Spec.ClusterIP != "" {
			identifiers = append(identifiers, sc.buildEndpointExternalID(service.Spec.ClusterIP))
		}
	case v1.ServiceTypeExternalName:
	default:
	}

	for _, inpoint := range service.Status.LoadBalancer.Ingress {
		if inpoint.IP != "" {
			identifiers = append(identifiers, fmt.Sprintf("urn:ingress-point:/%s", inpoint.IP))
		}

		if inpoint.Hostname != "" {
			identifiers = append(identifiers, fmt.Sprintf("urn:ingress-point:/%s", inpoint.Hostname))

		}
	}

	// add identifier for this service name
	serviceID := buildServiceID(service.Namespace, service.Name)
	identifiers = append(identifiers, fmt.Sprintf("urn:service:/%s:%s", sc.GetInstance().URL, serviceID))

	return identifiers
}

func (sc *ServiceCollector) serviceToExternalServiceComponent(service v1.Service) *topology.Component {
	log.Tracef("Mapping kubernetes pod ExternalName service to extra StackState component: %s", service.String())
	// create identifier list to merge with StackState components
	identifiers := make([]string, 0)

	if service.Spec.ExternalName != "None" && service.Spec.ExternalName != "" {
		identifiers = append(identifiers, fmt.Sprintf("urn:endpoint:/%s", service.Spec.ExternalName))
		// If targetPorts are specified, use those
		for _, port := range service.Spec.Ports {
			// map all the node ports
			if port.Port != 0 {
				identifiers = append(identifiers, sc.buildEndpointExternalID(fmt.Sprintf("%s:%d", service.Spec.ExternalName, port.Port)))
			}
		}

		addrs, err := sc.DNS(service.Spec.ExternalName)
		if err != nil {
			log.Warnf("Could not lookup IP addresses for host '%s' (Error: %s)", service.Spec.ExternalName, err.Error())
		} else {
			for _, addr := range addrs {
				identifiers = append(identifiers, sc.buildEndpointExternalID(addr))
				// If targetPorts are specified, use those
				for _, port := range service.Spec.Ports {
					// map all the node ports
					if port.Port != 0 {
						identifiers = append(identifiers, sc.buildEndpointExternalID(fmt.Sprintf("%s:%d", addr, port.Port)))
					}
				}
			}
		}
	}

	// add identifier for this service name
	serviceID := buildServiceID(service.Namespace, service.Name)
	identifiers = append(identifiers, fmt.Sprintf("urn:external-service:/%s:%s", sc.GetInstance().URL, serviceID))

	log.Tracef("Created identifiers for %s: %v", service.Name, identifiers)

	externalID := sc.GetURNBuilder().BuildComponentExternalID("external-service", service.Namespace, service.Name)

	tags := sc.initTags(service.ObjectMeta, metav1.TypeMeta{Kind: "ExternalService"})

	component := &topology.Component{
		ExternalID: externalID,
		Type:       topology.Type{Name: "external-service"},
		Data: map[string]interface{}{
			"name":              service.Name,
			"kind":              "ExternalService",
			"creationTimestamp": service.CreationTimestamp,
			"tags":              tags,
			"identifiers":       identifiers,
			"uid":               service.UID,
		},
	}

	log.Tracef("Created StackState external-service component %s: %v", externalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes / OpenShift Service to Namespace relation
func (sc *ServiceCollector) namespaceToServiceStackStateRelation(namespaceExternalID, serviceExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes namespace to service relation: %s -> %s", namespaceExternalID, serviceExternalID)

	relation := sc.CreateRelation(namespaceExternalID, serviceExternalID, "encloses")

	log.Tracef("Created StackState namespace -> service relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}

// Creates a StackState relation from a Kubernetes / OpenShift Service to 'ExternalService' relation
func (sc *ServiceCollector) serviceToExternalServiceStackStateRelation(serviceExternalID, externalServiceExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes service to external service relation: %s -> %s", serviceExternalID, externalServiceExternalID)

	relation := sc.CreateRelation(serviceExternalID, externalServiceExternalID, "uses")

	log.Tracef("Created StackState service -> external service relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}

// buildServiceID - combination of the service namespace and service name
func buildServiceID(serviceNamespace, serviceName string) string {
	return fmt.Sprintf("%s:%s", serviceNamespace, serviceName)
}
