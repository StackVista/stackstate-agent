package handler

import (
	"fmt"
	checkState "github.com/StackVista/stackstate-agent/pkg/collector/check/state"
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

// CancelTransaction triggers a transaction failure and reloads the check
func (ch *checkHandler) CancelTransaction(reason string) {
	ch.currentTransactionChannel <- CancelTransaction{
		Reason: reason,
	}
}

// SetTransactionState is used to set state transactionaly. This state is only committed once a transaction has been
// completed successfully.
func (ch *checkHandler) SetTransactionState(key string, state string) {
	ch.currentTransactionChannel <- SubmitSetTransactionState{
		Key:   key,
		State: state,
	}
}

// SetState is used to commit state for a given state key and CheckState
func (ch *checkHandler) SetState(key string, state string) {
	err := checkState.GetCheckStateManager().SetState(key, state)
	if err != nil {
		reason := fmt.Sprintf("error occurred when setting state for %s->%s, %s", key, state, err)
		// trigger cancel transaction, check reload
		ch.CancelTransaction(reason)
	}
}

// GetState returns a CheckState for a given key
func (ch *checkHandler) GetState(key string) string {
	s, err := checkState.GetCheckStateManager().GetState(key)
	if err != nil {
		_ = log.Errorf("error occurred when reading state for check %s for key %s: %s", ch.ID(), key, err)
	}
	return s
}

// SubmitComponent submits a component to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitComponent(instance topology.Instance, component topology.Component) {
	ch.currentTransactionChannel <- SubmitComponent{
		Instance:  instance,
		Component: component,
	}
}

// SubmitRelation submits a relation to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitRelation(instance topology.Instance, relation topology.Relation) {
	ch.currentTransactionChannel <- SubmitRelation{
		Instance: instance,
		Relation: relation,
	}

}

// SubmitStartSnapshot submits a start snapshot to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitStartSnapshot(instance topology.Instance) {
	ch.currentTransactionChannel <- SubmitStartSnapshot{Instance: instance}

}

// SubmitStopSnapshot submits a stop snapshot to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitStopSnapshot(instance topology.Instance) {
	ch.currentTransactionChannel <- SubmitStopSnapshot{Instance: instance}
}

// SubmitDelete submits a topology element delete to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitDelete(instance topology.Instance, topologyElementID string) {
	ch.currentTransactionChannel <- SubmitDelete{
		Instance:          instance,
		TopologyElementID: topologyElementID,
	}
}

// SubmitHealthCheckData submits health check data to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitHealthCheckData(stream health.Stream, data health.CheckData) {
	ch.currentTransactionChannel <- SubmitHealthCheckData{
		Stream: stream,
		Data:   data,
	}
}

// SubmitHealthStartSnapshot submits a health start snapshot to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int) {
	ch.currentTransactionChannel <- SubmitHealthStartSnapshot{
		Stream:          stream,
		IntervalSeconds: intervalSeconds,
		ExpirySeconds:   expirySeconds,
	}
}

// SubmitHealthStopSnapshot submits a health stop snapshot to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitHealthStopSnapshot(stream health.Stream) {
	ch.currentTransactionChannel <- SubmitHealthStopSnapshot{
		Stream: stream,
	}
}

// SubmitRawMetricsData submits a raw metric value to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitRawMetricsData(data telemetry.RawMetrics) {
	ch.currentTransactionChannel <- SubmitRawMetric{
		Value: data,
	}
}

// SubmitComplete submits a complete to the current transaction channel to be forwarded.
func (ch *checkHandler) SubmitComplete() {
	ch.currentTransactionChannel <- SubmitComplete{}
}
