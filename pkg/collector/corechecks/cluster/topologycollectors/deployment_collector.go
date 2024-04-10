//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentCollector implements the ClusterTopologyCollector interface.
type DeploymentCollector struct {
	ClusterTopologyCollector
}

// NewDeploymentCollector
func NewDeploymentCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &DeploymentCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*DeploymentCollector) GetName() string {
	return "Deployment Collector"
}

// Collects and Published the Deployment Components
func (dmc *DeploymentCollector) CollectorFunction() error {
	deployments, err := dmc.GetAPIClient().GetDeployments()
	if err != nil {
		return err
	}

	for _, dep := range deployments {
		component := dmc.deploymentToStackStateComponent(dep)
		dmc.SubmitComponent(component)

		dmc.SubmitRelation(dmc.namespaceToDeploymentStackStateRelation(dmc.buildNamespaceExternalID(dep.Namespace), component.ExternalID))
	}

	return nil
}

func (dmc *DeploymentCollector) DeploymentToStackStateComponent(deployment v1.Deployment) *topology.Component {
	return dmc.deploymentToStackStateComponent(deployment)
}

// Creates a StackState deployment component from a Kubernetes / OpenShift Cluster
func (dmc *DeploymentCollector) deploymentToStackStateComponent(deployment v1.Deployment) *topology.Component {
	log.Tracef("Mapping Deployment to StackState component: %s", deployment.String())

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := dmc.initTags(deployment.ObjectMeta, metav1.TypeMeta{Kind: "Deployment"})

	deploymentExternalID := dmc.buildDeploymentExternalID(deployment.Namespace, deployment.Name)
	component := &topology.Component{
		ExternalID: deploymentExternalID,
		Type:       topology.Type{Name: "deployment"},
		Data: map[string]interface{}{
			"name": deployment.Name,
			"tags": tags,
		},
	}

	if dmc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if dmc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&deployment)
		} else {
			sourceProperties = makeSourceProperties(&deployment)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("kind", deployment.Kind)
		component.Data.PutNonEmpty("uid", deployment.UID)
		component.Data.PutNonEmpty("creationTimestamp", deployment.CreationTimestamp)
		component.Data.PutNonEmpty("generateName", deployment.GenerateName)
		component.Data.PutNonEmpty("deploymentStrategy", deployment.Spec.Strategy.Type)
		component.Data.PutNonEmpty("desiredReplicas", deployment.Spec.Replicas)
	}

	log.Tracef("Created StackState Deployment component %s: %v", deploymentExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes / OpenShift Namespace to Deployment relation
func (dmc *DeploymentCollector) namespaceToDeploymentStackStateRelation(namespaceExternalID, deploymentExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes namespace to deployment relation: %s -> %s", namespaceExternalID, deploymentExternalID)

	relation := dmc.CreateRelation(namespaceExternalID, deploymentExternalID, "encloses")

	log.Tracef("Created StackState namespace -> deployment relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}
