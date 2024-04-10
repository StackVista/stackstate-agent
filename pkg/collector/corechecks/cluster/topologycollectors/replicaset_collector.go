//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicaSetCollector implements the ClusterTopologyCollector interface.
type ReplicaSetCollector struct {
	ClusterTopologyCollector
}

// GetName returns the name of the Collector
func (*ReplicaSetCollector) GetName() string {
	return "ReplicaSet Collector"
}

// NewReplicaSetCollector
func NewReplicaSetCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &ReplicaSetCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// Collects and Published the ReplicaSet Components
func (rsc *ReplicaSetCollector) CollectorFunction() error {
	replicaSets, err := rsc.GetAPIClient().GetReplicaSets()
	if err != nil {
		return err
	}

	for _, rs := range replicaSets {
		component := rsc.replicaSetToStackStateComponent(rs)
		rsc.SubmitComponent(component)

		controlled := false
		// check to see if this pod is "controlled" by a deployment
		for _, ref := range rs.OwnerReferences {
			switch kind := ref.Kind; kind {
			case Deployment:
				dmExternalID := rsc.buildDeploymentExternalID(rs.Namespace, ref.Name)
				rsc.SubmitRelation(rsc.deploymentToReplicaSetStackStateRelation(dmExternalID, component.ExternalID))
				controlled = true
			}
		}

		if !controlled {
			rsc.SubmitRelation(rsc.namespaceToReplicaSetStackStateRelation(rsc.buildNamespaceExternalID(rs.Namespace), component.ExternalID))
		}

	}

	return nil
}

// Creates a StackState component from a Kubernetes / OpenShift Cluster
func (rsc *ReplicaSetCollector) replicaSetToStackStateComponent(replicaSet v1.ReplicaSet) *topology.Component {
	log.Tracef("Mapping ReplicaSet to StackState component: %s", replicaSet.String())

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := rsc.initTags(replicaSet.ObjectMeta, metav1.TypeMeta{Kind: "ReplicaSet"})

	replicaSetExternalID := rsc.buildReplicaSetExternalID(replicaSet.Namespace, replicaSet.Name)
	component := &topology.Component{
		ExternalID: replicaSetExternalID,
		Type:       topology.Type{Name: "replicaset"},
		Data: map[string]interface{}{
			"name": replicaSet.Name,
			"tags": tags,
		},
	}

	if rsc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if rsc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&replicaSet)
		} else {
			sourceProperties = makeSourceProperties(&replicaSet)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("kind", replicaSet.Kind)
		component.Data.PutNonEmpty("creationTimestamp", replicaSet.CreationTimestamp)
		component.Data.PutNonEmpty("generateName", replicaSet.GenerateName)
		component.Data.PutNonEmpty("uid", replicaSet.UID)
		component.Data.PutNonEmpty("desiredReplicas", replicaSet.Spec.Replicas)
	}

	log.Tracef("Created StackState ReplicaSet component %s: %v", replicaSetExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes / OpenShift Deployment to ReplicaSet relation
func (rsc *ReplicaSetCollector) deploymentToReplicaSetStackStateRelation(deploymentExternalID, replicaSetExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes deployment to replica set relation: %s -> %s", deploymentExternalID, replicaSetExternalID)

	relation := rsc.CreateRelation(deploymentExternalID, replicaSetExternalID, "controls")

	log.Tracef("Created StackState deployment -> replica set relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}

// Creates a StackState relation from a Kubernetes / OpenShift Namespace to Pod relation
func (rsc *ReplicaSetCollector) namespaceToReplicaSetStackStateRelation(namespaceExternalID, replicaSetExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes namespace to replica set relation: %s -> %s", namespaceExternalID, replicaSetExternalID)

	relation := rsc.CreateRelation(namespaceExternalID, replicaSetExternalID, "encloses")

	log.Tracef("Created StackState namespace -> replica set relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}
