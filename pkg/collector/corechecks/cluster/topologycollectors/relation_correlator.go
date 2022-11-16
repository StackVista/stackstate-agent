//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// RelationCorrelator ContainerCorrelator implements the ClusterTopologyCollector interface.
type RelationCorrelator struct {
	ComponentIDChannel     chan string
	RelationCorrChan       chan *topology.Relation
	RelationChan           chan<- *topology.Relation
	CollectorsFinishedChan <-chan bool
	ClusterTopologyCorrelator
}

// NewRelationCorrelator creates a RelationCorrelator
func NewRelationCorrelator(componentIDChannel chan string, relationCorrChannel chan *topology.Relation,
	relationChannel chan<- *topology.Relation,
	collectorsFinishedChan chan bool,
	clusterTopologyCorrelator ClusterTopologyCorrelator) ClusterTopologyCorrelator {
	return &RelationCorrelator{
		ComponentIDChannel:        componentIDChannel,
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
	componentIDs := make(map[string]struct{})
	var possibleRelations []*topology.Relation
loop:
	for {
		select {
		case id := <-rc.ComponentIDChannel:
			componentIDs[id] = struct{}{}
		case relation := <-rc.RelationCorrChan:
			possibleRelations = append(possibleRelations, relation)
		case <-rc.CollectorsFinishedChan:
			break loop
		}
	}

	for _, relation := range possibleRelations {
		_, sourceExists := componentIDs[relation.SourceID]
		_, targetExists := componentIDs[relation.TargetID]
		if sourceExists && targetExists {
			// TODO remove debug
			log.Debugf("Created relation '%s'", relation.ExternalID)
			rc.RelationChan <- relation
		} else {
			if !sourceExists {
				log.Debugf("Ignoring relation '%s' because source does not exist", relation.ExternalID)
			} else {
				log.Debugf("Ignoring relation '%s' because target does not exist", relation.ExternalID)
			}
		}
	}

	return nil
}
