package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/google/uuid"
)

// CheckAPI contains all the operations that can be done by an Agent Check. This acts as a proxy to forward data
// where it needs to go.
type CheckAPI interface {
	// Transactionality
	StartTransaction() string
	StopTransaction()

	// Topology
	SubmitComponent(instance topology.Instance, component topology.Component)
	SubmitRelation(instance topology.Instance, relation topology.Relation)
	SubmitStartSnapshot(instance topology.Instance)
	SubmitStopSnapshot(instance topology.Instance)
	SubmitDelete(instance topology.Instance, topologyElementID string)

	// Health
	SubmitHealthCheckData(stream health.Stream, data health.CheckData)
	SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int)
	SubmitHealthStopSnapshot(stream health.Stream)

	// Raw Metrics
	SubmitRawMetricsData(data telemetry.RawMetrics)

	// lifecycle
	SubmitComplete()
}

// StartTransaction submits a start transaction for the check handler. This blocks any future transactions until
// this one completes, fails or is timed out.
func (ch *checkHandler) StartTransaction() string {
	transactionID := uuid.New().String()
	ch.transactionChannel <- StartTransaction{
		CheckID:       ch.ID(),
		TransactionID: transactionID,
	}
	return transactionID
}

// StopTransaction submits a complete to the Transactional Batcher, to send the final payload of the transaction
// and mark the current transaction as complete.
func (ch *checkHandler) StopTransaction() {
	transactionbatcher.GetTransactionalBatcher().SubmitCompleteTransaction(ch.ID(), ch.GetCurrentTransaction())
}

// SubmitComponent submits a component to the Transactional Batcher to be batched.
func (ch *checkHandler) SubmitComponent(instance topology.Instance, component topology.Component) {
	transactionbatcher.GetTransactionalBatcher().SubmitComponent(ch.ID(), ch.GetCurrentTransaction(), instance, component)
}

// SubmitRelation submits a relation to the Transactional Batcher to be batched.
func (ch *checkHandler) SubmitRelation(instance topology.Instance, relation topology.Relation) {
	transactionbatcher.GetTransactionalBatcher().SubmitRelation(ch.ID(), ch.GetCurrentTransaction(), instance, relation)
}

// SubmitStartSnapshot submits a start snapshot to the Transactional Batcher to be batched.
func (ch *checkHandler) SubmitStartSnapshot(instance topology.Instance) {
	transactionbatcher.GetTransactionalBatcher().SubmitStartSnapshot(ch.ID(), ch.GetCurrentTransaction(), instance)
}

// SubmitStopSnapshot submits a stop snapshot to the Transactional Batcher to be batched.
func (ch *checkHandler) SubmitStopSnapshot(instance topology.Instance) {
	transactionbatcher.GetTransactionalBatcher().SubmitStopSnapshot(ch.ID(), ch.GetCurrentTransaction(), instance)
}

// SubmitDelete submits a topology element delete to the Transactional Batcher to be batched.
func (ch *checkHandler) SubmitDelete(instance topology.Instance, topologyElementID string) {
	transactionbatcher.GetTransactionalBatcher().SubmitDelete(ch.ID(), ch.GetCurrentTransaction(), instance, topologyElementID)
}

// SubmitHealthCheckData submits health check data to the Transactional Batcher to be batched.
func (ch *checkHandler) SubmitHealthCheckData(stream health.Stream, data health.CheckData) {
	transactionbatcher.GetTransactionalBatcher().SubmitHealthCheckData(ch.ID(), ch.GetCurrentTransaction(), stream, data)
}

// SubmitHealthStartSnapshot submits a health start snapshot to the Transactional Batcher to be batched.
func (ch *checkHandler) SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int) {
	transactionbatcher.GetTransactionalBatcher().SubmitHealthStartSnapshot(ch.ID(), ch.GetCurrentTransaction(), stream,
		intervalSeconds, expirySeconds)
}

// SubmitHealthStopSnapshot submits a health stop snapshot to the Transactional Batcher to be batched.
func (ch *checkHandler) SubmitHealthStopSnapshot(stream health.Stream) {
	transactionbatcher.GetTransactionalBatcher().SubmitHealthStopSnapshot(ch.ID(), ch.GetCurrentTransaction(), stream)
}

// SubmitRawMetricsData submits a raw metric value to the Transactional Batcher to be batched.
func (ch *checkHandler) SubmitRawMetricsData(data telemetry.RawMetrics) {
	transactionbatcher.GetTransactionalBatcher().SubmitRawMetricsData(ch.ID(), ch.GetCurrentTransaction(), data)
}

// SubmitComplete submits a complete to the Transactional Batcher.
func (ch *checkHandler) SubmitComplete() {
	transactionbatcher.GetTransactionalBatcher().SubmitComplete(ch.ID())
}
