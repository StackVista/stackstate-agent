package transactionbatcher

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

// TransactionalBatcher interface can receive data for sending to the intake and will accumulate the data in batches. This does
// not work on a fixed schedule like the aggregator but flushes either when data exceeds a threshold, when
// data is complete.
type TransactionalBatcher interface {
	// Topology
	SubmitComponent(checkID check.ID, transactionID string, instance topology.Instance, component topology.Component)
	SubmitRelation(checkID check.ID, transactionID string, instance topology.Instance, relation topology.Relation)
	SubmitStartSnapshot(checkID check.ID, transactionID string, instance topology.Instance)
	SubmitStopSnapshot(checkID check.ID, transactionID string, instance topology.Instance)
	SubmitDelete(checkID check.ID, transactionID string, instance topology.Instance, topologyElementID string)

	// Health
	SubmitHealthCheckData(checkID check.ID, transactionID string, stream health.Stream, data health.CheckData)
	SubmitHealthStartSnapshot(checkID check.ID, transactionID string, stream health.Stream, intervalSeconds int, expirySeconds int)
	SubmitHealthStopSnapshot(checkID check.ID, transactionID string, stream health.Stream)

	// Raw Metrics
	SubmitRawMetricsData(checkID check.ID, transactionID string, data telemetry.RawMetrics)

	// lifecycle
	SubmitComplete(checkID check.ID)
	Shutdown()
}

// SubmitComponent is used to submit a component to the input channel
type SubmitComponent struct {
	CheckID       check.ID
	TransactionID string
	Instance      topology.Instance
	Component     topology.Component
}

// SubmitRelation is used to submit a relation to the input channel
type SubmitRelation struct {
	CheckID       check.ID
	TransactionID string
	Instance      topology.Instance
	Relation      topology.Relation
}

// SubmitStartSnapshot is used to submit a start of a snapshot to the input channel
type SubmitStartSnapshot struct {
	CheckID       check.ID
	TransactionID string
	Instance      topology.Instance
}

// SubmitStopSnapshot is used to submit a stop of a snapshot to the input channel
type SubmitStopSnapshot struct {
	CheckID       check.ID
	TransactionID string
	Instance      topology.Instance
}

// SubmitHealthCheckData is used to submit health check data to the input channel
type SubmitHealthCheckData struct {
	CheckID       check.ID
	TransactionID string
	Stream        health.Stream
	Data          health.CheckData
}

// SubmitHealthStartSnapshot is used to submit health check start snapshot to the input channel
type SubmitHealthStartSnapshot struct {
	CheckID         check.ID
	TransactionID   string
	Stream          health.Stream
	IntervalSeconds int
	ExpirySeconds   int
}

// SubmitHealthStopSnapshot is used to submit health check stop snapshot to the input channel
type SubmitHealthStopSnapshot struct {
	CheckID       check.ID
	TransactionID string
	Stream        health.Stream
}

// SubmitDelete is used to submit a topology delete to the input channel
type SubmitDelete struct {
	CheckID       check.ID
	TransactionID string
	Instance      topology.Instance
	DeleteID      string
}

// SubmitRawMetricsData is used to submit a raw metric value to the input channel
type SubmitRawMetricsData struct {
	CheckID       check.ID
	TransactionID string
	RawMetric     telemetry.RawMetrics
}

// SubmitComplete is used to submit a check run completion to the input channel
type SubmitComplete struct {
	CheckID check.ID
}

// SubmitShutdown is used to submit a shutdown of the transactionbatcher to the input channel
type SubmitShutdown struct{}
