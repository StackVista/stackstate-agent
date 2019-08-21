// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
// +build kubeapiserver

package kubeapi

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	core "github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"
)

const (
	kubernetesAPITopologyCheckName = "kubernetes_api_topology"
)

// TopologyConfig is the config of the API server.
type TopologyConfig struct {
	CollectTopology bool `yaml:"collect_topology"`
	CheckID         check.ID
	Instance        topology.Instance
}

// TopologyCheck grabs events from the API server.
type TopologyCheck struct {
	CommonCheck
	instance *TopologyConfig
}

func (c *TopologyConfig) parse(data []byte) error {
	// default values
	c.CollectTopology = config.Datadog.GetBool("collect_kubernetes_topology")

	return yaml.Unmarshal(data, c)
}

// Configure parses the check configuration and init the check.
func (t *TopologyCheck) Configure(config, initConfig integration.Data) error {
	err := t.ConfigureKubeApiCheck(config)
	if err != nil {
		return err
	}

	err = t.instance.parse(config)
	if err != nil {
		_ = log.Error("could not parse the config for the API topology check")
		return err
	}

	log.Debugf("Running config %s", config)
	return nil
}

// Run executes the check.
func (t *TopologyCheck) Run() error {
	// initialize kube api check
	err := t.InitKubeApiCheck()
	if err == apiserver.ErrNotLeader {
		log.Debug("Agent is not leader, will not run the check")
		return nil
	} else if err != nil {
		return err
	}

	// Running the event collection.
	if !t.instance.CollectTopology {
		return nil
	}

	// set the check "instance id" for snapshots
	t.instance.CheckID = kubernetesAPITopologyCheckName
	t.instance.Instance = topology.Instance{Type: "kubernetes", URL: t.KubeAPIServerHostname}

	// start the topology snapshot with the batch-er
	batcher.GetBatcher().SubmitStartSnapshot(t.instance.CheckID, t.instance.Instance)

	// get all the nodes
	_, _ = t.getAllNodes()

	// get all the pods
	_, _ = t.getAllPods()

	// get all the services
	_, _ = t.getAllServices()

	// get all the containers
	batcher.GetBatcher().SubmitStopSnapshot(t.instance.CheckID, t.instance.Instance)
	batcher.GetBatcher().SubmitComplete(t.instance.CheckID)

	return nil
}

// get all the nodes in the k8s cluster
func (t *TopologyCheck) getAllNodes() ([]*topology.Component, error) {
	nodes, err := t.ac.GetNodes()
	if err != nil {
		return nil, err
	}

	components := make([]*topology.Component, 0)

	for _, node := range nodes {
		// creates and publishes StackState node component
		component := t.mapAndSubmitNode(node)
		components = append(components, &component)
	}

	return components, nil
}

// get all the pods in the k8s cluster
func (t *TopologyCheck) getAllPods() ([]*topology.Component, error) {
	pods, err := t.ac.GetPods()
	if err != nil {
		return nil, err
	}

	components := make([]*topology.Component, 0)

	for _, pod := range pods {
		// creates and publishes StackState pod component with relations
		component := t.mapAndSubmitPodWithRelations(pod)
		components = append(components, &component)
	}

	return components, nil
}

// get all the services in the k8s cluster
func (t *TopologyCheck) getAllServices() ([]*topology.Component, error) {
	services, err := t.ac.GetServices()
	if err != nil {
		return nil, err
	}

	components := make([]*topology.Component, 0)

	for _, svc := range services {
		// creates a StackState component for the kubernetes pod
		log.Tracef("Mapping kubernetes service to StackState Component: %s", svc.String())

	}

	return components, nil
}

// Map and Submit the Kubernetes Node into a StackState component
func (t *TopologyCheck) mapAndSubmitNode(node v1.Node) topology.Component {

	// submit the StackState component for publishing to StackState
	component := nodeToStackStateComponent(node)
	log.Tracef("Publishing StackState node component for %s: %v", component.ExternalID, component.JSONString())
	batcher.GetBatcher().SubmitComponent(t.instance.CheckID, t.instance.Instance, component)

	return component
}

