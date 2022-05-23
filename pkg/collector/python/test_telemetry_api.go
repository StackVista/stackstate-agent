//go:build python && test
// +build python,test

package python

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/aggregator/mocksender"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/checkmanager"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/handler"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/health"
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
		Title: "ev_title",
		Text:  "ev_text",
		Ts:    21,
		Host:  "ev_host",
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

var expectedRawMetricsData = telemetry.RawMetrics{
	Name:      "name",
	Timestamp: 123456,
	HostName:  "hostname",
	Value:     10,
	Tags: []string{
		"foo",
		"bar",
	},
}

func testRawMetricsData(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	checkId := C.CString("check-id")
	name := C.CString(expectedRawMetricsData.Name)
	value := C.float(expectedRawMetricsData.Value)
	tags := []*C.char{C.CString("foo"), C.CString("bar"), nil}
	hostname := C.CString(expectedRawMetricsData.HostName)
	timestamp := C.longlong(expectedRawMetricsData.Timestamp)

	SubmitRawMetricsData(checkId, name, value, &tags[0], hostname, timestamp)

	actualState := mockTransactionalBatcher.CollectedTopology.Flush()

	assert.Exactly(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Transaction: &transactionbatcher.BatchTransaction{
					TransactionID:        "", // no start transaction, so the transaction is empty in this case
					CompletedTransaction: false,
				},
				Metrics: &telemetry.Metrics{Values: []telemetry.RawMetrics{expectedRawMetricsData}},
				Health:  map[string]health.Health{},
			},
		}), actualState)
}
