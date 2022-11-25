//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// RelationCorrelator implements the ClusterTopologyCollector interface.
type RelationCorrelator struct {
	RelationChan       chan<- *topology.Relation
	CollectorsDoneChan <-chan bool
	ClusterTopologyCorrelator
}

// NewRelationCorrelator creates a RelationCorrelator
func NewRelationCorrelator(
	relationChannel chan<- *topology.Relation,
	collectorsDoneChan chan bool,
	clusterTopologyCorrelator ClusterTopologyCorrelator) ClusterTopologyCorrelator {
	return &RelationCorrelator{
		RelationChan:              relationChannel,
		CollectorsDoneChan:        collectorsDoneChan,
		ClusterTopologyCorrelator: clusterTopologyCorrelator,
	}
}

// GetName returns the name of the Collector
func (*RelationCorrelator) GetName() string {
	return "Container Correlator"
}

// CorrelateFunction Collects and publishes relations where both source and target exist
func (rc *RelationCorrelator) CorrelateFunction() error {
	// wait until all collectors are done
	<-rc.CollectorsDoneChan

	componentCache := rc.GetComponentIDCache()
	for _, relation := range rc.GetPossibleRelations() {
		_, sourceExists := componentCache[relation.SourceID]
		_, targetExists := componentCache[relation.TargetID]
		if sourceExists && targetExists {
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
