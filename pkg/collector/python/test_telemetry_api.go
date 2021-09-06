// +build python,test

package python

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/aggregator/mocksender"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/stretchr/testify/assert"
	"testing"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testTopologyEvent(t *testing.T) {
	sender := mocksender.NewMockSender("testID")
	sender.SetupAcceptAll()
	c := &metrics.Event{
		Title:          "ev_title",
		Text:           "ev_text",
		Ts:             21,
		Priority:       "ev_priority",
		Host:           "ev_host",
		Tags:           []string{"tag1", "tag2"},
		AlertType:      "alert_type",
		AggregationKey: "aggregation_key",
		SourceTypeName: "source_type",
		EventType:      "event_type",
		EventContext: &metrics.EventContext{
			SourceIdentifier:   "ctx_source_id",
			ElementIdentifiers: []string{"ctx_elem_id1", "ctx_elem_id2"},
			Source:             "ctx_source",
			Category:           "ctx_category",
			Data: map[string]interface{}{
				"nestedobject": map[string]interface{}{
					"nestedkey": "nestedValue",
					"animals": map[string]interface{}{
						"legs":  "dog",
						"wings": "eagle",
						"tail":  "crocodile",
					},
				},
			},
			SourceLinks: []metrics.SourceLink{
				{
					Title: "source1_title",
					URL:   "source1_url",
				},
				{
					Title: "source2_title",
					URL:   "source2_url",
				},
			},
		},
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	ev := C.CString(string(data))

	SubmitTopologyEvent(C.CString("testID"), ev)

	expectedEvent := metrics.Event{
		Title:          "ev_title",
		Text:           "ev_text",
		Ts:             21,
		Priority:       "ev_priority",
		Host:           "ev_host",
		Tags:           []string{"tag1", "tag2"},
		AlertType:      "alert_type",
		AggregationKey: "aggregation_key",
		SourceTypeName: "source_type",
		EventType:      "event_type",
		EventContext: &metrics.EventContext{
			SourceIdentifier:   "ctx_source_id",
			ElementIdentifiers: []string{"ctx_elem_id1", "ctx_elem_id2"},
			Source:             "ctx_source",
			Category:           "ctx_category",
			Data: map[string]interface{}{
				"nestedobject": map[string]interface{}{
					"nestedkey": "nestedValue",
					"animals": map[string]interface{}{
						"legs":  "dog",
						"wings": "eagle",
						"tail":  "crocodile",
					},
				},
			},
			SourceLinks: []metrics.SourceLink{
				{Title: "source1_title", URL: "source1_url"},
				{Title: "source2_title", URL: "source2_url"},
			},
		},
	}
	for _, event := range sender.SentEvents {
		_ = event.String()
	}
	sender.AssertEvent(t, expectedEvent, 0)
}

var expectedRawMetricsData = telemetry.RawMetricsCheckData{
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

func testTopologyEventMissingFields(t *testing.T) {
	sender := mocksender.NewMockSender("testID")
	sender.SetupAcceptAll()

	c := &metrics.Event{
		Title: "ev_title",
		Text:  "ev_text",
		Ts:    21,
		Host:  "ev_host",
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	ev := C.CString(string(data))

	SubmitTopologyEvent(C.CString("testID"), ev)

	expectedEvent := metrics.Event{
		Title:    "ev_title",
		Text:     "ev_text",
		Ts:       21,
		Host:     "ev_host",
	}
	sender.AssertEvent(t, expectedEvent, 0)
}

func testTopologyEventWrongFieldType(t *testing.T) {
	sender := mocksender.NewMockSender("testID")
	sender.SetupAcceptAll()

	ev := C.CString(`{msg_title: 42}`)

	SubmitTopologyEvent(C.CString("testID"), ev)

	sender.AssertNotCalled(t, "Event")
}

func testRawMetricsData(t *testing.T) {
	mockBatcher := batcher.NewMockBatcher()

	c := &telemetry.RawMetricsPayload{
		Data: expectedRawMetricsData,
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	checkId := C.CString("check-id")
	SubmitRawMetricsData(checkId, C.CString(string(data)))

	expectedState := mockBatcher.CollectedTopology.Flush()

	assert.Equal(t, batcher.CheckInstanceBatchStates(map[check.ID]batcher.CheckInstanceBatchState{
		"check-id": {
			Metrics: &telemetry.RawMetrics{
				CheckStates: []telemetry.RawMetricsCheckData{
					expectedRawMetricsData,
				},
			},
		},
	}), expectedState)
}
