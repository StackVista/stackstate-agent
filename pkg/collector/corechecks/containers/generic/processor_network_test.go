// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package generic

import (
	"testing"

	"github.com/DataDog/datadog-agent/pkg/aggregator/mocksender"
	"github.com/DataDog/datadog-agent/pkg/tagger"
	"github.com/DataDog/datadog-agent/pkg/tagger/collectors"
	"github.com/DataDog/datadog-agent/pkg/tagger/local"
	"github.com/DataDog/datadog-agent/pkg/util/containers"
	"github.com/DataDog/datadog-agent/pkg/util/containers/metrics"
	"github.com/DataDog/datadog-agent/pkg/util/containers/metrics/mock"
	"github.com/DataDog/datadog-agent/pkg/util/pointer"
)

func TestNetworkProcessorExtension(t *testing.T) {
	mockSender := mocksender.NewMockSender("network-extension")
	mockSender.SetupAcceptAll()

	fakeTagger := local.NewFakeTagger()
	tagger.SetDefaultTagger(fakeTagger)

	mockCollector := mock.NewCollector("testCollector")

	networkProcessor := NewProcessorNetwork()

	// Test setup:
	// container1 & container2 share the same network namespace (should report metrics once with orchestrator tags)
	// container3 has no namespace information (should report with high tags)
	// container4 is standalone is a namespace (should report with high tags)
	// container5 is using host network (should not report at all)
	container1 := CreateContainerMeta("docker", "1")
	fakeTagger.SetTags(containers.BuildTaggerEntityName(container1.ID), "foo", []string{"low:common"}, []string{"orch:common12"}, []string{"id:container1"}, nil)
	mockCollector.SetContainerEntry(container1.ID, mock.ContainerEntry{
		NetworkStats: &metrics.ContainerNetworkStats{
			BytesSent:   pointer.Float64Ptr(12),
			BytesRcvd:   pointer.Float64Ptr(12),
			PacketsSent: pointer.Float64Ptr(12),
			PacketsRcvd: pointer.Float64Ptr(12),
			Interfaces: map[string]metrics.InterfaceNetStats{
				"eth0": {
					BytesSent:   pointer.Float64Ptr(12),
					BytesRcvd:   pointer.Float64Ptr(12),
					PacketsSent: pointer.Float64Ptr(12),
					PacketsRcvd: pointer.Float64Ptr(12),
				},
			},
			NetworkIsolationGroupID: pointer.UInt64Ptr(100),
			UsingHostNetwork:        pointer.BoolPtr(false),
		},
	})

	container2 := CreateContainerMeta("docker", "2")
	fakeTagger.SetTags(containers.BuildTaggerEntityName(container2.ID), "foo", []string{"low:common"}, []string{"orch:common12"}, []string{"id:container2"}, nil)
	mockCollector.SetContainerEntry(container2.ID, mock.ContainerEntry{
		NetworkStats: &metrics.ContainerNetworkStats{
			BytesSent:   pointer.Float64Ptr(12),
			BytesRcvd:   pointer.Float64Ptr(12),
			PacketsSent: pointer.Float64Ptr(12),
			PacketsRcvd: pointer.Float64Ptr(12),
			Interfaces: map[string]metrics.InterfaceNetStats{
				"eth0": {
					BytesSent:   pointer.Float64Ptr(12),
					BytesRcvd:   pointer.Float64Ptr(12),
					PacketsSent: pointer.Float64Ptr(12),
					PacketsRcvd: pointer.Float64Ptr(12),
				},
			},
			NetworkIsolationGroupID: pointer.UInt64Ptr(100),
			UsingHostNetwork:        pointer.BoolPtr(false),
		},
	})

	container3 := CreateContainerMeta("docker", "3")
	fakeTagger.SetTags(containers.BuildTaggerEntityName(container3.ID), "foo", []string{"low:common"}, []string{"orch:standalone3"}, []string{"id:container3"}, nil)
	mockCollector.SetContainerEntry(container3.ID, mock.ContainerEntry{
		NetworkStats: &metrics.ContainerNetworkStats{
			BytesSent:   pointer.Float64Ptr(3),
			BytesRcvd:   pointer.Float64Ptr(3),
			PacketsSent: pointer.Float64Ptr(3),
			PacketsRcvd: pointer.Float64Ptr(3),
			Interfaces: map[string]metrics.InterfaceNetStats{
				"eth0": {
					BytesSent:   pointer.Float64Ptr(3),
					BytesRcvd:   pointer.Float64Ptr(3),
					PacketsSent: pointer.Float64Ptr(3),
					PacketsRcvd: pointer.Float64Ptr(3),
				},
			},
		},
	})

	container4 := CreateContainerMeta("docker", "4")
	fakeTagger.SetTags(containers.BuildTaggerEntityName(container4.ID), "foo", []string{"low:common"}, []string{"orch:standalone4"}, []string{"id:container4"}, nil)
	mockCollector.SetContainerEntry(container4.ID, mock.ContainerEntry{
		NetworkStats: &metrics.ContainerNetworkStats{
			BytesSent:   pointer.Float64Ptr(4),
			BytesRcvd:   pointer.Float64Ptr(4),
			PacketsSent: pointer.Float64Ptr(4),
			PacketsRcvd: pointer.Float64Ptr(4),
			Interfaces: map[string]metrics.InterfaceNetStats{
				"eth0": {
					BytesSent:   pointer.Float64Ptr(4),
					BytesRcvd:   pointer.Float64Ptr(4),
					PacketsSent: pointer.Float64Ptr(4),
					PacketsRcvd: pointer.Float64Ptr(4),
				},
			},
			NetworkIsolationGroupID: pointer.UInt64Ptr(400),
			UsingHostNetwork:        pointer.BoolPtr(false),
		},
	})

	container5 := CreateContainerMeta("docker", "5")
	fakeTagger.SetTags(containers.BuildTaggerEntityName(container5.ID), "foo", []string{"low:common"}, []string{"orch:standalone5"}, []string{"id:container5"}, nil)
	mockCollector.SetContainerEntry(container5.ID, mock.ContainerEntry{
		NetworkStats: &metrics.ContainerNetworkStats{
			BytesSent:   pointer.Float64Ptr(5),
			BytesRcvd:   pointer.Float64Ptr(5),
			PacketsSent: pointer.Float64Ptr(5),
			PacketsRcvd: pointer.Float64Ptr(5),
			Interfaces: map[string]metrics.InterfaceNetStats{
				"eth0": {
					BytesSent:   pointer.Float64Ptr(5),
					BytesRcvd:   pointer.Float64Ptr(5),
					PacketsSent: pointer.Float64Ptr(5),
					PacketsRcvd: pointer.Float64Ptr(5),
				},
			},
			NetworkIsolationGroupID: pointer.UInt64Ptr(1),
			UsingHostNetwork:        pointer.BoolPtr(true),
		},
	})

	// Running them through the ProcessorExtension
	networkProcessor.PreProcess(MockSendMetric, mockSender)

	container1Tags, _ := fakeTagger.Tag("container_id://1", collectors.HighCardinality)
	networkProcessor.Process(container1Tags, container1, mockCollector, 0)
	container2Tags, _ := fakeTagger.Tag("container_id://2", collectors.HighCardinality)
	networkProcessor.Process(container2Tags, container2, mockCollector, 0)
	container3Tags, _ := fakeTagger.Tag("container_id://3", collectors.HighCardinality)
	networkProcessor.Process(container3Tags, container3, mockCollector, 0)
	container4Tags, _ := fakeTagger.Tag("container_id://4", collectors.HighCardinality)
	networkProcessor.Process(container4Tags, container4, mockCollector, 0)
	container5Tags, _ := fakeTagger.Tag("container_id://5", collectors.HighCardinality)
	networkProcessor.Process(container5Tags, container5, mockCollector, 0)

	networkProcessor.PostProcess()

	// Checking results
	mockSender.AssertNumberOfCalls(t, "Rate", 12)

	// Container 1 & 2
	mockSender.AssertMetric(t, "Rate", "container.net.sent", 12, "", []string{"low:common", "orch:common12", "interface:eth0"})
	mockSender.AssertMetric(t, "Rate", "container.net.sent.packets", 12, "", []string{"low:common", "orch:common12", "interface:eth0"})
	mockSender.AssertMetric(t, "Rate", "container.net.rcvd", 12, "", []string{"low:common", "orch:common12", "interface:eth0"})
	mockSender.AssertMetric(t, "Rate", "container.net.rcvd.packets", 12, "", []string{"low:common", "orch:common12", "interface:eth0"})

	// Container 3
	mockSender.AssertMetric(t, "Rate", "container.net.sent", 3, "", []string{"low:common", "orch:standalone3", "id:container3", "interface:eth0"})
	mockSender.AssertMetric(t, "Rate", "container.net.sent.packets", 3, "", []string{"low:common", "orch:standalone3", "id:container3", "interface:eth0"})
	mockSender.AssertMetric(t, "Rate", "container.net.rcvd", 3, "", []string{"low:common", "orch:standalone3", "id:container3", "interface:eth0"})
	mockSender.AssertMetric(t, "Rate", "container.net.rcvd.packets", 3, "", []string{"low:common", "orch:standalone3", "id:container3", "interface:eth0"})

	// Container 4
	mockSender.AssertMetric(t, "Rate", "container.net.sent", 4, "", []string{"low:common", "orch:standalone4", "id:container4", "interface:eth0"})
	mockSender.AssertMetric(t, "Rate", "container.net.sent.packets", 4, "", []string{"low:common", "orch:standalone4", "id:container4", "interface:eth0"})
	mockSender.AssertMetric(t, "Rate", "container.net.rcvd", 4, "", []string{"low:common", "orch:standalone4", "id:container4", "interface:eth0"})
	mockSender.AssertMetric(t, "Rate", "container.net.rcvd.packets", 4, "", []string{"low:common", "orch:standalone4", "id:container4", "interface:eth0"})
}