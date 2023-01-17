//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
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

	tags := nsc.initTags(namespace.ObjectMeta)
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
