package health

type CheckData struct {
	UnstructuredCheckState
	*CheckState
	*CheckStateDeleted
}

func (c *CheckData) IsEmpty() bool {
	return c.CheckState == nil && c.CheckStateDeleted == nil && len(c.UnstructuredCheckState) == 0
}

// UnstructuredCheckState is the data for a health check - usually that comes from Python check
type UnstructuredCheckState map[string]interface{}

// CheckState describes state of a health stream
// see also:
// https://docs.stackstate.com/configure/health/send-health-data/repeat_snapshots
// https://docs.stackstate.com/configure/health/send-health-data/repeat_states
// https://docs.stackstate.com/configure/health/send-health-data/transactional_increments
type CheckState struct {
	CheckStateId              string `json:"checkStateId"`              // Identifier for the check state in the external system
	Message                   string `json:"message,omitempty"`         // Message to display in StackState UI. Data will be interpreted as markdown allowing to have links to the external system check that generated the external check state.
	Health                    State  `json:"health"`                    // StackState Health state
	TopologyElementIdentifier string `json:"topologyElementIdentifier"` // Used to bind the check state to a StackState topology element
	Name                      string `json:"name"`                      // Name of the external check state.
}

// CheckStateDeleted describes signals StackState to delete a health stream
type CheckStateDeleted struct {
	CheckStateId string `json:"checkStateId"` // Identifier for the check state in the external system
	Delete       bool   `json:"delete"`       // should be true
}
