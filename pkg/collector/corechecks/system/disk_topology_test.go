package system

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMakeTopologyCollector(t *testing.T) {
	dtc := MakeTopologyCollector()
	assert.Equal(t, check.ID("disk_topology"), dtc.CheckID)
	expectedInstance := topology.Instance{
		Type: "disk",
		URL:  "agents",
	}
	assert.Equal(t, expectedInstance, dtc.TopologyInstance)
}

func TestDiskTopologyCollector_createComponent(t *testing.T) {
	dtc := MakeTopologyCollector()
	testHostname := "test-hostname"
	partitions := []disk.PartitionStat{
		{
			Device:     "abcd",
		},
		{
			Device:     "1234",
		},
		{
			Device:     "ecdf",
		},
		{
			Device:     "my/device/path",
		},
		{
			Device:     "1234",
		},
		{
			Device:     "abcd",
		},
	}
	diskComponent := dtc.createDiskComponent(testHostname, partitions)
	assert.Equal(t, fmt.Sprintf("urn:host:/%s", testHostname), diskComponent.ExternalID)
	assert.Equal(t, topology.Type(topology.Type{Name:"host"}), diskComponent.Type)
	expectedData := topology.Data{
		"devices": []string{"abcd", "1234", "ecdf", "my/device/path"},
	}
	assert.Equal(t, expectedData, diskComponent.Data)
}

func TestDiskTopologyCollector_BuildTopology(t *testing.T) {
	// set up the mock batcher
	mockBatcher := batcher.NewMockBatcher()

	dtc := MakeTopologyCollector()
	partitions := []disk.PartitionStat{
		{
			Device:     "abcd",
		},
		{
			Device:     "1234",
		},
		{
			Device:     "ecdf",
		},
		{
			Device:     "my/device/path",
		},
		{
			Device:     "1234",
		},
		{
			Device:     "abcd",
		},
	}

	err := dtc.BuildTopology(partitions)
	assert.NoError(t, err)

	producedTopology := mockBatcher.CollectedTopology.Flush()
	expectedTopology := batcher.Topologies{
		"disk_topology": {
			StartSnapshot: false,
			StopSnapshot:  false,
			Instance:      topology.Instance{Type: "disk", URL:  "agents",},
			Components:    nil,
			Relations:     nil,
		},
	}

	assert.Equal(t, expectedData, producedTopology)
}


