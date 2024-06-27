//go:build !windows

package disk

import (
	"context"
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/hostname"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/shirou/gopsutil/v3/disk"
)

const diskCheckID = "disk_topology"

// TopologyCollector contains all the metadata needed to produce disk topology
type TopologyCollector struct {
	corechecks.CheckTopologyCollector
}

// MakeTopologyCollector returns an instance of the DiskTopologyCollector
func MakeTopologyCollector() *TopologyCollector {
	return &TopologyCollector{
		corechecks.MakeCheckTopologyCollector(diskCheckID, topology.Instance{
			Type: "disk",
			URL:  "agents",
		}),
	}
}

// BuildTopology creates / collects and produces disk topology
func (dtc *TopologyCollector) BuildTopology(partitions []disk.PartitionStat) error {
	sender := batcher.GetBatcher()

	// try to get the agent hostname to use in the host component
	hostnameData, err := hostname.GetWithProvider(context.TODO())
	if err != nil {
		log.Warnf("Can't get hostname for host running the disk integration, not reporting a host: %s", err)
		return err
	}

	// produce a host component with all the disk devices as metadata
	diskComponent := dtc.createDiskComponent(hostnameData.Hostname, hostnameData.Identifiers, partitions)
	sender.SubmitComponent(dtc.CheckID, dtc.TopologyInstance, diskComponent)

	sender.SubmitComplete(dtc.CheckID)

	return nil
}

// createDiskComponent creates a topology.Component given a hostname and disk partitions
func (dtc *TopologyCollector) createDiskComponent(hostname string, identifiers []string, partitions []disk.PartitionStat) topology.Component {
	deviceMap := make(map[string]bool, 0)
	hostDevices := make([]string, 0)
	for _, part := range partitions {
		// filter out duplicate partitions
		if _, value := deviceMap[part.Device]; !value {
			deviceMap[part.Device] = true
			hostDevices = append(hostDevices, part.Device)
		}
	}

	return topology.Component{
		ExternalID: fmt.Sprintf("urn:host:/%s", hostname),
		Type:       topology.Type{Name: "host"},
		Data: topology.Data{
			"host":        hostname,
			"identifiers": identifiers,
			"devices":     hostDevices,
		},
	}
}
