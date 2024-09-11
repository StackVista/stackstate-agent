package disk

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/collector/python"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/health"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMakeTopologyCollector(t *testing.T) {
	dtc := MakeTopologyCollector()
	assert.Equal(t, checkid.ID("disk_topology"), dtc.CheckID)
	expectedInstance := topology.Instance{
		Type: "disk",
		URL:  "agents",
	}
	assert.Equal(t, expectedInstance, dtc.TopologyInstance)
}

func TestDiskTopologyCollector_createComponent(t *testing.T) {
	dtc := MakeTopologyCollector()
	testHostname := "test-hostname"
	testIdentifiers := []string{"urn:azure:/subscriptions/bla/bla/bla/virtualMachine/bla"}
	partitions := []disk.PartitionStat{
		{
			Device: "abcd",
		},
		{
			Device: "1234",
		},
		{
			Device: "ecdf",
		},
		{
			Device: "my/device/path",
		},
		{
			Device: "1234",
		},
		{
			Device: "abcd",
		},
	}
	diskComponent := dtc.createDiskComponent(testHostname, testIdentifiers, partitions)
	assert.Equal(t, fmt.Sprintf("urn:host:/%s", testHostname), diskComponent.ExternalID)
	assert.Equal(t, topology.Type{Name: "host"}, diskComponent.Type)
	expectedData := topology.Data{
		"host":        testHostname,
		"identifiers": testIdentifiers,
		"devices":     []string{"abcd", "1234", "ecdf", "my/device/path"},
	}
	assert.Equal(t, expectedData, diskComponent.Data)
}

func TestDiskTopologyCollector_BuildTopology(t *testing.T) {
	// set up the mock batcher
	mockBatcher, _, _, checkManager := python.SetupTransactionalComponents()
	// set mock hostname
	testHostname := "test-hostname"
	config.Datadog.SetWithoutSource("hostname", testHostname)

	dtc := MakeTopologyCollector()
	partitions := []disk.PartitionStat{
		{
			Device: "abcd",
		},
		{
			Device: "1234",
		},
		{
			Device: "ecdf",
		},
		{
			Device: "my/device/path",
		},
		{
			Device: "1234",
		},
		{
			Device: "abcd",
		},
	}

	err := dtc.BuildTopology(partitions, checkManager.GetCheckHandler(diskCheckID))
	assert.NoError(t, err)

	producedTopology := mockBatcher.CollectedTopology.Flush()
	expectedTopology := batcher.CheckInstanceBatchStates(map[checkid.ID]batcher.CheckInstanceBatchState{
		"disk_topology": {
			Health: make(map[string]health.Health),
			Topology: &topology.Topology{
				StartSnapshot: false,
				StopSnapshot:  false,
				Instance:      topology.Instance{Type: "disk", URL: "agents"},
				Components: []topology.Component{
					{
						ExternalID: fmt.Sprintf("urn:host:/%s", testHostname),
						Type: topology.Type{
							Name: "host",
						},
						Data: topology.Data{
							"host":        testHostname,
							"devices":     []string{"abcd", "1234", "ecdf", "my/device/path"},
							"identifiers": []string{},
						},
					},
				},
				Relations: []topology.Relation{},
				DeleteIDs: []string{},
			},
		},
	})

	assert.Equal(t, expectedTopology, producedTopology)
}
