//go:build kubeapiserver

package topologycollectors

import (
	"errors"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterCollector implements the ClusterTopologyCollector interface.
type ClusterCollector struct {
	ClusterTopologyCollector
}

// NewClusterTopologyCollector
func NewClusterCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &ClusterCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*ClusterCollector) GetName() string {
	return "Cluster Collector"
}

// Collects and Published the Cluster Component
func (cc *ClusterCollector) CollectorFunction() error {
	if cc.GetInstance().Type == "" || cc.GetInstance().URL == "" {
		return errors.New("cluster name or cluster instance type could not be detected, " +
			"therefore we are unable to create the cluster component")
	}

	cc.SubmitComponent(cc.clusterToStackStateComponent())
	return nil
}

// Creates a StackState component from a Kubernetes / OpenShift Cluster
func (cc *ClusterCollector) clusterToStackStateComponent() *topology.Component {
	clusterExternalID := cc.buildClusterExternalID()

	tags := cc.initTags(metav1.ObjectMeta{}, metav1.TypeMeta{Kind: "Cluster"})

	component := &topology.Component{
		ExternalID: clusterExternalID,
		Type:       topology.Type{Name: "cluster"},
		Data: map[string]interface{}{
			"name": cc.GetInstance().URL,
			"tags": tags,
		},
	}

	log.Tracef("Created StackState cluster component %s: %v", clusterExternalID, component.JSONString())

	return component
}
