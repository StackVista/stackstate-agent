//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/urn"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"k8s.io/api/core/v1"
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

	tags := nc.initTags(node.ObjectMeta)

	instanceID := urn.GetInstanceID(node)
	component := &topology.Component{
		ExternalID: nodeExternalID,
		Type:       topology.Type{Name: "node"},
		Data: map[string]interface{}{
			"name":        node.Name,
			"tags":        tags,
			"identifiers": identifiers,
			// for backward compatibility with K8s/OpenShift stackpack
			// we specify instanceId in data even if it's also in the sourceProperties
			"instanceId": instanceID,
		},
	}

	if nc.IsSourcePropertiesFeatureEnabled() || nc.IsExposeKubernetesStatusEnabled() {
		component.SourceProperties = if nc.IsExposeKubernetesStatusEnabled() {
			makeSourcePropertiesFullDetails(&node)
		} else {
			makeSourceProperties(&node)
		}
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

	return component, instanceID
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
