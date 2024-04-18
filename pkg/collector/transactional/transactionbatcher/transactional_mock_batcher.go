package transactionbatcher

import (
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/health"
	"github.com/DataDog/datadog-agent/pkg/metrics/event"
	"github.com/DataDog/datadog-agent/pkg/telemetry"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"sync"
)

// MockTransactionalBatcher mocks implementation of a transactionbatcher
type MockTransactionalBatcher struct {
	CollectedTopology TransactionBatchBuilder
	mux               sync.Mutex
}

func newMockTransactionalBatcher() *MockTransactionalBatcher {
	return &MockTransactionalBatcher{
		CollectedTopology: NewTransactionalBatchBuilder(1000),
	}
}

// SubmitComponent submits a component to the batch
func (mtb *MockTransactionalBatcher) SubmitComponent(checkID checkid.ID, transactionID string, instance topology.Instance, component topology.Component) {
	mtb.mux.Lock()
	mtb.CollectedTopology.AddComponent(checkID, transactionID, instance, component)
	mtb.mux.Unlock()
}

// SubmitRelation submits a relation to the batch
func (mtb *MockTransactionalBatcher) SubmitRelation(checkID checkid.ID, transactionID string, instance topology.Instance, relation topology.Relation) {
	mtb.mux.Lock()
	mtb.CollectedTopology.AddRelation(checkID, transactionID, instance, relation)
	mtb.mux.Unlock()
}

// SubmitStartSnapshot submits start of a snapshot
func (mtb *MockTransactionalBatcher) SubmitStartSnapshot(checkID checkid.ID, transactionID string, instance topology.Instance) {
	mtb.mux.Lock()
	mtb.CollectedTopology.TopologyStartSnapshot(checkID, transactionID, instance)
	mtb.mux.Unlock()
}

// SubmitStopSnapshot submits a stop of a snapshot. This always causes a flush of the data downstream
func (mtb *MockTransactionalBatcher) SubmitStopSnapshot(checkID checkid.ID, transactionID string, instance topology.Instance) {
	mtb.mux.Lock()
	mtb.CollectedTopology.TopologyStopSnapshot(checkID, transactionID, instance)
	mtb.mux.Unlock()
}

// SubmitDelete submits a deletion of topology element.
func (mtb *MockTransactionalBatcher) SubmitDelete(checkID checkid.ID, transactionID string, instance topology.Instance, topologyElementID string) {
	mtb.mux.Lock()
	mtb.CollectedTopology.Delete(checkID, transactionID, instance, topologyElementID)
	mtb.mux.Unlock()
}

// SubmitHealthCheckData submits a Health check data record to the batch
func (mtb *MockTransactionalBatcher) SubmitHealthCheckData(checkID checkid.ID, transactionID string, stream health.Stream, data health.CheckData) {
	mtb.mux.Lock()
	mtb.CollectedTopology.AddHealthCheckData(checkID, transactionID, stream, data)
	mtb.mux.Unlock()
}

// SubmitHealthStartSnapshot submits start of a Health snapshot
func (mtb *MockTransactionalBatcher) SubmitHealthStartSnapshot(checkID checkid.ID, transactionID string, stream health.Stream, intervalSeconds int, expirySeconds int) {
	mtb.mux.Lock()
	mtb.CollectedTopology.HealthStartSnapshot(checkID, transactionID, stream, intervalSeconds, expirySeconds)
	mtb.mux.Unlock()
}

// SubmitHealthStopSnapshot submits a stop of a Health snapshot. This always causes a flush of the data downstream
func (mtb *MockTransactionalBatcher) SubmitHealthStopSnapshot(checkID checkid.ID, transactionID string, stream health.Stream) {
	mtb.mux.Lock()
	mtb.CollectedTopology.HealthStopSnapshot(checkID, transactionID, stream)
	mtb.mux.Unlock()
}

// SubmitRawMetricsData submits a raw metrics data record to the batch
func (mtb *MockTransactionalBatcher) SubmitRawMetricsData(checkID checkid.ID, transactionID string, rawMetric telemetry.RawMetrics) {
	mtb.mux.Lock()
	mtb.CollectedTopology.AddRawMetricsData(checkID, transactionID, rawMetric)
	mtb.mux.Unlock()
}

// SubmitEvent submits an event to the batch
func (mtb *MockTransactionalBatcher) SubmitEvent(checkID checkid.ID, transactionID string, event event.Event) {
	mtb.mux.Lock()
	mtb.CollectedTopology.AddEvent(checkID, transactionID, event)
	mtb.mux.Unlock()
}

// StartTransaction starts a transaction for the given check ID
func (mtb *MockTransactionalBatcher) StartTransaction(checkID checkid.ID, transactionID string) {
	mtb.mux.Lock()
	mtb.CollectedTopology.StartTransaction(checkID, transactionID)
	mtb.mux.Unlock()
}

// SubmitCompleteTransaction marks a transaction as complete
func (mtb *MockTransactionalBatcher) SubmitCompleteTransaction(checkID checkid.ID, transactionID string) {
	mtb.mux.Lock()
	// for the mock let's set this as the state again so we can assert it in the tests
	flushedStates := mtb.CollectedTopology.MarkTransactionComplete(checkID, transactionID)
	mtb.CollectedTopology.states = flushedStates
	mtb.mux.Unlock()
}

// GetCheckState returns the TransactionCheckInstanceBatchState for a given check ID
func (mtb *MockTransactionalBatcher) GetCheckState(checkID checkid.ID) (TransactionCheckInstanceBatchState, bool) {
	mtb.mux.Lock()
	state, ok := mtb.CollectedTopology.states[checkID]
	mtb.mux.Unlock()
	return state, ok
}

// SubmitClearState clears the batch state for a given checkID
func (mtb *MockTransactionalBatcher) SubmitClearState(checkID checkid.ID) {
	mtb.mux.Lock()
	mtb.CollectedTopology.ClearState(checkID)
	mtb.mux.Unlock()
}

// Stop shuts down the transactionbatcher and resets the singleton init
func (mtb *MockTransactionalBatcher) Stop() {
	// reset the tmInit to re-init the batcher
	batcherInit = new(sync.Once)
}
