package batcher

import (
	"encoding/json"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// CheckInstanceBatchState is the type representing batched data per check instance
type CheckInstanceBatchState struct {
	Topology *topology.Topology
	Metrics  *[]telemetry.RawMetrics
	Health   map[string]health.Health
}

// CheckInstanceBatchStates is the type representing batched data for all check instances
type CheckInstanceBatchStates map[check.ID]CheckInstanceBatchState

// BatchBuilder is a helper class to build Topology based on submitted data, this data structure is not thread safe
type BatchBuilder struct {
	states CheckInstanceBatchStates
	// Count the amount of elements we gathered
	elementCount int
	// Amount of elements when we flush
	maxCapacity int
	// Disable flushing data when topology/health snapshot stop signal is received
	noForceFlash bool
	// Maximum message size
	maxBatchSize int
	// Current message size
	currentBatchSize int
}

// NewBatchBuilder constructs a BatchBuilder
func NewBatchBuilder(maxCapacity int) BatchBuilder {
	return BatchBuilder{
		states:           make(map[check.ID]CheckInstanceBatchState),
		elementCount:     0,
		maxCapacity:      maxCapacity,
		noForceFlash:     false,
		currentBatchSize: 0,
		maxBatchSize:     config.GetBatcherMaxMessageSize(),
	}
}

// DisabledForceFlush disables flushing data when topology/health snapshot stop signal is received
func (builder *BatchBuilder) DisabledForceFlush() {
	builder.noForceFlash = true
}

func (builder *BatchBuilder) getOrCreateState(checkID check.ID) CheckInstanceBatchState {
	if value, ok := builder.states[checkID]; ok {
		return value
	}

	state := CheckInstanceBatchState{
		Topology: nil,
		Health:   make(map[string]health.Health),
		Metrics:  nil,
	}
	builder.states[checkID] = state
	return state
}

func (builder *BatchBuilder) getOrCreateTopology(checkID check.ID, instance topology.Instance) *topology.Topology {
	state := builder.getOrCreateState(checkID)

	if state.Topology != nil {
		return state.Topology
	}

	builder.states[checkID] = CheckInstanceBatchState{
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

func (builder *BatchBuilder) getOrCreateHealth(checkID check.ID, stream health.Stream) health.Health {
	state := builder.getOrCreateState(checkID)

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

func (builder *BatchBuilder) getOrCreateRawMetrics(checkID check.ID) *[]telemetry.RawMetrics {
	state := builder.getOrCreateState(checkID)

	if state.Metrics != nil {
		return state.Metrics
	}

	builder.states[checkID] = CheckInstanceBatchState{
		Topology: state.Topology,
		Health:   state.Health,
		Metrics:  &[]telemetry.RawMetrics{},
	}

	return builder.states[checkID].Metrics
}

// AddComponent adds a component
func (builder *BatchBuilder) AddComponent(checkID check.ID, instance topology.Instance, component topology.Component) CheckInstanceBatchStates {
	return builder.tryIncrementAndFlush(checkID, component, func() {
		topologyData := builder.getOrCreateTopology(checkID, instance)
		topologyData.Components = append(topologyData.Components, component)
	})
}

// AddRelation adds a relation
func (builder *BatchBuilder) AddRelation(checkID check.ID, instance topology.Instance, relation topology.Relation) CheckInstanceBatchStates {
	return builder.tryIncrementAndFlush(checkID, relation, func() {
		topologyData := builder.getOrCreateTopology(checkID, instance)
		topologyData.Relations = append(topologyData.Relations, relation)
	})
}

// TopologyStartSnapshot starts a snapshot
func (builder *BatchBuilder) TopologyStartSnapshot(checkID check.ID, instance topology.Instance) CheckInstanceBatchStates {
	topologyData := builder.getOrCreateTopology(checkID, instance)
	topologyData.StartSnapshot = true
	return nil
}

// TopologyStopSnapshot stops a snapshot. This will always flush
func (builder *BatchBuilder) TopologyStopSnapshot(checkID check.ID, instance topology.Instance) CheckInstanceBatchStates {
	topologyData := builder.getOrCreateTopology(checkID, instance)
	topologyData.StopSnapshot = true
	// We always flush after a TopologyStopSnapshot to limit latency
	if !builder.noForceFlash {
		return builder.Flush()
	}
	return nil
}

// Delete deletes a topology element
func (builder *BatchBuilder) Delete(checkID check.ID, instance topology.Instance, deleteID string) CheckInstanceBatchStates {
	return builder.tryIncrementAndFlush(checkID, deleteID, func() {
		topologyData := builder.getOrCreateTopology(checkID, instance)
		topologyData.DeleteIDs = append(topologyData.DeleteIDs, deleteID)
	})
}

// AddHealthCheckData adds a component
func (builder *BatchBuilder) AddHealthCheckData(checkID check.ID, stream health.Stream, data health.CheckData) CheckInstanceBatchStates {
	return builder.tryIncrementAndFlush(checkID, data, func() {
		healthData := builder.getOrCreateHealth(checkID, stream)
		healthData.CheckStates = append(healthData.CheckStates, data)
		builder.states[checkID].Health[stream.GoString()] = healthData
	})
}

// HealthStartSnapshot starts a Health snapshot
func (builder *BatchBuilder) HealthStartSnapshot(checkID check.ID, stream health.Stream, repeatIntervalSeconds int, expirySeconds int) CheckInstanceBatchStates {
	healthData := builder.getOrCreateHealth(checkID, stream)
	healthData.StartSnapshot = &health.StartSnapshotMetadata{
		RepeatIntervalS: repeatIntervalSeconds,
		ExpiryIntervalS: expirySeconds,
	}
	builder.states[checkID].Health[stream.GoString()] = healthData
	return nil
}

// HealthStopSnapshot stops a Health snapshot. This will always flush
func (builder *BatchBuilder) HealthStopSnapshot(checkID check.ID, stream health.Stream) CheckInstanceBatchStates {
	healthData := builder.getOrCreateHealth(checkID, stream)
	healthData.StopSnapshot = &health.StopSnapshotMetadata{}
	builder.states[checkID].Health[stream.GoString()] = healthData
	// We always flush after a TopologyStopSnapshot to limit latency
	if !builder.noForceFlash {
		return builder.Flush()
	}
	return nil
}

// AddRawMetricsData adds raw metric data
func (builder *BatchBuilder) AddRawMetricsData(checkID check.ID, rawMetric telemetry.RawMetrics) CheckInstanceBatchStates {
	return builder.tryIncrementAndFlush(checkID, rawMetric, func() {
		rawMetricsData := builder.getOrCreateRawMetrics(checkID)
		*rawMetricsData = append(*rawMetricsData, rawMetric)
	})
}

// Flush the collected data. Returning the data and wiping the current build up Topology
func (builder *BatchBuilder) Flush() CheckInstanceBatchStates {
	log.Infof("Flushing batch with %d elements and size %d", builder.elementCount, builder.currentBatchSize)
	data := builder.states
	builder.states = make(map[check.ID]CheckInstanceBatchState)
	builder.elementCount = 0
	builder.currentBatchSize = 0
	return data
}

func getElementSize(element interface{}) (int, error) {
	jsonData, err := json.Marshal(element)
	if err != nil {
		return 0, fmt.Errorf("unable to marshal element %+v", element)
	}
	return len(jsonData), nil
}

func (builder *BatchBuilder) tryIncrementAndFlush(checkID check.ID, element interface{}, getAndAppend func()) CheckInstanceBatchStates {
	elementSize, err := getElementSize(element)
	if err != nil {
		log.Errorf("Could not get size of element from check %s. The element will be ignored: %v", checkID, err)
		return nil
	}
	if elementSize > builder.maxBatchSize {
		log.Errorf("Element size (%d) is bigger than maximum batch size (%d). The element will be ignored", elementSize, builder.maxBatchSize)
		return nil
	} else if builder.currentBatchSize+elementSize > builder.maxBatchSize {
		data := builder.Flush()
		getAndAppend()
		builder.increment(elementSize)
		return data
	} else {
		builder.increment(elementSize)
		if builder.elementCount >= builder.maxCapacity {
			return builder.Flush()
		}
	}
	return nil
}

func (builder *BatchBuilder) increment(elementSize int) {
	builder.elementCount += 1
	builder.currentBatchSize += elementSize
}

func (builder *BatchBuilder) incrementAndTryFlush() CheckInstanceBatchStates {
	builder.elementCount = builder.elementCount + 1

	if builder.elementCount >= builder.maxCapacity {
		return builder.Flush()
	}

	return nil
}

// FlushIfDataProduced checks whether the check produced data, if so, flush
func (builder *BatchBuilder) FlushIfDataProduced(checkID check.ID) CheckInstanceBatchStates {
	if _, ok := builder.states[checkID]; ok {
		return builder.Flush()
	}

	return nil
}
