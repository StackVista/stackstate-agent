package health

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshallUnstructured(t *testing.T) {
	data := CheckData{Unstructured: map[string]interface{}{
		"test": "test",
	}}

	result, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Equal(t, `{"test":"test"}`, string(result))
}

func TestMarshallStructuredState(t *testing.T) {
	data := CheckData{
		CheckState: &CheckState{
			CheckStateID:              "checkStateId",
			Message:                   "message",
			Health:                    Critical,
			TopologyElementIdentifier: "identifier",
			Name:                      "name",
		},
	}

	result, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Equal(t, `{"checkStateId":"checkStateId","message":"message","health":"CRITICAL","topologyElementIdentifier":"identifier","name":"name"}`, string(result))
}

func TestMarshallStructuredDeleteState(t *testing.T) {
	data := CheckData{
		CheckStateDeleted: &CheckStateDeleted{
			CheckStateID: "checkStateId",
			Delete:       true,
		},
	}

	result, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Equal(t, `{"checkStateId":"checkStateId","delete":true}`, string(result))
}
