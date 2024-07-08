// StackState
package topology

import (
	"context"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	cspec "github.com/DataDog/datadog-agent/pkg/collector/corechecks/containers/spec"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/config/model"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/hostname"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type MockUtil struct {
}

func (m MockUtil) GetContainers(ctx context.Context) ([]*cspec.Container, error) {
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

func TestMakeNodeAgentContainerTopologyCollector(t *testing.T) {
	// Let configIsFeaturePresent(config.Kubernetes) return true
	os.Setenv("DOCKER_DD_AGENT", "true")
	config.SetFeatures(t, config.Docker)
	config.Datadog.Set("hostname", "host", model.SourceDefault)
	hostname, err := hostname.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, &ContainerTopologyCollector{
		CheckTopologyCollector: corechecks.MakeCheckTopologyCollector("container_topology", topology.Instance{
			Type: "container",
			URL:  "agents",
		}),
		Hostname: hostname,
		Runtime:  "test",
	}, MakeContainerTopologyCollector("test"))
	os.Unsetenv("DOCKER_DD_AGENT")
	config.Datadog.UnsetForSource("hostname", model.SourceDefault)
}

func TestMakeClusterAgentContainerTopologyCollector(t *testing.T) {
	config.Datadog.Set("hostname", "host", model.SourceDefault)
	hostname, err := hostname.Get(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, &ContainerTopologyCollector{
		CheckTopologyCollector: corechecks.MakeCheckTopologyCollector("container_topology", topology.Instance{
			Type: "container",
			URL:  "agents",
		}),
		Hostname: hostname,
		Runtime:  "test",
	}, MakeContainerTopologyCollector("test"))
	config.Datadog.UnsetForSource("hostname", model.SourceDefault)
}

func TestBuildContainerTopology(t *testing.T) {
	collector := ContainerTopologyCollector{
		CheckTopologyCollector: corechecks.MakeCheckTopologyCollector("checkName", topology.Instance{
			Type: "checkName",
			URL:  "agents",
		}),
		Hostname: "host",
		Runtime:  "test",
	}

	components, err := collector.collectContainers(MockUtil{})
	assert.NoError(t, err)
	assert.Equal(t, []*topology.Component{
		{
			ExternalID: "urn:container:containerd:/host:containerId1",
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
			ExternalID: "urn:container:docker:/host:containerId2",
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
