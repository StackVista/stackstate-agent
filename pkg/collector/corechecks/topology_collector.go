package corechecks

import (
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
)

// CheckTopologyCollector contains all the metadata needed to produce disk topology
type CheckTopologyCollector struct {
	CheckID          checkid.ID
	TopologyInstance topology.Instance
}

// MakeCheckProcessTopologyCollector returns an instance of the CheckTopologyCollector
func MakeCheckProcessTopologyCollector(checkID checkid.ID) CheckTopologyCollector {
	return CheckTopologyCollector{
		CheckID: checkID,
		TopologyInstance: topology.Instance{
			Type: "process",
			URL:  "agents",
		},
	}
}

// MakeCheckTopologyCollector returns an instance of the CheckTopologyCollector
func MakeCheckTopologyCollector(checkID checkid.ID, instance topology.Instance) CheckTopologyCollector {
	return CheckTopologyCollector{
		CheckID:          checkID,
		TopologyInstance: instance,
	}
}
