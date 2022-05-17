package transactionbatcher

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

type BatchTransaction struct {
	TransactionID        string
	CompletedTransaction bool
}

// TransactionCheckInstanceBatchState is the type representing batched data per check Instance
type TransactionCheckInstanceBatchState struct {
	Transaction *BatchTransaction
	Topology    *topology.Topology
	Metrics     *telemetry.Metrics
	Health      map[string]health.Health
}

// TransactionCheckInstanceBatchStates is the type representing batched data for all check instances
type TransactionCheckInstanceBatchStates map[check.ID]TransactionCheckInstanceBatchState

// TransactionBatchBuilder is a helper class to build Topology based on submitted data, this data structure is not thread safe
type TransactionBatchBuilder struct {
	states TransactionCheckInstanceBatchStates
	// Count the amount of elements we gathered
	elementCount int
	// Amount of elements when we flush
	maxCapacity int
}

// NewTransactionalBatchBuilder constructs a TransactionBatchBuilder
func NewTransactionalBatchBuilder(maxCapacity int) TransactionBatchBuilder {
	return TransactionBatchBuilder{
		states:       make(map[check.ID]TransactionCheckInstanceBatchState),
		elementCount: 0,
		maxCapacity:  maxCapacity,
	}
}

func (builder *TransactionBatchBuilder) getOrCreateState(checkID check.ID, transactionID string) TransactionCheckInstanceBatchState {
	if value, ok := builder.states[checkID]; ok {
		return value
	}

	state := TransactionCheckInstanceBatchState{
		Transaction: &BatchTransaction{
			TransactionID: transactionID,
		},
		Topology: nil,
		Health:   make(map[string]health.Health),
		Metrics:  nil,
	}
	builder.states[checkID] = state
	return state
}

func (builder *TransactionBatchBuilder) getOrCreateTopology(checkID check.ID, transactionID string, instance topology.Instance) *topology.Topology {
	state := builder.getOrCreateState(checkID, transactionID)

	if state.Topology != nil {
		return state.Topology
	}

	builder.states[checkID] = TransactionCheckInstanceBatchState{
		Transaction: &BatchTransaction{
			TransactionID: transactionID,
		},
		Topology: &topology.Topology{
			StartSnapshot: false,
			StopSnapshot:  false,
			Instance:      instance,
			Components:    make([]topology.Component, 0),
			Relations:     make([]topology.Relation, 0),
			DeleteIDs:     make([]string, 0),
		},
		Health:  state.Health,
		Metrics: state.Metrics,
	}
	return builder.states[checkID].Topology
}

func (builder *TransactionBatchBuilder) getOrCreateHealth(checkID check.ID, transactionID string, stream health.Stream) health.Health {
	state := builder.getOrCreateState(checkID, transactionID)

	if value, ok := state.Health[stream.GoString()]; ok {
		return value
	}

	builder.states[checkID].Health[stream.GoString()] = health.Health{
		StartSnapshot: nil,
		StopSnapshot:  nil,
		Stream:        stream,
		CheckStates:   make([]health.CheckData, 0),
	}

	return builder.states[checkID].Health[stream.GoString()]
}

func (builder *TransactionBatchBuilder) getOrCreateRawMetrics(checkID check.ID, transactionID string) *telemetry.Metrics {
	state := builder.getOrCreateState(checkID, transactionID)

	if state.Metrics != nil {
		return state.Metrics
	}

	builder.states[checkID] = TransactionCheckInstanceBatchState{
		Transaction: &BatchTransaction{
			TransactionID: transactionID,
		},
		Topology: state.Topology,
		Health:   state.Health,
		Metrics:  &telemetry.Metrics{},
	}

	return builder.states[checkID].Metrics
}

// AddComponent adds a component
func (builder *TransactionBatchBuilder) AddComponent(checkID check.ID, transactionID string, instance topology.Instance, component topology.Component) TransactionCheckInstanceBatchStates {
	topologyData := builder.getOrCreateTopology(checkID, transactionID, instance)
	topologyData.Components = append(topologyData.Components, component)
	return builder.incrementAndTryFlush()
}

// AddRelation adds a relation
func (builder *TransactionBatchBuilder) AddRelation(checkID check.ID, transactionID string, instance topology.Instance, relation topology.Relation) TransactionCheckInstanceBatchStates {
	topologyData := builder.getOrCreateTopology(checkID, transactionID, instance)
	topologyData.Relations = append(topologyData.Relations, relation)
	return builder.incrementAndTryFlush()
}

// TopologyStartSnapshot starts a snapshot
func (builder *TransactionBatchBuilder) TopologyStartSnapshot(checkID check.ID, transactionID string, instance topology.Instance) TransactionCheckInstanceBatchStates {
	topologyData := builder.getOrCreateTopology(checkID, transactionID, instance)
	topologyData.StartSnapshot = true
	return nil
}

