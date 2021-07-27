package topology

//go:generate msgp

import (
	"encoding/json"
	"fmt"
)

// Relation is a representation of a topology relation
type Relation struct {
	ExternalID string `json:"externalId" msg:"externalId"`
	SourceID   string `json:"sourceId" msg:"sourceId"`
	TargetID   string `json:"targetId" msg:"targetId"`
	Type       Type   `json:"type" msg:"type"`
	Data       Data   `json:"data" msg:"data"`
}

// JSONString returns a JSON string of the Relation
func (r Relation) JSONString() string {
	b, err := json.Marshal(r)
	if err != nil {
		fmt.Println(err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}
