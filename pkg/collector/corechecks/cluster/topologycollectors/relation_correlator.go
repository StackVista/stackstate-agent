//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

// ContainerCorrelator implements the ClusterTopologyCollector interface.
type RelationCorrelator struct {
	ComponentIdChan  <-chan string
	RelationCorrChan <-chan *topology.Relation
	RelationChan     chan<- *topology.Relation
	ClusterTopologyCorrelator
}

// NewContainerCorrelator
func NewRelationCorrelator(componentIdChannel <-chan string, relationCorrChannel <-chan *topology.Relation,
	relationChannel chan<- *topology.Relation,
	clusterTopologyCorrelator ClusterTopologyCorrelator) ClusterTopologyCorrelator {
	return &RelationCorrelator{
		ComponentIdChan:           componentIdChannel,
		RelationCorrChan:          relationCorrChannel,
		RelationChan:              relationChannel,
		ClusterTopologyCorrelator: clusterTopologyCorrelator,
	}
}

// GetName returns the name of the Collector
func (*RelationCorrelator) GetName() string {
	return "Container Correlator"
}

// Collects and Published the Cluster Component
func (rc *RelationCorrelator) CorrelateFunction() error {
	componentIds := make(map[string]struct{})
	// map containers that require the Node instanceId
	for id := range rc.ComponentIdChan {
		componentIds[id] = struct{}{}
	}

	for relation := range rc.RelationCorrChan {
		_, sourceExists := componentIds[relation.SourceID]
		_, targetExists := componentIds[relation.TargetID]
		if sourceExists && targetExists {
			rc.RelationChan <- relation
		}
	}

	return nil
}
