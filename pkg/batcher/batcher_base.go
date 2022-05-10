package batcher

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// Batcher interface can receive data for sending to the intake and will accumulate the data in batches. This does
// not work on a fixed schedule like the aggregator but flushes either when data exceeds a threshold, when
// data is complete.
type Batcher interface {
	// Topology
	SubmitComponent(checkID check.ID, instance topology.Instance, component topology.Component)
	SubmitRelation(checkID check.ID, instance topology.Instance, relation topology.Relation)
	SubmitStartSnapshot(checkID check.ID, instance topology.Instance)
	SubmitStopSnapshot(checkID check.ID, instance topology.Instance)
	SubmitDelete(checkID check.ID, instance topology.Instance, topologyElementID string)

	// Health
	SubmitHealthCheckData(checkID check.ID, stream health.Stream, data health.CheckData)
	SubmitHealthStartSnapshot(checkID check.ID, stream health.Stream, intervalSeconds int, expirySeconds int)
	SubmitHealthStopSnapshot(checkID check.ID, stream health.Stream)

	// Raw Metrics
	SubmitRawMetricsData(checkID check.ID, data telemetry.RawMetrics)

	// lifecycle
	SubmitComplete(checkID check.ID)
	Shutdown()
}

// SubmitComponent is used to submit a component to the input channel
type SubmitComponent struct {
	CheckID   check.ID
	Instance  topology.Instance
	Component topology.Component
}

// SubmitRelation is used to submit a relation to the input channel
type SubmitRelation struct {
	CheckID  check.ID
	Instance topology.Instance
	Relation topology.Relation
}

// SubmitStartSnapshot is used to submit a start of a snapshot to the input channel
type SubmitStartSnapshot struct {
	CheckID  check.ID
	Instance topology.Instance
}

// SubmitStopSnapshot is used to submit a stop of a snapshot to the input channel
type SubmitStopSnapshot struct {
	CheckID  check.ID
	Instance topology.Instance
}

// SubmitHealthCheckData is used to submit health check data to the input channel
type SubmitHealthCheckData struct {
	CheckID check.ID
	Stream  health.Stream
	Data    health.CheckData
}

// SubmitHealthStartSnapshot is used to submit health check start snapshot to the input channel
type SubmitHealthStartSnapshot struct {
	CheckID         check.ID
	Stream          health.Stream
	IntervalSeconds int
	ExpirySeconds   int
}

// SubmitHealthStopSnapshot is used to submit health check stop snapshot to the input channel
type SubmitHealthStopSnapshot struct {
	CheckID check.ID
	Stream  health.Stream
}

// SubmitDelete is used to submit a topology delete to the input channel
type SubmitDelete struct {
	CheckID  check.ID
	Instance topology.Instance
	DeleteID string
}

// SubmitRawMetricsData is used to submit a raw metric value to the input channel
type SubmitRawMetricsData struct {
	CheckID   check.ID
	RawMetric telemetry.RawMetrics
}

// SubmitComplete is used to submit a check run completion to the input channel
type SubmitComplete struct {
	CheckID check.ID
}

// SubmitShutdown is used to submit a shutdown of the transactionbatcher to the input channel
type SubmitShutdown struct{}

// BatcherBase contains the base functionality of a transactionbatcher that can be extending with additional functionality
type BatcherBase struct {
	Hostname, agentName string
	Input               chan interface{}
}

// MakeBatcherBase creates a transactionbatcher base Instance
func MakeBatcherBase(hostname, agentName string, maxCapacity int) BatcherBase {
	return BatcherBase{
		Hostname:  hostname,
		agentName: agentName,
		Input:     make(chan interface{}, maxCapacity+1),
	}
}

// SubmitComponent submits a component to the batch
func (batcher BatcherBase) SubmitComponent(checkID check.ID, instance topology.Instance, component topology.Component) {
	batcher.Input <- SubmitComponent{
		CheckID:   checkID,
		Instance:  instance,
		Component: component,
	}
}

// SubmitRelation submits a relation to the batch
func (batcher BatcherBase) SubmitRelation(checkID check.ID, instance topology.Instance, relation topology.Relation) {
	batcher.Input <- SubmitRelation{
		CheckID:  checkID,
		Instance: instance,
		Relation: relation,
	}
}

// SubmitStartSnapshot submits start of a snapshot
func (batcher BatcherBase) SubmitStartSnapshot(checkID check.ID, instance topology.Instance) {
	batcher.Input <- SubmitStartSnapshot{
		CheckID:  checkID,
		Instance: instance,
	}
}

// SubmitStopSnapshot submits a stop of a snapshot. This always causes a flush of the data downstream
func (batcher BatcherBase) SubmitStopSnapshot(checkID check.ID, instance topology.Instance) {
	batcher.Input <- SubmitStopSnapshot{
		CheckID:  checkID,
		Instance: instance,
	}
}

// SubmitDelete submits a deletion of topology element.
func (batcher BatcherBase) SubmitDelete(checkID check.ID, instance topology.Instance, topologyElementID string) {
	batcher.Input <- SubmitDelete{
		CheckID:  checkID,
		Instance: instance,
		DeleteID: topologyElementID,
	}
}

// SubmitHealthCheckData submits a Health check data record to the batch
func (batcher BatcherBase) SubmitHealthCheckData(checkID check.ID, stream health.Stream, data health.CheckData) {
	log.Debugf("Submitting Health check data for check [%s] stream [%s]: %s", checkID, stream.GoString(), data.JSONString())
	batcher.Input <- SubmitHealthCheckData{
		CheckID: checkID,
		Stream:  stream,
		Data:    data,
	}
}

// SubmitHealthStartSnapshot submits start of a Health snapshot
func (batcher BatcherBase) SubmitHealthStartSnapshot(checkID check.ID, stream health.Stream, intervalSeconds int, expirySeconds int) {
	batcher.Input <- SubmitHealthStartSnapshot{
		CheckID:         checkID,
		Stream:          stream,
		IntervalSeconds: intervalSeconds,
		ExpirySeconds:   expirySeconds,
	}
}

// SubmitHealthStopSnapshot submits a stop of a Health snapshot. This always causes a flush of the data downstream
func (batcher BatcherBase) SubmitHealthStopSnapshot(checkID check.ID, stream health.Stream) {
	batcher.Input <- SubmitHealthStopSnapshot{
		CheckID: checkID,
		Stream:  stream,
	}
}

// SubmitRawMetricsData submits a raw metrics data record to the batch
func (batcher BatcherBase) SubmitRawMetricsData(checkID check.ID, rawMetric telemetry.RawMetrics) {
	if rawMetric.HostName == "" {
		rawMetric.HostName = batcher.Hostname
	}

	batcher.Input <- SubmitRawMetricsData{
		CheckID:   checkID,
		RawMetric: rawMetric,
	}
}

// SubmitComplete signals completion of a check. May trigger a flush only if the check produced data
func (batcher BatcherBase) SubmitComplete(checkID check.ID) {
	log.Debugf("Submitting complete for check [%s]", checkID)
	batcher.Input <- SubmitComplete{
		CheckID: checkID,
	}
}

// Shutdown shuts down the transactionbatcher
func (batcher BatcherBase) Shutdown() {
	batcher.Input <- SubmitShutdown{}
}
