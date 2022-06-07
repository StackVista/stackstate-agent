package handler

import (
	checkState "github.com/StackVista/stackstate-agent/pkg/collector/check/state"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/google/uuid"
)

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
	ch.currentTransactionChannel <- StopTransaction{}
}

// SetStateTransactional is used to set state transactionaly. This state is only committed once a transaction has been
// completed successfully.
func (ch *checkHandler) SetStateTransactional(key string, state string) {
	ch.currentTransactionChannel <- SubmitSetStateTransactional{
		Key:   key,
		State: state,
	}
}

// SetState is used to commit state for a given state key and CheckState
func (ch *checkHandler) SetState(key string, state string) error {
	return checkState.GetCheckStateManager().SetState(key, state)
}

// GetState returns a CheckState for a given key
func (ch *checkHandler) GetState(key string) string {
	s, err := checkState.GetCheckStateManager().GetState(key)
	if err != nil {
		_ = log.Errorf("error occurred when reading state for check %s for key %s: %s", ch.ID(), key, err)
	}
	return s
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
