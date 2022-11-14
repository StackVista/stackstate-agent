//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
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

// NewContainerCorrelator
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

// Collects and Published the Cluster Component
func (rc *RelationCorrelator) CorrelateFunction1() error {
	fmt.Println("RelationCorrelator start")
	//defer close(rc.RelationCorrChan)
	componentIds := make(map[string]struct{})
	for id := range rc.ComponentIdChan {
		componentIds[id] = struct{}{}
	}

	fmt.Println("RelationCorrelator after comp ids")

	for relation := range rc.RelationCorrChan {
		_, sourceExists := componentIds[relation.SourceID]
		_, targetExists := componentIds[relation.TargetID]

		if sourceExists && targetExists {
			rc.RelationChan <- relation
			fmt.Println("Sending relation")
		}
	}
	fmt.Println("RelationCorrelator after relations")

	return nil
}

// Collects and Published the Cluster Component
func (rc *RelationCorrelator) CorrelateFunction() error {
	defer close(rc.ComponentIdChan)
	defer close(rc.RelationCorrChan)
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

	fmt.Println("RelationCorrelator after comp ids")

	for _, relation := range possibleRelations {
		_, sourceExists := componentIds[relation.SourceID]
		_, targetExists := componentIds[relation.TargetID]
		if sourceExists && targetExists {
			rc.RelationChan <- relation
		}
	}
	fmt.Println("RelationCorrelator after relations")

	return nil
}
