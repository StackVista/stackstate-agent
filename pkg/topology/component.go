package topology

import (
	"encoding/json"
	"fmt"
)

// Component is a representation of a topology component
type Component struct {
	ExternalID string                 `json:"externalId"`
	Type       Type                   `json:"type"`
	Data       map[string]interface{} `json:"data"`
}

// Prints a JSON string of the Component
func (c Component) JsonString() string {
	b, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}
