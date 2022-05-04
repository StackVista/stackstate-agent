package util

import (
	"encoding/json"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// JSONString encodes input into JSON while also encoding an error - for logging purpose
func JSONString(c interface{}) string {
	b, err := json.Marshal(c)
	if err != nil {
		_ = log.Warnf("Failed to serialize JSON: %v", err)
		return fmt.Sprintf("{\"error\": \"%s\"}", err.Error())
	}
	return string(b)
}
