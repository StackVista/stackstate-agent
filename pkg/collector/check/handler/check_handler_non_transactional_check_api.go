package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	checkState "github.com/StackVista/stackstate-agent/pkg/collector/check/state"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// StartTransaction "upgrades" the non-transactional check handler to a transactional check handler, registers it in the
// check manager and calls StartTransaction on the newly created transactional check handler.
func (ch *NonTransactionalCheckHandler) StartTransaction() string {
	transactionalCheckHandler := GetCheckManager().MakeCheckHandlerTransactional(ch.ID())
	if transactionalCheckHandler != nil {
		return transactionalCheckHandler.StartTransaction()
	}
	return ""
}

// DiscardTransaction logs a warning for the non-transactional check handler. This should never be called.
func (ch *NonTransactionalCheckHandler) DiscardTransaction(string) {
	_ = log.Warnf("DiscardTransaction called on NonTransactionalCheckHandler. This should never happen.")
}

// StopTransaction logs a warning for the non-transactional check handler. This should never be called.
func (ch *NonTransactionalCheckHandler) StopTransaction() {
	_ = log.Warnf("StopTransaction called on NonTransactionalCheckHandler. This should never happen.")
}

// SetTransactionState logs a warning for the non-transactional check handler. This should never be called.
func (ch *NonTransactionalCheckHandler) SetTransactionState(string, string) {
	_ = log.Warnf("SetTransactionState called on NonTransactionalCheckHandler. This should never happen.")
}

// SetState is used to commit state for a given state key and CheckState
func (ch *NonTransactionalCheckHandler) SetState(key string, state string) {
	err := checkState.GetCheckStateManager().SetState(key, state)
	if err != nil {
		_ = log.Errorf("error occurred when setting state for check %s with value %s->%s, %s", ch.ID(), key, state, err)
	}
}

// GetState returns a CheckState for a given key
func (ch *NonTransactionalCheckHandler) GetState(key string) string {
	s, err := checkState.GetCheckStateManager().GetState(key)
	if err != nil {
		_ = log.Errorf("error occurred when reading state for check %s for key %s: %s", ch.ID(), key, err)
	}
	log.Infof("Retrieved state for check %s, state key: %s, value: %s", ch.ID(), key, s)
	return s
}

// SubmitComponent submits a component to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitComponent(instance topology.Instance, component topology.Component) {
	batcher.GetBatcher().SubmitComponent(ch.ID(), instance, component)
}

// SubmitRelation submits a relation to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitRelation(instance topology.Instance, relation topology.Relation) {
	batcher.GetBatcher().SubmitRelation(ch.ID(), instance, relation)
}

// SubmitStartSnapshot submits a start snapshot to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitStartSnapshot(instance topology.Instance) {
	batcher.GetBatcher().SubmitStartSnapshot(ch.ID(), instance)
}

// SubmitStopSnapshot submits a stop snapshot to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitStopSnapshot(instance topology.Instance) {
	batcher.GetBatcher().SubmitStopSnapshot(ch.ID(), instance)
}

// SubmitDelete submits a topology element delete to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitDelete(instance topology.Instance, topologyElementID string) {
	batcher.GetBatcher().SubmitDelete(ch.ID(), instance, topologyElementID)
}

// SubmitHealthCheckData submits health check data to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitHealthCheckData(stream health.Stream, data health.CheckData) {
	batcher.GetBatcher().SubmitHealthCheckData(ch.ID(), stream, data)
}

// SubmitHealthStartSnapshot submits a health start snapshot to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int) {
	batcher.GetBatcher().SubmitHealthStartSnapshot(ch.ID(), stream, intervalSeconds, expirySeconds)
}

// SubmitHealthStopSnapshot submits a health stop snapshot to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitHealthStopSnapshot(stream health.Stream) {
	batcher.GetBatcher().SubmitHealthStopSnapshot(ch.ID(), stream)
}

// SubmitRawMetricsData submits a raw metric value to the Global Batcher to be batched.
func (ch *NonTransactionalCheckHandler) SubmitRawMetricsData(data telemetry.RawMetrics) {
	batcher.GetBatcher().SubmitRawMetricsData(ch.ID(), data)
}

// SubmitEvent submits an event to the forwarder.
func (ch *NonTransactionalCheckHandler) SubmitEvent(event metrics.Event) {
	sender, err := aggregator.GetSender(ch.ID())
	if err != nil || sender == nil {
		_ = log.Errorf("Error submitting metric to the Sender: %v", err)
		return
	}

	sender.Event(event)
}

// SubmitComplete submits a complete to the Global Batcher.
func (ch *NonTransactionalCheckHandler) SubmitComplete() {
	batcher.GetBatcher().SubmitComplete(ch.ID())
}
