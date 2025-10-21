package batcher

import (
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/health"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/telemetry"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
)

// MockBatcher mocks implementation of a batcher
type MockBatcher struct {
	CollectedTopology BatchBuilder
}

func createMockBatcher() *MockBatcher {
	batchBuilder := NewBatchBuilder(1000)
	batchBuilder.DisabledForceFlush()
	return &MockBatcher{
		CollectedTopology: batchBuilder,
	}
}

// SubmitComponent mock
func (batcher *MockBatcher) SubmitComponent(checkID checkid.ID, instance topology.Instance, component topology.Component) {
	batcher.CollectedTopology.AddComponent(checkID, instance, component)
}

// SubmitRelation mock
func (batcher *MockBatcher) SubmitRelation(checkID checkid.ID, instance topology.Instance, relation topology.Relation) {
	batcher.CollectedTopology.AddRelation(checkID, instance, relation)
}

// SubmitStartSnapshot mock
func (batcher *MockBatcher) SubmitStartSnapshot(checkID checkid.ID, instance topology.Instance) {
	batcher.CollectedTopology.TopologyStartSnapshot(checkID, instance)
}

// SubmitStopSnapshot mock
func (batcher *MockBatcher) SubmitStopSnapshot(checkID checkid.ID, instance topology.Instance) {
	batcher.CollectedTopology.TopologyStopSnapshot(checkID, instance)
}

// SubmitDelete mock
func (batcher *MockBatcher) SubmitDelete(checkID checkid.ID, instance topology.Instance, topologyElementID string) {
	batcher.CollectedTopology.Delete(checkID, instance, topologyElementID)
}

// SubmitHealthCheckData mock
func (batcher *MockBatcher) SubmitHealthCheckData(checkID checkid.ID, stream health.Stream, data health.CheckData) {
	batcher.CollectedTopology.AddHealthCheckData(checkID, stream, data)
}

// SubmitHealthStartSnapshot mock
func (batcher *MockBatcher) SubmitHealthStartSnapshot(checkID checkid.ID, stream health.Stream, intervalSeconds int, repeatSeconds int) {
	batcher.CollectedTopology.HealthStartSnapshot(checkID, stream, intervalSeconds, repeatSeconds)
}

// SubmitHealthStopSnapshot mock
func (batcher *MockBatcher) SubmitHealthStopSnapshot(checkID checkid.ID, stream health.Stream) {
	batcher.CollectedTopology.HealthStopSnapshot(checkID, stream)
}

// SubmitRawMetricsData mock
func (batcher *MockBatcher) SubmitRawMetricsData(checkID checkid.ID, rawMetric telemetry.RawMetric) {
	batcher.CollectedTopology.AddRawMetricsData(checkID, rawMetric)
}

// SubmitComplete mock
func (batcher *MockBatcher) SubmitComplete(checkID checkid.ID) {

}

// Shutdown mock
func (batcher *MockBatcher) Shutdown() {}
