//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSetCollector implements the ClusterTopologyCollector interface.
type DaemonSetCollector struct {
	ClusterTopologyCollector
}

// NewDaemonSetCollector
func NewDaemonSetCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &DaemonSetCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*DaemonSetCollector) GetName() string {
	return "DaemonSet Collector"
}

// Collects and Published the DaemonSet Components
func (dsc *DaemonSetCollector) CollectorFunction() error {
	daemonSets, err := dsc.GetAPIClient().GetDaemonSets()
	if err != nil {
		return err
	}

	for _, ds := range daemonSets {
		component := dsc.daemonSetToStackStateComponent(ds)
		dsc.SubmitComponent(component)
		dsc.SubmitRelation(dsc.namespaceToDaemonSetStackStateRelation(dsc.buildNamespaceExternalID(ds.Namespace), component.ExternalID))
	}

	return nil
}

// Creates a StackState daemonset component from a Kubernetes / OpenShift Cluster
func (dsc *DaemonSetCollector) daemonSetToStackStateComponent(daemonSet v1.DaemonSet) *topology.Component {
	log.Tracef("Mapping DaemonSet to StackState component: %s", daemonSet.String())

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := dsc.initTags(daemonSet.ObjectMeta, metav1.TypeMeta{Kind: "DaemonSet"})

	daemonSetExternalID := dsc.buildDaemonSetExternalID(daemonSet.Namespace, daemonSet.Name)
	component := &topology.Component{
		ExternalID: daemonSetExternalID,
		Type:       topology.Type{Name: "daemonset"},
		Data: map[string]interface{}{
			"name": daemonSet.Name,
			"tags": tags,
		},
	}

	if dsc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if dsc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&daemonSet)
		} else {
			sourceProperties = makeSourceProperties(&daemonSet)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("kind", daemonSet.Kind)
		component.Data.PutNonEmpty("uid", daemonSet.UID)
		component.Data.PutNonEmpty("creationTimestamp", daemonSet.CreationTimestamp)
		component.Data.PutNonEmpty("generateName", daemonSet.GenerateName)
		component.Data.PutNonEmpty("updateStrategy", daemonSet.Spec.UpdateStrategy.Type)
	}

	log.Tracef("Created StackState DaemonSet component %s: %v", daemonSetExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes / OpenShift Namespace to DaemonSet relation
func (dsc *DaemonSetCollector) namespaceToDaemonSetStackStateRelation(namespaceExternalID, daemonSetExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes namespace to daemon set relation: %s -> %s", namespaceExternalID, daemonSetExternalID)

	relation := dsc.CreateRelation(namespaceExternalID, daemonSetExternalID, "encloses")

	log.Tracef("Created StackState namespace -> daemon set relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}
