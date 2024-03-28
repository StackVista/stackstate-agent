package health

import (
	"encoding/json"
	"fmt"
)

// State health state of a health stream
type State string

const (
	// Clear designates healthy component
	Clear State = "CLEAR"
	// Deviating means a component is working but something is wrong[
	Deviating State = "DEVIATING"
	// Critical means that component is suffering
	Critical State = "CRITICAL"
)

// Health is a batch of health synchronization data
type Health struct {
	StartSnapshot *StartSnapshotMetadata `json:"start_snapshot,omitempty"`
	StopSnapshot  *StopSnapshotMetadata  `json:"stop_snapshot,omitempty"`
	Stream        Stream                 `json:"stream"`
	CheckStates   []CheckData            `json:"check_states"`
}

// Payload is a single payload for the batch of health synchronization data
type Payload struct {
	Stream Stream
	Data   CheckData
}

// JSONString returns a JSON string of the Payload
func (p Payload) JSONString() string {
	b, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}
