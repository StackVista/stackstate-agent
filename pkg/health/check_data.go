package health

import (
	"encoding/json"
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// CheckData describes state of a health stream
// it's an enumeration - only one of the fields should be filled
// Unstructured is intended to be used to pass through Python health state data
// see other type's description
type CheckData struct {
	Unstructured      map[string]interface{}
	CheckState        *CheckState
	CheckStateDeleted *CheckStateDeleted
}

var _ json.Marshaler = CheckData{}

// MarshalJSON ensures one of the cases is rendered
func (c CheckData) MarshalJSON() ([]byte, error) {
	if c.CheckState != nil {
		return json.Marshal(c.CheckState)
	} else if c.CheckStateDeleted != nil {
		return json.Marshal(c.CheckStateDeleted)
	} else {
		return json.Marshal(c.Unstructured)
	}
}

// IsEmpty checks if the data is empty (can come from Python check)
func (c *CheckData) IsEmpty() bool {
	return c.CheckState == nil && c.CheckStateDeleted == nil && len(c.Unstructured) == 0
}

// UnmarshalJSON unmarshalls data from json
func (c *CheckData) UnmarshalJSON(buf []byte) error {
	unstructured := map[string]interface{}{}
	if err := json.Unmarshal(buf, &unstructured); err != nil {
		return err
	}
	c.Unstructured = unstructured
	return nil
}

// JSONString encodes input into JSON while also encoding an error - for logging purpose
func (c *CheckData) JSONString() string {
	b, err := c.MarshalJSON()
	if err != nil {
		_ = log.Warnf("Failed to serialize JSON: %v", err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}

// CheckState describes state of a health stream
// see also:
// https://docs.stackstate.com/configure/health/send-health-data/repeat_snapshots
// https://docs.stackstate.com/configure/health/send-health-data/repeat_states
// https://docs.stackstate.com/configure/health/send-health-data/transactional_increments
type CheckState struct {
	CheckStateID              string `json:"checkStateId"`              // Identifier for the check state in the external system
	Message                   string `json:"message,omitempty"`         // Message to display in StackState UI. Data will be interpreted as markdown allowing to have links to the external system check that generated the external check state.
	Health                    State  `json:"health"`                    // StackState Health state
	TopologyElementIdentifier string `json:"topologyElementIdentifier"` // Used to bind the check state to a StackState topology element
	Name                      string `json:"name"`                      // Name of the external check state.
}

// CheckStateDeleted describes signals StackState to delete a health stream
type CheckStateDeleted struct {
	CheckStateID string `json:"checkStateId"` // Identifier for the check state in the external system
	Delete       bool   `json:"delete"`       // should be true
}
