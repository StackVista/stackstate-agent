package corechecks

import (
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMakeCheckTopologyCollector(t *testing.T) {
	checkID := checkid.ID("process_check_topology")
	instance := topology.Instance{
		Type: "test",
		URL:  "url",
	}
	ptc := MakeCheckTopologyCollector(checkID, instance)
	assert.Equal(t, checkID, ptc.CheckID)
	assert.Equal(t, instance, ptc.TopologyInstance)
}

func TestMakeCheckProcessTopologyCollector(t *testing.T) {
	checkID := checkid.ID("process_check_topology")
	ptc := MakeCheckProcessTopologyCollector(checkID)
	assert.Equal(t, checkID, ptc.CheckID)
	expectedInstance := topology.Instance{
		Type: "process",
		URL:  "agents",
	}
	assert.Equal(t, expectedInstance, ptc.TopologyInstance)
}
