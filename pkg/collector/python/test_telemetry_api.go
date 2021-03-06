// +build python,test

package python

import (
	"github.com/StackVista/stackstate-agent/pkg/aggregator/mocksender"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
	"testing"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testTopologyEvent(t *testing.T) {
	sender := mocksender.NewMockSender("testID")
	sender.SetupAcceptAll()

	ev := C.CString(`
msg_title: ev_title
msg_text: ev_text
timestamp: 21
priority: ev_priority
host: ev_host
tags:
  - tag1
  - tag2
alert_type: alert_type
aggregation_key: aggregation_key
source_type_name: source_type
event_type: event_type
context:
  source_identifier: ctx_source_id
  element_identifiers:
    - ctx_elem_id1
    - ctx_elem_id2
  source: ctx_source
  category: ctx_category
  data:
    nestedobject:
      nestedkey: nestedValue
      animals:
        legs: dog
        wings: eagle
        tail: crocodile
  source_links:
    - title: source1_title
      url: source1_url
    - title: source2_title
      url: source2_url
`)

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

	ev := C.CString(`xxx: yyy`)

	SubmitTopologyEvent(C.CString("testID"), ev)

	sender.AssertEvent(t, metrics.Event{}, 0)
}

func testTopologyEventWrongFieldType(t *testing.T) {
	sender := mocksender.NewMockSender("testID")
	sender.SetupAcceptAll()

	ev := C.CString(`msg_title: 42`)

	SubmitTopologyEvent(C.CString("testID"), ev)

	sender.AssertNotCalled(t, "Event")
}
