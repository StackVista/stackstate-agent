package handler

import (
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
	SetStateTransactional(key string, state string)
}

// CheckStateAPI contains all the state operations for a check
type CheckStateAPI interface {
	SetState(key string, state string) error
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

// CheckHealthAPI contains all the health state operations for a check
type CheckHealthAPI interface {
	SubmitHealthCheckData(stream health.Stream, data health.CheckData)
	SubmitHealthStartSnapshot(stream health.Stream, intervalSeconds int, expirySeconds int)
	SubmitHealthStopSnapshot(stream health.Stream)
}

// CheckTelemetryAPI contains all the telemetry operations for a check
type CheckTelemetryAPI interface {
	SubmitRawMetricsData(data telemetry.RawMetrics)
}

// CheckLifecycleAPI contains all the lifecylce operations for a check
type CheckLifecycleAPI interface {
	SubmitComplete()
}
