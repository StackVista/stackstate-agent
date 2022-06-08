package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

// CheckAPI contains all the operations that can be done by an Agent Check. This acts as a proxy to forward data
// where it needs to go.
type CheckAPI interface {
	CheckTransactionalAPI
	CheckStateAPI
	CheckTopologyAPI
	CheckHealthAPI
	CheckTelemetryAPI
	CheckLifecycleAPI
}

// CheckTransactionalAPI contains all the transactionality operations for a check
type CheckTransactionalAPI interface {
	StartTransaction() string
	StopTransaction()
	SetStateTransactional(key, state string)
}

// StartTransaction is used to start a transaction to the input channel
type StartTransaction struct {
	CheckID       check.ID
	TransactionID string
}

// StopTransaction is used to stop the current transaction
type StopTransaction struct{}

// SubmitSetStateTransactional is used to submit a set state transactional for the current transaction
type SubmitSetStateTransactional struct {
	Key, State string
}

// CheckStateAPI contains all the state operations for a check
type CheckStateAPI interface {
	SetState(key, state string) error
	GetState(key string) string
}

// CheckTopologyAPI contains all the topology operations for a check
type CheckTopologyAPI interface {
	SubmitComponent(instance topology.Instance, component topology.Component)
	SubmitRelation(instance topology.Instance, relation topology.Relation)
	SubmitStartSnapshot(instance topology.Instance)
	SubmitStopSnapshot(instance topology.Instance)
	SubmitDelete(instance topology.Instance, topologyElementID string)
}

// SubmitComponent is used to submit a topology component for the current transaction
type SubmitComponent struct {
	Instance  topology.Instance
	Component topology.Component
}

// SubmitRelation is used to submit a topology relation for the current transaction
type SubmitRelation struct {
	Instance topology.Instance
	Relation topology.Relation
}

// SubmitStartSnapshot is used to submit a topology start snapshot for the current transaction
type SubmitStartSnapshot struct {
	Instance topology.Instance
}

// SubmitStopSnapshot is used to submit a topology stop snapshot for the current transaction
type SubmitStopSnapshot struct {
	Instance topology.Instance
}

// SubmitDelete is used to submit a topology delete for the current transaction
type SubmitDelete struct {
	Instance          topology.Instance
	TopologyElementID string
}

// CheckHealthAPI contains all the health state operations for a check
type CheckHealthAPI interface {
	SubmitHealthCheckData(stream health.Stream, data health.CheckData)
	SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int)
	SubmitHealthStopSnapshot(stream health.Stream)
}

// SubmitHealthCheckData is used to submit a health check data for the current transaction
type SubmitHealthCheckData struct {
	Stream health.Stream
	Data   health.CheckData
}

// SubmitHealthStartSnapshot is used to submit a health start snapshot for the current transaction
type SubmitHealthStartSnapshot struct {
	Stream                         health.Stream
	IntervalSeconds, ExpirySeconds int
}

// SubmitHealthStopSnapshot is used to submit a health stop snapshot for the current transaction
type SubmitHealthStopSnapshot struct {
	Stream health.Stream
}

// CheckTelemetryAPI contains all the telemetry operations for a check
type CheckTelemetryAPI interface {
	SubmitRawMetricsData(data telemetry.RawMetrics)
}

// SubmitRawMetric is used to submit a raw metric value for the current transaction
type SubmitRawMetric struct {
	Value telemetry.RawMetrics
}

// CheckLifecycleAPI contains all the lifecylce operations for a check
type CheckLifecycleAPI interface {
	SubmitComplete()
}

// SubmitComplete is used to submit a check complete for the current transaction
type SubmitComplete struct{}
