//go:build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/hostname"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeCollector implements the ClusterTopologyCollector interface.
type NodeCollector struct {
	NodeIdentifierCorrChan chan<- *NodeIdentifierCorrelation
	ClusterTopologyCollector
}

// NodeStatus is the StackState representation of a Kubernetes / Openshift Node Status
type NodeStatus struct {
	Phase           v1.NodePhase      `json:"phase,omitempty"`
	NodeInfo        v1.NodeSystemInfo `json:"nodeInfo,omitempty"`
	KubeletEndpoint v1.DaemonEndpoint `json:"kubeletEndpoint,omitempty"`
}

// NewNodeCollector
func NewNodeCollector(
	nodeIdentifierCorrChan chan<- *NodeIdentifierCorrelation, clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &NodeCollector{
		NodeIdentifierCorrChan:   nodeIdentifierCorrChan,
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*NodeCollector) GetName() string {
	return "Node Collector"
}

// Collects and Publishes the Node Components
func (nc *NodeCollector) CollectorFunction() error {
	// get all the nodes in the cluster
	nodes, err := nc.GetAPIClient().GetNodes()
	if err != nil {
		return err
	}

	for _, node := range nodes {
		// creates and publishes StackState node component
		component, nodeIdentifier := nc.nodeToStackStateComponent(node)
		// creates a StackState relation for the cluster node -> cluster
		relation := nc.nodeToClusterStackStateRelation(node)

		nc.SubmitComponent(component)
		nc.SubmitRelation(relation)

		// send the node identifier to be correlated
		nc.NodeIdentifierCorrChan <- &NodeIdentifierCorrelation{node.Name, nodeIdentifier, component.ExternalID}
	}

	close(nc.NodeIdentifierCorrChan)

	return nil
}

// Creates a StackState component from a Kubernetes Node
func (nc *NodeCollector) nodeToStackStateComponent(node v1.Node) (*topology.Component, string) {
	// creates a StackState component for the kubernetes node
	log.Tracef("Mapping kubernetes node to StackState component: %s", node.String())

	identifiers := nc.GetURNBuilder().BuildNodeURNs(node)
	log.Debugf("Created identifiers for %s: %v", node.Name, identifiers)

	nodeExternalID := nc.buildNodeExternalID(node.Name)

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := nc.initTags(node.ObjectMeta, metav1.TypeMeta{Kind: "Node"})

	hostname, err := hostname.GetHostname(node)
	if err != nil {
		hostname = node.Name
	}

	component := &topology.Component{
		ExternalID: nodeExternalID,
		Type:       topology.Type{Name: "node"},
		Data: map[string]interface{}{
			"name":        node.Name,
			"tags":        tags,
			"identifiers": identifiers,
			// instanceId is deprecated because it is an AWS specific term
			// For backward compatibility with K8s/OpenShift stackpack
			// it is still included in the data
			// this should be replaced in the stackpack with `sts_host` which is the name that is used
			// in the metric labels for example
			"instanceId": hostname,
			"sts_host":   hostname,
		},
	}

	if nc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if nc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&node)
		} else {
			sourceProperties = makeSourceProperties(&node)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("creationTimestamp", node.CreationTimestamp)
		component.Data.PutNonEmpty("uid", node.UID)
		component.Data.PutNonEmpty("generateName", node.GenerateName)
		component.Data.PutNonEmpty("kind", node.Kind)
		component.Data.PutNonEmpty("status", NodeStatus{
			Phase:           node.Status.Phase,
			NodeInfo:        node.Status.NodeInfo,
			KubeletEndpoint: node.Status.DaemonEndpoints.KubeletEndpoint,
		})
	}

	log.Tracef("Created StackState node component %s: %v", nodeExternalID, component.JSONString())

	return component, hostname
}

// Creates a StackState relation from a Kubernetes Pod to Node relation
func (nc *NodeCollector) nodeToClusterStackStateRelation(node v1.Node) *topology.Relation {
	nodeExternalID := nc.buildNodeExternalID(node.Name)
	clusterExternalID := nc.buildClusterExternalID()

	log.Tracef("Mapping kubernetes node to cluster relation: %s -> %s", nodeExternalID, clusterExternalID)

	relation := nc.CreateRelation(nodeExternalID, clusterExternalID, "belongs_to")

	log.Tracef("Created StackState node -> cluster relation %s->%s", relation.SourceID, relation.TargetID)

	return relation
}
