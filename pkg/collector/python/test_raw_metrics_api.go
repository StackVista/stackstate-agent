// +build python,test

package python

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO: Raw Metrics

// #include <datadog_agent_rtloader.h>
import "C"

var expectedRawMetricsData = metrics.RawMetricsCheckData{
	"key":        "value ®",
	"stringlist": []interface{}{"a", "b", "c"},
	"boollist":   []interface{}{true, false},
	"intlist":    []interface{}{float64(1)},
	"doublelist": []interface{}{0.7, 1.42},
	"emptykey":   nil,
	"nestedobject": map[string]interface{}{
		"nestedkey": "nestedValue",
		"animals": map[string]interface{}{
			"legs":  "dog",
			"wings": "eagle",
			"tail":  "crocodile",
		},
	},
}

func testRawMetricsData(t *testing.T) {
	mockBatcher := batcher.NewMockBatcher()

	c := &metrics.RawMetricsPayload{
		Stream: metrics.RawMetricsStream{
			Urn:       "myurn",
			SubStream: "substream",
		},
		Data: expectedRawMetricsData,
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	checkId := C.CString("check-id")
	stream := C.raw_metrics_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")
	SubmitRawMetricsStartSnapshot(checkId, &stream)
	SubmitRawMetricsData(
		checkId,
		&stream,
		C.CString(string(data)))
	SubmitRawMetricsStopSnapshot(checkId, &stream)

	expectedState := mockBatcher.CollectedTopology.Flush()
	expectedStream := metrics.RawMetricsStream{Urn: "myurn", SubStream: "substream"}

	assert.Equal(t, batcher.CheckInstanceBatchStates(map[check.ID]batcher.CheckInstanceBatchState{
		"check-id": {
			Metrics: map[string]metrics.RawMetrics{
				expectedStream.GoString(): {
					Stream:        expectedStream,
					CheckStates: []metrics.RawMetricsCheckData{
						expectedRawMetricsData,
					},
				},
			},
		},
	}), expectedState)
}

func testRawMetricsStartSnapshot(t *testing.T) {
	mockBatcher := batcher.NewMockBatcher()

	checkId := C.CString("check-id")
	stream := C.raw_metrics_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")
	SubmitRawMetricsStartSnapshot(checkId, &stream)

	expectedState := mockBatcher.CollectedTopology.Flush()
	expectedStream := metrics.RawMetricsStream{Urn: "myurn", SubStream: "substream"}

	assert.Equal(t, batcher.CheckInstanceBatchStates(map[check.ID]batcher.CheckInstanceBatchState{
		"check-id": {
			Metrics: map[string]metrics.RawMetrics{
				expectedStream.GoString(): {
					Stream:        expectedStream,
					CheckStates:   []metrics.RawMetricsCheckData{},
				},
			},
		},
	}), expectedState)
}

func testRawMetricsStopSnapshot(t *testing.T) {
	mockBatcher := batcher.NewMockBatcher()

	checkId := C.CString("check-id")
	stream := C.raw_metrics_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")
	SubmitRawMetricsStopSnapshot(checkId, &stream)

	expectedState := mockBatcher.CollectedTopology.Flush()
	expectedStream := metrics.RawMetricsStream{Urn: "myurn", SubStream: "substream"}

	assert.Equal(t, batcher.CheckInstanceBatchStates(map[check.ID]batcher.CheckInstanceBatchState{
		"check-id": {
			Metrics: map[string]metrics.RawMetrics{
				expectedStream.GoString(): {
					Stream:       expectedStream,
					CheckStates:  []metrics.RawMetricsCheckData{},
				},
			},
		},
	}), expectedState)
}

func testNoRawMetricsSubStream(t *testing.T) {
	mockBatcher := batcher.NewMockBatcher()

	checkId := C.CString("check-id")
	stream := C.raw_metrics_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("")
	SubmitRawMetricsStartSnapshot(checkId, &stream)

	expectedState := mockBatcher.CollectedTopology.Flush()
	expectedStream := metrics.RawMetricsStream{Urn: "myurn"}

	assert.Equal(t, batcher.CheckInstanceBatchStates(map[check.ID]batcher.CheckInstanceBatchState{
		"check-id": {
			Metrics: map[string]metrics.RawMetrics{
				expectedStream.GoString(): {
					Stream:        expectedStream,
					CheckStates:   []metrics.RawMetricsCheckData{},
				},
			},
		},
	}), expectedState)
}
