//go:build python && test

package python

import (
	"encoding/json"
	"github.com/DataDog/datadog-agent/pkg/aggregator/mocksender"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"github.com/DataDog/datadog-agent/pkg/collector/check/test"
	"github.com/DataDog/datadog-agent/pkg/metrics/event"
	check2 "github.com/StackVista/stackstate-receiver-go-client/pkg/model/check"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/health"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/telemetry"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testTopologyEvent(t *testing.T) {
	_, mockTransactionalBatcher, _, checkManager := SetupTransactionalComponents()

	testCheck := &test.STSTestCheck{Name: "check-id-topology-event-test"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	c := &event.Event{
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
		EventContext: &event.EventContext{
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
			SourceLinks: []event.SourceLink{
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

	checkId := C.CString(testCheck.String())

	StartTransaction(checkId)
	SubmitTopologyEvent(checkId, ev)

	expectedEvent := event.Event{
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
		EventContext: &event.EventContext{
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
			SourceLinks: []event.SourceLink{
				{Title: "source1_title", URL: "source1_url"},
				{Title: "source2_title", URL: "source2_url"},
			},
		},
	}

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
		Health:      map[string]health.Health{},
		Events:      &telemetry.IntakeEvents{Events: []telemetry.Event{handler.ConvertToStsEvent(expectedEvent)}},
	}
	assert.Equal(t, expectedTopology, actualTopology)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}

func testTopologyEventMissingFields(t *testing.T) {
	_, mockTransactionalBatcher, _, checkManager := SetupTransactionalComponents()

	testCheck := &test.STSTestCheck{Name: "check-id-topology-event-missing-fields-test"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	c := &event.Event{
		Title: "ev_title",
		Text:  "ev_text",
		Ts:    21,
		Host:  "ev_host",
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	ev := C.CString(string(data))

	checkId := C.CString(testCheck.String())

	StartTransaction(checkId)
	SubmitTopologyEvent(checkId, ev)

	expectedEvent := event.Event{
		Title: "ev_title",
		Text:  "ev_text",
		Ts:    21,
		Host:  "ev_host",
	}

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
		Health:      map[string]health.Health{},
		Events:      &telemetry.IntakeEvents{Events: []telemetry.Event{handler.ConvertToStsEvent(expectedEvent)}},
	}
	assert.Equal(t, expectedTopology, actualTopology)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}

func testTopologyEventWrongFieldType(t *testing.T) {
	_, mockTransactionalBatcher, _, checkManager := SetupTransactionalComponents()

	testCheck := &test.STSTestCheck{Name: "check-id-topology-event-wrong-field-type-test"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	sender := mocksender.NewMockSender("testID")
	sender.SetupAcceptAll()

	ev := C.CString(`{msg_title: 42}`)

	checkId := C.CString(testCheck.String())

	StartTransaction(checkId)
	SubmitTopologyEvent(checkId, ev)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
		Health:      map[string]health.Health{},
		Events:      nil, // assert we have no events
	}
	assert.Equal(t, expectedTopology, actualTopology)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}

var expectedRawMetricsData = telemetry.RawMetric{
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
	_, mockTransactionalBatcher, _, checkManager := SetupTransactionalComponents()

	testCheck := &test.STSTestCheck{Name: "check-id-raw-metrics"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(testCheck.String())
	name := C.CString(expectedRawMetricsData.Name)
	value := C.float(expectedRawMetricsData.Value)
	tags := []*C.char{C.CString("foo"), C.CString("bar"), nil}
	hostname := C.CString(expectedRawMetricsData.HostName)
	timestamp := C.longlong(expectedRawMetricsData.Timestamp)

	StartTransaction(checkId)
	SubmitRawMetricsData(checkId, name, value, &tags[0], hostname, timestamp)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
		Metrics:     &telemetry.Metrics{Values: []telemetry.RawMetric{expectedRawMetricsData}},
		Health:      map[string]health.Health{},
	}, actualTopology)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}
