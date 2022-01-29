// StackState
package topology

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	cspec "github.com/StackVista/stackstate-agent/pkg/collector/corechecks/containers/spec"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockUtil struct {
}

func (m MockUtil) GetContainers() ([]*cspec.Container, error) {
	return []*cspec.Container{
		{
			Name:    "container1",
			Runtime: "containerd",
			ID:      "containerId1",
			Image:   "image1",
			Mounts: []specs.Mount{
				{Source: "source1", Destination: "dest1"},
				{Source: "source2", Destination: "dest2"},
			},
			State: "running",
		},
		{
			Name:    "container2",
			Runtime: "docker",
			ID:      "containerId2",
			Image:   "image2",
			Mounts: []specs.Mount{
				{Source: "source1", Destination: "dest1"},
				{Source: "source2", Destination: "dest2"},
			},
			State: "running",
		},
	}, nil
}

func TestMakeContainerTopologyCollector(t *testing.T) {
	hostname, err := util.GetHostname()
	assert.NoError(t, err)
	assert.Equal(t, &ContainerTopologyCollector{
		CheckTopologyCollector: corechecks.MakeCheckTopologyCollector("container_topology", topology.Instance{
			Type: "container",
			URL:  "agents",
		}),
		Hostname: hostname,
	}, MakeContainerTopologyCollector())
}

func TestBuildContainerTopology(t *testing.T) {
	collector := ContainerTopologyCollector{
		CheckTopologyCollector: corechecks.MakeCheckTopologyCollector("checkName", topology.Instance{
			Type: "checkName",
			URL:  "agents",
		}),
		Hostname: "host",
	}

	components, err := collector.collectContainers(MockUtil{})
	assert.NoError(t, err)
	assert.Equal(t, []*topology.Component{
		{
			ExternalID: "urn:container:checkName:/host:containerId1",
			Type:       topology.Type{Name: "container"},
			Data: topology.Data{
				"name":        "container1",
				"state":       "running",
				"type":        "containerd",
				"containerId": "containerId1",
				"image":       "image1",
				"mounts": []specs.Mount{
					{
						Source:      "source1",
						Destination: "dest1",
						Type:        "",
						Options:     nil,
					},
					{
						Destination: "dest2",
						Type:        "",
						Source:      "source2",
						Options:     nil,
					},
				},
				"identifiers": []string{"urn:container:/host:containerId1"},
				"labels":      []string{"runtime:containerd"},
			},
		},
		{
			ExternalID: "urn:container:checkName:/host:containerId2",
			Type: topology.Type{
				Name: "container",
			},
			Data: topology.Data{
				"containerId": "containerId2",
				"image":       "image2",
				"mounts": []specs.Mount{
					{
						Destination: "dest1",
						Type:        "",
						Source:      "source1",
						Options:     nil,
					},
					{
						Destination: "dest2",
						Type:        "",
						Source:      "source2",
						Options:     nil,
					},
				},
				"name":        "container2",
				"state":       "running",
				"type":        "docker",
				"identifiers": []string{"urn:container:/host:containerId2"},
				"labels":      []string{"runtime:docker"},
			},
		},
	}, components)
}
