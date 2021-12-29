// +build kubeapiserver

package topologycollectors

import (
	"bytes"
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"k8s.io/api/apps/v1"
)

// DeploymentCollector implements the ClusterTopologyCollector interface.
type DeploymentCollector struct {
	ComponentChan chan<- *topology.Component
	RelationChan  chan<- *topology.Relation
	ClusterTopologyCollector
}

// NewDeploymentCollector
func NewDeploymentCollector(componentChannel chan<- *topology.Component, relationChannel chan<- *topology.Relation, clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &DeploymentCollector{
		ComponentChan:            componentChannel,
		RelationChan:             relationChannel,
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
		dmc.ComponentChan <- component

		dmc.RelationChan <- dmc.namespaceToDeploymentStackStateRelation(dmc.buildNamespaceExternalID(dep.Namespace), component.ExternalID)
	}

	return nil
}

var JsonMarshaler = jsonpb.Marshaler{
	EnumsAsInts:  false,
	EmitDefaults: false,
}

func marshalToData(msg proto.Message) (map[string]interface{}, error) {
	var buf bytes.Buffer
	if err := JsonMarshaler.Marshal(&buf, msg); err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (dmc *DeploymentCollector) DeploymentToStackStateComponent(deployment v1.Deployment) *topology.Component {
	return dmc.deploymentToStackStateComponent(deployment)
}

// Creates a StackState deployment component from a Kubernetes / OpenShift Cluster
func (dmc *DeploymentCollector) deploymentToStackStateComponent(deployment v1.Deployment) *topology.Component {
	log.Tracef("Mapping Deployment to StackState component: %s", deployment.String())

	tags := dmc.initTags(deployment.ObjectMeta)

	sourceProperties, err := marshalToData(&deployment)
	if err != nil {
		_ = log.Warnf("Can't serialize sourceProperties for Deployment %s/%s: %v", deployment.Namespace, deployment.Name, err)
	}

	deploymentExternalID := dmc.buildDeploymentExternalID(deployment.Namespace, deployment.Name)
	component := &topology.Component{
		ExternalID: deploymentExternalID,
		Type:       topology.Type{Name: "deployment"},
		Data: map[string]interface{}{
			"name":               deployment.Name,
			"creationTimestamp":  deployment.CreationTimestamp,
			"tags":               tags,
			"deploymentStrategy": deployment.Spec.Strategy.Type,
			"desiredReplicas":    deployment.Spec.Replicas,
			"uid":                deployment.UID,
		},
		SourceProperties: sourceProperties,
	}

	component.Data.PutNonEmpty("generateName", deployment.GenerateName)
	component.Data.PutNonEmpty("kind", deployment.Kind)

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
