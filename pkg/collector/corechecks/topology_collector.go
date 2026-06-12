package corechecks

import (
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
)

// CheckTopologyCollector contains all the metadata needed to produce disk topology
type CheckTopologyCollector struct {
	TopologyInstance topology.Instance
}

// MakeCheckTopologyCollector returns an instance of the CheckTopologyCollector
func MakeCheckTopologyCollector(instance topology.Instance) CheckTopologyCollector {
	return CheckTopologyCollector{
		TopologyInstance: instance,
	}
}
