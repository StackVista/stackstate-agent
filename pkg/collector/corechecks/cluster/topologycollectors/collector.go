//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"errors"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

const (
	Deployment  = "Deployment"
	DaemonSet   = "DaemonSet"
	StatefulSet = "StatefulSet"
	ReplicaSet  = "ReplicaSet"
	CronJob     = "CronJob"
	Job         = "Job"
)

// ClusterTopologyCollector collects cluster components and relations.
type ClusterTopologyCollector interface {
	CollectorFunction() error
	SubmitComponent(component *topology.Component)
	ClusterTopologyCommon
}

type clusterTopologyCollector struct {
	ComponentChan   chan<- *topology.Component
	ComponentIdChan chan<- string
	ClusterTopologyCommon
}

// NewClusterTopologyCollector
func NewClusterTopologyCollector(
	ComponentChan chan<- *topology.Component,
	ComponentIdChan chan<- string,
	clusterTopologyCommon ClusterTopologyCommon,
) ClusterTopologyCollector {
	return &clusterTopologyCollector{
		ComponentChan,
		ComponentIdChan,
		clusterTopologyCommon,
	}
}

// CollectorFunction
func (c *clusterTopologyCollector) CollectorFunction() error {
	return errors.New("CollectorFunction NotImplemented")
}

func (c *clusterTopologyCollector) SubmitComponent(component *topology.Component) {
	c.ComponentChan <- component
	c.ComponentIdChan <- component.ExternalID
}