// TopologyStopSnapshot stops a snapshot. This will always flush
func (builder *TransactionBatchBuilder) TopologyStopSnapshot(checkID check.ID, transactionID string, instance topology.Instance) TransactionCheckInstanceBatchStates {
	topologyData := builder.getOrCreateTopology(checkID, transactionID, instance)
	topologyData.StopSnapshot = true
	return builder.incrementAndTryFlush()
}

// Delete deletes a topology element
func (builder *TransactionBatchBuilder) Delete(checkID check.ID, transactionID string, instance topology.Instance, deleteID string) TransactionCheckInstanceBatchStates {
	topologyData := builder.getOrCreateTopology(checkID, transactionID, instance)
	topologyData.DeleteIDs = append(topologyData.DeleteIDs, deleteID)
	return builder.incrementAndTryFlush()
}

// AddHealthCheckData adds a component
func (builder *TransactionBatchBuilder) AddHealthCheckData(checkID check.ID, transactionID string, stream health.Stream, data health.CheckData) TransactionCheckInstanceBatchStates {
	healthData := builder.getOrCreateHealth(checkID, transactionID, stream)
	healthData.CheckStates = append(healthData.CheckStates, data)
	builder.states[checkID].Health[stream.GoString()] = healthData
	return builder.incrementAndTryFlush()
}

// HealthStartSnapshot starts a Health snapshot
func (builder *TransactionBatchBuilder) HealthStartSnapshot(checkID check.ID, transactionID string, stream health.Stream, repeatIntervalSeconds int, expirySeconds int) TransactionCheckInstanceBatchStates {
	healthData := builder.getOrCreateHealth(checkID, transactionID, stream)
	healthData.StartSnapshot = &health.StartSnapshotMetadata{
		RepeatIntervalS: repeatIntervalSeconds,
		ExpiryIntervalS: expirySeconds,
	}
	builder.states[checkID].Health[stream.GoString()] = healthData
	return nil
}

// HealthStopSnapshot stops a Health snapshot. This will always flush
func (builder *TransactionBatchBuilder) HealthStopSnapshot(checkID check.ID, transactionID string, stream health.Stream) TransactionCheckInstanceBatchStates {
	healthData := builder.getOrCreateHealth(checkID, transactionID, stream)
	healthData.StopSnapshot = &health.StopSnapshotMetadata{}
	builder.states[checkID].Health[stream.GoString()] = healthData
	return builder.incrementAndTryFlush()
}

// AddRawMetricsData adds raw metric data
func (builder *TransactionBatchBuilder) AddRawMetricsData(checkID check.ID, transactionID string, rawMetric telemetry.RawMetrics) TransactionCheckInstanceBatchStates {
	rawMetricsData := builder.getOrCreateRawMetrics(checkID, transactionID)
	rawMetricsData.Values = append(rawMetricsData.Values, rawMetric)
	return builder.incrementAndTryFlush()
}

// Flush the collected data. Returning the data and wiping the current build up Topology
func (builder *TransactionBatchBuilder) Flush() TransactionCheckInstanceBatchStates {
	data := builder.states
	builder.states = make(map[check.ID]TransactionCheckInstanceBatchState)
	builder.elementCount = 0
	return data
}

func (builder *TransactionBatchBuilder) incrementAndTryFlush() TransactionCheckInstanceBatchStates {
	builder.elementCount = builder.elementCount + 1

	if builder.elementCount >= builder.maxCapacity {
		return builder.Flush()
	}

	return nil
}

// MarkTransactionComplete marks a transaction as complete and flushes the data if produced
func (builder *TransactionBatchBuilder) MarkTransactionComplete(checkID check.ID, transactionID string) TransactionCheckInstanceBatchStates {
	if state, ok := builder.states[checkID]; ok {
		if state.Transaction.TransactionID == transactionID {
			state.Transaction.CompletedTransaction = true
			return builder.Flush()
		}
	} else {
		/*
			We don't have any state for this check which means that it was flushed as a result of another check completing,
			the flush interval triggering a flush, etc.

			We have a couple of options
			1. Create a new state for this check + transaction and wait for it to be flushed with a new check completion
			  or the check interval, running the very likely risk of the handler starting a new transaction for this check
			  in the meantime.
			2. Complete this transaction immediately, assuming that the data has been successfully produced with the
			  previous payload, running the risk that the previous payload is still in the forwarder attempting to be
			  sent to StackState. In this case an Action would have been committed for this transaction already. If we
			  complete too early, the transaction will be failed and the check rolled back. This is not ideal, but
			  tolerable because we won't lose data.

			We have decided to go for option 2 which seems to be less likely to have undesired consequences.

			TODO: refactor this in such a way to avoid this whole scenario.
		*/
		transactionmanager.GetTransactionManager().CompleteTransaction(transactionID)
	}

	return nil
}

// FlushOnComplete checks whether the check produced data, if so, flush
func (builder *TransactionBatchBuilder) FlushOnComplete(checkID check.ID) TransactionCheckInstanceBatchStates {
	if _, ok := builder.states[checkID]; ok {
		return builder.Flush()
	}

	return nil
}
