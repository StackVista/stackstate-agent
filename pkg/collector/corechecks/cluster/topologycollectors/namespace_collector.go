//go:build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NamespaceCollector implements the ClusterTopologyCollector interface.
type NamespaceCollector struct {
	ClusterTopologyCollector
}

// NewNamespaceCollector
func NewNamespaceCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &NamespaceCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*NamespaceCollector) GetName() string {
	return "Namespace Collector"
}

// Collects and Published the Namespace Components
func (nsc *NamespaceCollector) CollectorFunction() error {
	namespaces, err := nsc.GetAPIClient().GetNamespaces()
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		nsc.SubmitComponent(nsc.namespaceToStackStateComponent(ns))
	}

	return nil
}

// Creates a StackState Namespace component from a Kubernetes / OpenShift Cluster
func (nsc *NamespaceCollector) namespaceToStackStateComponent(namespace v1.Namespace) *topology.Component {
	log.Tracef("Mapping Namespace to StackState component: %s", namespace.String())

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := nsc.initTags(namespace.ObjectMeta, metav1.TypeMeta{Kind: "Namespace"})
	namespaceExternalID := nsc.buildNamespaceExternalID(namespace.Name)

	component := &topology.Component{
		ExternalID: namespaceExternalID,
		Type:       topology.Type{Name: "namespace"},
		Data: map[string]interface{}{
			"name":        namespace.Name,
			"tags":        tags,
			"identifiers": []string{namespaceExternalID},
		},
	}

	if nsc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if nsc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&namespace)
		} else {
			sourceProperties = makeSourceProperties(&namespace)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("creationTimestamp", namespace.CreationTimestamp)
		component.Data.PutNonEmpty("uid", namespace.UID)
		component.Data.PutNonEmpty("generateName", namespace.GenerateName)
		component.Data.PutNonEmpty("kind", namespace.Kind)
	}

	log.Tracef("Created StackState Namespace component %s: %v", namespaceExternalID, component.JSONString())

	return component
}
