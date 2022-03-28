package transactional

import (
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

// TxCheckInstanceBatchState is the type representing batched data per check instance
type TxCheckInstanceBatchState struct {
	*batcher.CheckInstanceBatchState
	*IntakeTransaction
}

// TransactionalBatchBuilder builds topology for a check instance
type TransactionalBatchBuilder struct {
	batchState *TxCheckInstanceBatchState
	// Count the amount of elements we gathered
	elementCount int
	// Amount of elements when we flush
	maxCapacity int
}

// MakeTransactionalBatchBuilder returns a instance of a TransactionalBatchBuilder
func MakeTransactionalBatchBuilder(maxCapacity int) *TransactionalBatchBuilder {
	return &TransactionalBatchBuilder{
		batchState:   &TxCheckInstanceBatchState{},
		elementCount: 0,
		maxCapacity:  maxCapacity,
	}
}

func (builder *TransactionalBatchBuilder) getOrCreateTopology(instance topology.Instance) *topology.Topology {
	if builder.batchState.Topology != nil {
		return builder.batchState.Topology
	}

	topo := &topology.Topology{
		StartSnapshot: false,
		StopSnapshot:  false,
		Instance:      instance,
		Components:    make([]topology.Component, 0),
		Relations:     make([]topology.Relation, 0),
	}
	builder.batchState.Topology = topo
	return topo
}

func (builder *TransactionalBatchBuilder) getOrCreateHealth(stream health.Stream) health.Health {
	if builder.batchState.Health != nil {
		if value, ok := builder.batchState.Health[stream.GoString()]; ok {
			return value
		}
	} else {
		builder.batchState.Health = map[string]health.Health{}
	}

	builder.batchState.Health[stream.GoString()] = health.Health{
		StartSnapshot: nil,
		StopSnapshot:  nil,
		Stream:        stream,
		CheckStates:   make([]health.CheckData, 0),
	}

	return builder.batchState.Health[stream.GoString()]
}

func (builder *TransactionalBatchBuilder) getOrCreateRawMetrics() *[]telemetry.RawMetrics {
	if builder.batchState.Metrics != nil {
		return builder.batchState.Metrics
	}

	builder.batchState.Metrics = &[]telemetry.RawMetrics{}

	return builder.batchState.Metrics
}

// AddComponent adds a component
func (builder *TransactionalBatchBuilder) AddComponent(instance topology.Instance, component topology.Component) *TxCheckInstanceBatchState {
	topologyData := builder.getOrCreateTopology(instance)
	topologyData.Components = append(topologyData.Components, component)
	return builder.incrementAndTryFlush()
}

// AddRelation adds a relation
func (builder *TransactionalBatchBuilder) AddRelation(instance topology.Instance, relation topology.Relation) *TxCheckInstanceBatchState {
	topologyData := builder.getOrCreateTopology(instance)
	topologyData.Relations = append(topologyData.Relations, relation)
	return builder.incrementAndTryFlush()
}

// StartSnapshot starts a snapshot
func (builder *TransactionalBatchBuilder) StartSnapshot(instance topology.Instance) *TxCheckInstanceBatchState {
	topologyData := builder.getOrCreateTopology(instance)
	topologyData.StartSnapshot = true
	return builder.incrementAndTryFlush()
}

// StopSnapshot stops a snapshot. This will always flush
func (builder *TransactionalBatchBuilder) StopSnapshot(instance topology.Instance) *TxCheckInstanceBatchState {
	topologyData := builder.getOrCreateTopology(instance)
	topologyData.StopSnapshot = true
	// We always flush after a StopSnapshot to limit latency
	return builder.Flush()
}

// AddHealthCheckData adds a component
func (builder *TransactionalBatchBuilder) AddHealthCheckData(stream health.Stream, data health.CheckData) *TxCheckInstanceBatchState {
	healthData := builder.getOrCreateHealth(stream)
	healthData.CheckStates = append(healthData.CheckStates, data)
	builder.batchState.Health[stream.GoString()] = healthData
	return builder.incrementAndTryFlush()
}

// HealthStartSnapshot starts a Health snapshot
func (builder *TransactionalBatchBuilder) HealthStartSnapshot(stream health.Stream, repeatIntervalSeconds int, expirySeconds int) *TxCheckInstanceBatchState {
	healthData := builder.getOrCreateHealth(stream)
	healthData.StartSnapshot = &health.StartSnapshotMetadata{
		RepeatIntervalS: repeatIntervalSeconds,
		ExpiryIntervalS: expirySeconds,
	}
	builder.batchState.Health[stream.GoString()] = healthData
	return nil
}

// HealthStopSnapshot stops a Health snapshot. This will always flush
func (builder *TransactionalBatchBuilder) HealthStopSnapshot(stream health.Stream) *TxCheckInstanceBatchState {
	healthData := builder.getOrCreateHealth(stream)
	healthData.StopSnapshot = &health.StopSnapshotMetadata{}
	builder.batchState.Health[stream.GoString()] = healthData
	// We always flush after a TopologyStopSnapshot to limit latency
	return builder.Flush()
}

// AddRawMetricsData adds raw metric data
func (builder *TransactionalBatchBuilder) AddRawMetricsData(rawMetric telemetry.RawMetrics) *TxCheckInstanceBatchState {
	rawMetricsData := builder.getOrCreateRawMetrics()
	*rawMetricsData = append(*rawMetricsData, rawMetric)
	return builder.incrementAndTryFlush()
}

// Flush the collected data. Returning the data and wiping the current build up topology
func (builder *TransactionalBatchBuilder) Flush() *TxCheckInstanceBatchState {
	data := builder.batchState
	builder.batchState = nil
	builder.elementCount = 0
	return data
}

func (builder *TransactionalBatchBuilder) incrementAndTryFlush() *TxCheckInstanceBatchState {
	builder.elementCount = builder.elementCount + 1

	if builder.elementCount >= builder.maxCapacity {
		return builder.Flush()
	}

	return nil
}

// FlushIfDataProduced checks whether the check produced data, if so, flush
func (builder *TransactionalBatchBuilder) FlushIfDataProduced() *TxCheckInstanceBatchState {
	if builder.batchState != nil {
		return builder.Flush()
	}

	return nil
}
