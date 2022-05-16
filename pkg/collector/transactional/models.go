package transactional

import (
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

const (
	IntakePath = "intake"
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

func NewIntakePayload() IntakePayload {
	return IntakePayload{
		Topologies: make([]topology.Topology, 0),
		Health:     make([]health.Health, 0),
		Metrics:    make([]interface{}, 0),
	}
}
