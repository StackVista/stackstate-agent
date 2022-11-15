//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

// ContainerCorrelator implements the ClusterTopologyCollector interface.
type RelationCorrelator struct {
	ComponentIdChan        chan string
	RelationCorrChan       chan *topology.Relation
	RelationChan           chan<- *topology.Relation
	CollectorsFinishedChan <-chan bool
	ClusterTopologyCorrelator
}

// NewContainerCorrelator creates a RelationCorrelator
func NewRelationCorrelator(componentIdChannel chan string, relationCorrChannel chan *topology.Relation,
	relationChannel chan<- *topology.Relation,
	collectorsFinishedChan chan bool,
	clusterTopologyCorrelator ClusterTopologyCorrelator) ClusterTopologyCorrelator {
	return &RelationCorrelator{
		ComponentIdChan:           componentIdChannel,
		RelationCorrChan:          relationCorrChannel,
		RelationChan:              relationChannel,
		CollectorsFinishedChan:    collectorsFinishedChan,
		ClusterTopologyCorrelator: clusterTopologyCorrelator,
	}
}

// GetName returns the name of the Collector
func (*RelationCorrelator) GetName() string {
	return "Container Correlator"
}

// CorrelateFunction Collects and publishes relations where both source and target exist
func (rc *RelationCorrelator) CorrelateFunction() error {
	componentIds := make(map[string]struct{})
	var possibleRelations []*topology.Relation
loop:
	for {
		select {
		case id := <-rc.ComponentIdChan:
			componentIds[id] = struct{}{}
		case relation := <-rc.RelationCorrChan:
			possibleRelations = append(possibleRelations, relation)
		case <-rc.CollectorsFinishedChan:
			break loop
		}
	}

	for _, relation := range possibleRelations {
		_, sourceExists := componentIds[relation.SourceID]
		_, targetExists := componentIds[relation.TargetID]
		if sourceExists && targetExists {
			rc.RelationChan <- relation
		}
	}

	return nil
}