// Map and Submit the Kubernetes Pod into a StackState component
func (t *TopologyCheck) mapAndSubmitPodWithRelations(pod v1.Pod) topology.Component {
	// submit the StackState component for publishing to StackState
	podComponent := podToStackStateComponent(pod)
	log.Tracef("Publishing StackState pod component for %s: %v", podComponent.ExternalID, podComponent.JSONString())
	batcher.GetBatcher().SubmitComponent(t.instance.CheckID, t.instance.Instance, podComponent)

	// creates a StackState relation for the kubernetes node -> pod
	relation := podToNodeStackStateRelation(pod)
	log.Tracef("Publishing StackState pod -> node relation %s->%s", relation.SourceID, relation.TargetID)
	batcher.GetBatcher().SubmitRelation(t.instance.CheckID, t.instance.Instance, relation)

	// creates a StackState component for the kubernetes pod containers + relation to pod
	for _, container := range pod.Status.ContainerStatuses {

		// submit the StackState component for publishing to StackState
		containerComponent := containerToStackStateComponent(pod, container)
		log.Tracef("Publishing StackState container component for %s: %v", containerComponent.ExternalID, containerComponent.JSONString())
		batcher.GetBatcher().SubmitComponent(t.instance.CheckID, t.instance.Instance, containerComponent)

		// create the relation between the container and pod
		relation := containerToPodStackStateRelation(containerComponent.ExternalID, podComponent.ExternalID)
		log.Tracef("Publishing StackState container -> pod relation %s->%s", relation.SourceID, relation.TargetID)
		batcher.GetBatcher().SubmitRelation(t.instance.CheckID, t.instance.Instance, relation)
	}

	return podComponent
}

// Creates a StackState component from a Kubernetes Node
func nodeToStackStateComponent(node v1.Node) topology.Component {
	// creates a StackState component for the kubernetes node
	log.Tracef("Mapping kubernetes node to StackState component: %s", node.String())

	// create identifier list to merge with StackState components
	identifiers := make([]string, 0)
	for _, address := range node.Status.Addresses {
		switch addressType := address.Type; addressType {
		case v1.NodeInternalIP:
			identifiers = append(identifiers, fmt.Sprintf("urn:ip:/%s:%s:%s", node.ClusterName, node.Name, address.Address))
		case v1.NodeExternalIP:
			identifiers = append(identifiers, fmt.Sprintf("urn:ip:/%s:%s", node.ClusterName, address.Address))
		case v1.NodeHostName:
			identifiers = append(identifiers, fmt.Sprintf("urn:host:/%s", address.Address))
		default:
			continue
		}
	}

	log.Tracef("Created identifiers for %s: %v", node.Name, identifiers)

	nodeExternalID := buildNodeExternalId(node.ClusterName, node.Name)

	// clear out the unnecessary status array values
	nodeStatus := node.Status
	nodeStatus.Conditions = make([]v1.NodeCondition, 0)
	nodeStatus.Images = make([]v1.ContainerImage, 0)

	component := topology.Component{
		ExternalID: nodeExternalID,
		Type:       topology.Type{Name: "kubernetes-node"},
		Data: map[string]interface{}{
			"name":              node.Name,
			"kind":              node.Kind,
			"creationTimestamp": node.CreationTimestamp,
			"tags":              node.Labels,
			"status":            nodeStatus,
			"namespace":         node.Namespace,
			"identifiers":       identifiers,
			//"taints": node.Spec.Taints,
		},
	}

	log.Tracef("Created StackState node component %s: %v", nodeExternalID, component.JSONString())

	return component
}

