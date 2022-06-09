package health

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMarshallStream(t *testing.T) {
	data := Stream{
		Urn:       "urn",
		SubStream: "substr",
	}

	result, err := json.Marshal(data)
	assert.NoError(t, err)
	assert.Equal(t, `{"urn":"urn","sub_stream_id":"substr"}`, string(result))
}

func TestUnMarshallStream1(t *testing.T) {
	expected := Stream{
		Urn:       "urn",
		SubStream: "substr",
	}
	data := `{"urn":"urn","sub_stream_id":"substr"}`

	var result Stream
	err := json.Unmarshal([]byte(data), &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestUnMarshallStream2(t *testing.T) {
	expected := Stream{
		Urn:       "urn",
		SubStream: "substr",
	}
	data := `{"urn":"urn","sub_stream":"substr"}`

	var result Stream
	err := json.Unmarshal([]byte(data), &result)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
