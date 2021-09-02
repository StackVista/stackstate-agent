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
		Data: expectedRawMetricsData,
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	checkId := C.CString("check-id")
	SubmitRawMetricsData(checkId, C.CString(string(data)))

	expectedState := mockBatcher.CollectedTopology.Flush()

	assert.Equal(t, batcher.CheckInstanceBatchStates(map[check.ID]batcher.CheckInstanceBatchState{
		"check-id": {
			Metrics: &metrics.RawMetrics{
				CheckStates: []metrics.RawMetricsCheckData{
					expectedRawMetricsData,
				},
			},
		},
	}), expectedState)
}
