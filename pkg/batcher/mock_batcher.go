package batcher

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

// MockBatcher mocks implementation of a batcher
type MockBatcher struct {
	CollectedTopology BatchBuilder
	Errors            []error
}

func createMockBatcher() *MockBatcher {
	batchBuilder := NewBatchBuilder(1000)
	batchBuilder.DisabledForceFlush()
	return &MockBatcher{
		CollectedTopology: batchBuilder,
	}
}

// SubmitComponent mock
func (batcher *MockBatcher) SubmitComponent(checkID check.ID, instance topology.Instance, component topology.Component) {
	batcher.CollectedTopology.AddComponent(checkID, instance, component)
}

// SubmitRelation mock
func (batcher *MockBatcher) SubmitRelation(checkID check.ID, instance topology.Instance, relation topology.Relation) {
	batcher.CollectedTopology.AddRelation(checkID, instance, relation)
}

// SubmitStartSnapshot mock
func (batcher *MockBatcher) SubmitStartSnapshot(checkID check.ID, instance topology.Instance) {
	batcher.CollectedTopology.TopologyStartSnapshot(checkID, instance)
}

// SubmitStopSnapshot mock
func (batcher *MockBatcher) SubmitStopSnapshot(checkID check.ID, instance topology.Instance) {
	batcher.CollectedTopology.TopologyStopSnapshot(checkID, instance)
}

// SubmitDelete mock
func (batcher *MockBatcher) SubmitDelete(checkID check.ID, instance topology.Instance, topologyElementID string) {
	batcher.CollectedTopology.Delete(checkID, instance, topologyElementID)
}

// SubmitHealthCheckData mock
func (batcher *MockBatcher) SubmitHealthCheckData(checkID check.ID, stream health.Stream, data health.CheckData) {
	batcher.CollectedTopology.AddHealthCheckData(checkID, stream, data)
}

// SubmitHealthStartSnapshot mock
func (batcher *MockBatcher) SubmitHealthStartSnapshot(checkID check.ID, stream health.Stream, intervalSeconds int, repeatSeconds int) {
	batcher.CollectedTopology.HealthStartSnapshot(checkID, stream, intervalSeconds, repeatSeconds)
}

// SubmitHealthStopSnapshot mock
func (batcher *MockBatcher) SubmitHealthStopSnapshot(checkID check.ID, stream health.Stream) {
	batcher.CollectedTopology.HealthStopSnapshot(checkID, stream)
}

// SubmitRawMetricsData mock
func (batcher *MockBatcher) SubmitRawMetricsData(checkID check.ID, rawMetric telemetry.RawMetrics) {
	batcher.CollectedTopology.AddRawMetricsData(checkID, rawMetric)
}

// SubmitComplete mock
func (batcher *MockBatcher) SubmitComplete(checkID check.ID) {

}

// Shutdown mock
func (batcher *MockBatcher) Shutdown() {}

// SubmitError keeps track of thrown errors
func (batcher *MockBatcher) SubmitError(checkID check.ID, err error) {
	batcher.Errors = append(batcher.Errors, err)
}
