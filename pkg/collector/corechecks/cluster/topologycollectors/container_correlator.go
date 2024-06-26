//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"

	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ContainerToNodeCorrelation
type NodeIdentifierCorrelation struct {
	NodeName       string
	NodeIdentifier string
	NodeExternalID string
}

// ContainerPod
type ContainerPod struct {
	ExternalID string
	Name       string
	Labels     map[string]string
	PodIP      string
	Namespace  string
	NodeName   string
	Phase      string
}

// ContainerCorrelation
type ContainerCorrelation struct {
	Pod               ContainerPod
	Containers        []v1.Container
	ContainerStatuses []v1.ContainerStatus
}

// ContainerCorrelator implements the ClusterTopologyCollector interface.
type ContainerCorrelator struct {
	NodeIdentifierCorrChan <-chan *NodeIdentifierCorrelation
	ContainerCorrChan      <-chan *ContainerCorrelation
	ClusterTopologyCorrelator
}

// NewContainerCorrelator
func NewContainerCorrelator(
	nodeIdentifierCorrChan <-chan *NodeIdentifierCorrelation, containerCorrChannel <-chan *ContainerCorrelation,
	clusterTopologyCorrelator ClusterTopologyCorrelator) ClusterTopologyCorrelator {
	return &ContainerCorrelator{
		NodeIdentifierCorrChan:    nodeIdentifierCorrChan,
		ContainerCorrChan:         containerCorrChannel,
		ClusterTopologyCorrelator: clusterTopologyCorrelator,
	}
}

// GetName returns the name of the Collector
func (*ContainerCorrelator) GetName() string {
	return "Container Correlator"
}

// Collects and Published the Cluster Component
func (cc *ContainerCorrelator) CorrelateFunction() error {
	nodeMap := make(map[string]NodeIdentifierCorrelation)
	// map containers that require the Node instanceId
	for nodeNameToNodeIdentifierCorrelation := range cc.NodeIdentifierCorrChan {
		nodeMap[nodeNameToNodeIdentifierCorrelation.NodeName] = *nodeNameToNodeIdentifierCorrelation
	}

	for containerCorrelation := range cc.ContainerCorrChan {
		pod := containerCorrelation.Pod
		// map container to exposed ports
		containerPorts := make(map[string]ContainerPort)
		for _, c := range containerCorrelation.Containers {

			for _, port := range c.Ports {
				containerPorts[fmt.Sprintf("%s_%s", c.Image, c.Name)] = ContainerPort{
					HostPort:      port.HostPort,
					ContainerPort: port.ContainerPort,
				}
			}
		}

		// check to see if we have container statuses
		for _, container := range containerCorrelation.ContainerStatuses {
			containerPort := ContainerPort{}
			if cntPort, found := containerPorts[fmt.Sprintf("%s_%s", container.Image, container.Name)]; found {
				containerPort = cntPort
			}

			if nodeCorrelation, ok := nodeMap[pod.NodeName]; ok {
				// submit the StackState component for publishing to StackState
				containerComponent := cc.containerToStackStateComponent(nodeCorrelation.NodeIdentifier, pod, container, containerPort)
				cc.SubmitComponent(containerComponent)
				// create the relation between the container and pod
				cc.SubmitRelation(cc.podToContainerStackStateRelation(pod.ExternalID, containerComponent.ExternalID))
				// create the relation between the container and node
				cc.SubmitRelation(cc.containerToNodeStackStateRelation(containerComponent.ExternalID, nodeCorrelation.NodeExternalID))
			}
		}
	}

	return nil
}

// Creates a StackState component from a Kubernetes / OpenShift Pod Container
func (cc *ContainerCorrelator) containerToStackStateComponent(nodeIdentifier string, pod ContainerPod, container v1.ContainerStatus, containerPort ContainerPort) *topology.Component {
	log.Tracef("Mapping kubernetes pod container to StackState component: %s", container.String())
	// create identifier list to merge with StackState components

	var identifiers []string
	strippedContainerID := util.ExtractLastFragment(container.ContainerID)
	// in the case where the container could not be started due to some error
	if len(strippedContainerID) > 0 {
		identifier := ""
		if len(nodeIdentifier) > 0 {
			identifier = fmt.Sprintf("%s:%s", nodeIdentifier, strippedContainerID)
		} else {
			identifier = strippedContainerID
		}
		identifiers = []string{
			fmt.Sprintf("urn:container:/%s", identifier),
		}
	}

	log.Tracef("Created identifiers for %s: %v", container.Name, identifiers)

	containerExternalID := cc.buildContainerExternalID(pod.Namespace, pod.Name, container.Name)

	tags := cc.initTags(metav1.ObjectMeta{Namespace: pod.Namespace}, metav1.TypeMeta{Kind: "container"})

	data := map[string]interface{}{
		"name": container.Name,
		"docker": map[string]interface{}{
			"image":       container.Image,
			"imageId":     container.ImageID,
			"containerId": strippedContainerID,
		},
		"pod":          pod.Name,
		"podIP":        pod.PodIP,
		"podPhase":     pod.Phase,
		"restartCount": container.RestartCount,
		"tags":         tags,
	}

	if container.State.Running != nil {
		data["startTime"] = container.State.Running.StartedAt
	}

	if container.State.Terminated != nil {
		data["exitCode"] = container.State.Terminated.ExitCode
	} else if container.LastTerminationState.Terminated != nil {
		data["exitCode"] = container.LastTerminationState.Terminated.ExitCode
	}

	if containerPort.ContainerPort != 0 {
		data["containerPort"] = containerPort.ContainerPort
	}

	if containerPort.HostPort != 0 {
		data["hostPort"] = containerPort.HostPort
	}

	if len(identifiers) > 0 {
		data["identifiers"] = identifiers
	}

	component := &topology.Component{
		ExternalID: containerExternalID,
		Type:       topology.Type{Name: "container"},
		Data:       data,
	}

	log.Tracef("Created StackState container component %s: %v", containerExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Pod to Kubernetes / OpenShift Container relation
func (cc *ContainerCorrelator) podToContainerStackStateRelation(podExternalID, containerExternalID string) *topology.Relation {
	log.Tracef("Mapping kubernetes pod to container relation: %s -> %s", podExternalID, containerExternalID)

	relation := cc.CreateRelation(podExternalID, containerExternalID, "encloses")

	log.Tracef("Created StackState pod -> container relation %s -> %s", relation.SourceID, relation.TargetID)

	return relation
}

// Creates a StackState relation from a Container to Kubernetes / OpenShift Node relation
func (cc *ContainerCorrelator) containerToNodeStackStateRelation(containerExternalID, nodeIdentifier string) *topology.Relation {
	log.Tracef("Mapping kubernetes container to node relation: %s -> %s", containerExternalID, nodeIdentifier)

	relation := cc.CreateRelation(containerExternalID, nodeIdentifier, "runs_on")

	log.Tracef("Created StackState container -> node relation %s -> %s", relation.SourceID, relation.TargetID)

	return relation
}
