package transactionbatcher

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

// MockTransactionalBatcher mocks implementation of a transactionbatcher
type MockTransactionalBatcher struct {
	CollectedTopology TransactionBatchBuilder
}

func createMockTransactionalBatcher() MockTransactionalBatcher {
	return MockTransactionalBatcher{
		CollectedTopology: NewTransactionalBatchBuilder(1000),
	}
}

// SubmitComponent submits a component to the batch
func (mtb MockTransactionalBatcher) SubmitComponent(checkID check.ID, transactionID string, instance topology.Instance, component topology.Component) {
	mtb.CollectedTopology.AddComponent(checkID, transactionID, instance, component)
}

// SubmitRelation submits a relation to the batch
func (mtb MockTransactionalBatcher) SubmitRelation(checkID check.ID, transactionID string, instance topology.Instance, relation topology.Relation) {
	mtb.CollectedTopology.AddRelation(checkID, transactionID, instance, relation)
}

// SubmitStartSnapshot submits start of a snapshot
func (mtb MockTransactionalBatcher) SubmitStartSnapshot(checkID check.ID, transactionID string, instance topology.Instance) {
	mtb.CollectedTopology.TopologyStartSnapshot(checkID, transactionID, instance)
}

// SubmitStopSnapshot submits a stop of a snapshot. This always causes a flush of the data downstream
func (mtb MockTransactionalBatcher) SubmitStopSnapshot(checkID check.ID, transactionID string, instance topology.Instance) {
	mtb.CollectedTopology.TopologyStopSnapshot(checkID, transactionID, instance)
}

// SubmitDelete submits a deletion of topology element.
func (mtb MockTransactionalBatcher) SubmitDelete(checkID check.ID, transactionID string, instance topology.Instance, topologyElementID string) {
	mtb.CollectedTopology.Delete(checkID, transactionID, instance, topologyElementID)
}

// SubmitHealthCheckData submits a Health check data record to the batch
func (mtb MockTransactionalBatcher) SubmitHealthCheckData(checkID check.ID, transactionID string, stream health.Stream, data health.CheckData) {
	mtb.CollectedTopology.AddHealthCheckData(checkID, transactionID, stream, data)
}

// SubmitHealthStartSnapshot submits start of a Health snapshot
func (mtb MockTransactionalBatcher) SubmitHealthStartSnapshot(checkID check.ID, transactionID string, stream health.Stream, intervalSeconds int, expirySeconds int) {
	mtb.CollectedTopology.HealthStartSnapshot(checkID, transactionID, stream, intervalSeconds, expirySeconds)
}

// SubmitHealthStopSnapshot submits a stop of a Health snapshot. This always causes a flush of the data downstream
func (mtb MockTransactionalBatcher) SubmitHealthStopSnapshot(checkID check.ID, transactionID string, stream health.Stream) {
	mtb.CollectedTopology.HealthStopSnapshot(checkID, transactionID, stream)
}

// SubmitRawMetricsData submits a raw metrics data record to the batch
func (mtb MockTransactionalBatcher) SubmitRawMetricsData(checkID check.ID, transactionID string, rawMetric telemetry.RawMetrics) {
	mtb.CollectedTopology.AddRawMetricsData(checkID, transactionID, rawMetric)
}

// SubmitComplete signals completion of a check. May trigger a flush only if the check produced data
func (mtb MockTransactionalBatcher) SubmitComplete(checkID check.ID) {
}

// Shutdown shuts down the transactionbatcher
func (mtb MockTransactionalBatcher) Shutdown() {
}