// Creates a StackState component from a Kubernetes Pod
func podToStackStateComponent(pod v1.Pod) topology.Component {
	// creates a StackState component for the kubernetes pod
	log.Tracef("Mapping kubernetes pod to StackState Component: %s", pod.String())

	// create identifier list to merge with StackState components
	identifiers := []string{
		fmt.Sprintf("urn:ip:/%s:%s", pod.ClusterName, pod.Status.PodIP),
	}
	log.Tracef("Created identifiers for %s: %v", pod.Name, identifiers)

	podExternalID := buildPodExternalId(pod.ClusterName, pod.Name)

	// clear out the unnecessary status array values
	podStatus := pod.Status
	podStatus.Conditions = make([]v1.PodCondition, 0)
	podStatus.ContainerStatuses = make([]v1.ContainerStatus, 0)

	component := topology.Component{
		ExternalID: podExternalID,
		Type:       topology.Type{Name: "kubernetes-pod"},
		Data: map[string]interface{}{
			"name":              pod.Name,
			"kind":              pod.Kind,
			"creationTimestamp": pod.CreationTimestamp,
			"tags":              pod.Labels,
			"status":            podStatus,
			"namespace":         pod.Namespace,
			//"tolerations": pod.Spec.Tolerations,
			"restartPolicy": pod.Spec.RestartPolicy,
			"identifiers":   identifiers,
		},
	}

	log.Tracef("Created StackState pod component %s: %v", podExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes Pod to Node relation
func podToNodeStackStateRelation(pod v1.Pod) topology.Relation {
	podExternalID := buildPodExternalId(pod.ClusterName, pod.Name)
	nodeExternalID := buildNodeExternalId(pod.ClusterName, pod.Spec.NodeName)

	log.Tracef("Mapping kubernetes pod to node relation: %s -> %s", podExternalID, nodeExternalID)

	relation := topology.Relation{
		ExternalID: fmt.Sprintf("%s->%s", podExternalID, nodeExternalID),
		SourceID:   podExternalID,
		TargetID:   nodeExternalID,
		Type:       topology.Type{Name: "scheduled_on"},
		Data:       map[string]interface{}{},
	}

	log.Tracef("Created StackState pod -> node relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}

// Creates a StackState component from a Kubernetes Pod Container
func containerToStackStateComponent(pod v1.Pod, container v1.ContainerStatus) topology.Component {
	log.Tracef("Mapping kubernetes pod container to StackState component: %s", container.String())
	// create identifier list to merge with StackState components
	identifiers := []string{
		fmt.Sprintf("urn:container:/%s", container.ContainerID),
	}
	log.Tracef("Created identifiers for %s: %v", container.Name, identifiers)

	containerExternalID := buildContainerExternalId(pod.ClusterName, pod.Name, container.Name)

	data := map[string]interface{}{
		"name": container.Name,
		"docker": map[string]interface{}{
			"image":        container.Image,
			"container_id": container.ContainerID,
		},
		"restartCount": container.RestartCount,
		"identifiers":  identifiers,
	}

	if container.State.Running != nil {
		data["startTime"] = container.State.Running.StartedAt
	}

	component := topology.Component{
		ExternalID: containerExternalID,
		Type:       topology.Type{Name: "kubernetes-container"},
		Data:       data,
	}

	log.Tracef("Created StackState container component %s: %v", containerExternalID, component.JSONString())

	return component
}

// Creates a StackState relation from a Kubernetes Container to Pod relation
func containerToPodStackStateRelation(containerExternalID, podExternalID string) topology.Relation {
	log.Tracef("Mapping kubernetes container to pod relation: %s -> %s", containerExternalID, podExternalID)

	relation := topology.Relation{
		ExternalID: fmt.Sprintf("%s->%s", containerExternalID, podExternalID),
		SourceID:   containerExternalID,
		TargetID:   podExternalID,
		Type:       topology.Type{Name: "scheduled_on"},
		Data:       map[string]interface{}{},
	}

	log.Tracef("Created StackState container -> pod relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}

func buildNodeExternalId(clusterName, nodeName string) string {
	return fmt.Sprintf("urn:/kubernetes:%s:node:%s", clusterName, nodeName)
}

func buildPodExternalId(clusterName, podName string) string {
	return fmt.Sprintf("urn:/kubernetes:%s:pod:%s", clusterName, podName)
}

func buildContainerExternalId(clusterName, podName, containerName string) string {
	return fmt.Sprintf("urn:/kubernetes:%s:pod:%s:container:%s", clusterName, podName, containerName)
}

func buildServiceExternalId(clusterName, serviceID string) string {
	return fmt.Sprintf("urn:/kubernetes:%s:service:%s", clusterName, serviceID)
}

// KubernetesASFactory is exported for integration testing.
func KubernetesApiTopologyFactory() check.Check {
	return &TopologyCheck{
		CommonCheck: CommonCheck{
			CheckBase: core.NewCheckBase(kubernetesAPITopologyCheckName),
		},
		instance: &TopologyConfig{},
	}
}

func init() {
	core.RegisterCheck(kubernetesAPITopologyCheckName, KubernetesApiTopologyFactory)
}
