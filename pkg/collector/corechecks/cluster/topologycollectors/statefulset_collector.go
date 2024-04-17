//go:build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSetCollector implements the ClusterTopologyCollector interface.
type StatefulSetCollector struct {
	ClusterTopologyCollector
}

// NewStatefulSetCollector
func NewStatefulSetCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &StatefulSetCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*StatefulSetCollector) GetName() string {
	return "StatefulSet Collector"
}

// Collects and Published the StatefulSet Components
func (ssc *StatefulSetCollector) CollectorFunction() error {
	statefulSets, err := ssc.GetAPIClient().GetStatefulSets()
	if err != nil {
		return err
	}

	for _, ss := range statefulSets {
		component := ssc.statefulSetToStackStateComponent(ss)
		ssc.SubmitComponent(component)
		ssc.SubmitRelation(ssc.namespaceToStatefulSetStackStateRelation(ssc.buildNamespaceExternalID(ss.Namespace), component.ExternalID))
	}

	return nil
}

// Creates a StackState component from a Kubernetes / OpenShift Cluster
func (ssc *StatefulSetCollector) statefulSetToStackStateComponent(statefulSet v1.StatefulSet) *topology.Component {
	log.Tracef("Mapping StatefulSet to StackState component: %s", statefulSet.String())

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := ssc.initTags(statefulSet.ObjectMeta, metav1.TypeMeta{Kind: "StatefulSet"})

	statefulSetExternalID := ssc.buildStatefulSetExternalID(statefulSet.Namespace, statefulSet.Name)
	component := &topology.Component{
		ExternalID: statefulSetExternalID,
		Type:       topology.Type{Name: "statefulset"},
		Data: map[string]interface{}{
			"name": statefulSet.Name,
			"tags": tags,
		},
	}

	if ssc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if ssc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&statefulSet)
		} else {
			sourceProperties = makeSourceProperties(&statefulSet)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("kind", statefulSet.Kind)
		component.Data.PutNonEmpty("uid", statefulSet.UID)
		component.Data.PutNonEmpty("generateName", statefulSet.GenerateName)
		component.Data.PutNonEmpty("creationTimestamp", statefulSet.CreationTimestamp)
		component.Data.PutNonEmpty("updateStrategy", statefulSet.Spec.UpdateStrategy.Type)
		component.Data.PutNonEmpty("desiredReplicas", statefulSet.Spec.Replicas)
		component.Data.PutNonEmpty("podManagementPolicy", statefulSet.Spec.PodManagementPolicy)
		component.Data.PutNonEmpty("serviceName", statefulSet.Spec.ServiceName)
	}

	log.Tracef("Created StackState StatefulSet component %s: %v", statefulSetExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes / OpenShift Namespace to StatefulSet relation
func (ssc *StatefulSetCollector) namespaceToStatefulSetStackStateRelation(namespaceExternalID, statefulSetExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes namespace to stateful set relation: %s -> %s", namespaceExternalID, statefulSetExternalID)

	relation := ssc.CreateRelation(namespaceExternalID, statefulSetExternalID, "encloses")

	log.Tracef("Created StackState namespace -> stateful set relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}
