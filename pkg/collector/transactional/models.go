package transactional

import (
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

// PayloadTransaction ...
type PayloadTransaction struct {
	ActionID             string
	CompletedTransaction bool
}

type IntakePayload struct {
	InternalHostname string              `json:"internalHostname"`
	Topologies       []topology.Topology `json:"topologies"`
	Health           []health.Health     `json:"health"`
	Metrics          []interface{}       `json:"metrics"`
}
