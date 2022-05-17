package transactional

import (
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/topology"
)

const (
	IntakePath = "intake"
)

// PayloadTransaction is used to keep track of a given actionID and completion status of a transaction when submitting
// payloads
type PayloadTransaction struct {
	ActionID             string
	CompletedTransaction bool
}

// IntakePayload is a Go representation of the Receiver Intake structure
type IntakePayload struct {
	InternalHostname string              `json:"internalHostname"`
	Topologies       []topology.Topology `json:"topologies"`
	Health           []health.Health     `json:"health"`
	Metrics          []interface{}       `json:"metrics"`
}

// NewIntakePayload returns a IntakePayload with default values
func NewIntakePayload() IntakePayload {
	return IntakePayload{
		Topologies: make([]topology.Topology, 0),
		Health:     make([]health.Health, 0),
		Metrics:    make([]interface{}, 0),
	}
}
