package aws

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpanInterpreterEngine(t *testing.T) {
	for _, tc := range []struct {
		testCase string
		span     *pb.Span
		expected *pb.Span
		kind     string
		event    string
		resource string
		sType    string
		layer    string
		domain   string
		urn      string
		id       string
	}{
		{
			testCase: "A empty span ran through the AWS Span Builder should supply meta data",
			kind:     "kind-value",
			event:    "event-value",
			resource: "resource-value",
			sType:    "type-value",
			layer:    "layer-value",
			domain:   "domain-value",
			urn:      "urn:service:/hostname",
			id:       "id-value",
			span: &pb.Span{
				Meta: map[string]string{},
			},
			expected: &pb.Span{
				Service:  "open.telemetry.resource-value",
				Resource: "aws.resource-value",
				Type:     "type-value",
				Meta: map[string]string{
					"domain":                  "domain-value",
					"layer":                   "layer-value",
					"service":                 "open.telemetry.resource-value",
					"span.kind":               "kind-value",
					"span.serviceName":        "open.telemetry.resource-value.event-value",
					"span.serviceType":        "type-value",
					"span.serviceURN":         "urn:service:/hostname",
					"sts.service.identifiers": "id-value",
				},
			},
		},
		{
			testCase: "Span containing meta data should contain the new and original meta data through the AWS Span Builder",
			kind:     "kind-value",
			event:    "event-value",
			resource: "resource-value",
			sType:    "type-value",
			layer:    "layer-value",
			domain:   "domain-value",
			urn:      "urn:service:/hostname",
			id:       "id-value",
			span: &pb.Span{
				Meta: map[string]string{
					"extra-item-a": "value-a",
					"extra-item-b": "value-b",
				},
			},
			expected: &pb.Span{
				Service:  "open.telemetry.resource-value",
				Resource: "aws.resource-value",
				Type:     "type-value",
				Meta: map[string]string{
					"extra-item-a":            "value-a",
					"extra-item-b":            "value-b",
					"domain":                  "domain-value",
					"layer":                   "layer-value",
					"service":                 "open.telemetry.resource-value",
					"span.kind":               "kind-value",
					"span.serviceName":        "open.telemetry.resource-value.event-value",
					"span.serviceType":        "type-value",
					"span.serviceURN":         "urn:service:/hostname",
					"sts.service.identifiers": "id-value",
				},
			},
		},
		{
			testCase: "AWS Span Builder should overwrite existing data if it was predefined",
			kind:     "kind-value",
			event:    "event-value",
			resource: "resource-value",
			sType:    "type-value",
			layer:    "layer-value",
			domain:   "domain-value",
			urn:      "urn:service:/hostname",
			id:       "id-value",
			span: &pb.Span{
				Meta: map[string]string{
					"domain":                  "value-should-be-ignored",
					"layer":                   "value-should-be-ignored",
					"service":                 "value-should-be-ignored",
					"span.kind":               "value-should-be-ignored",
					"span.serviceName":        "value-should-be-ignored",
					"span.serviceType":        "value-should-be-ignored",
					"span.serviceURN":         "value-should-be-ignored",
					"sts.service.identifiers": "value-should-be-ignored",
				},
			},
			expected: &pb.Span{
				Service:  "open.telemetry.resource-value",
				Resource: "aws.resource-value",
				Type:     "type-value",
				Meta: map[string]string{
					"domain":                  "domain-value",
					"layer":                   "layer-value",
					"service":                 "open.telemetry.resource-value",
					"span.kind":               "kind-value",
					"span.serviceName":        "open.telemetry.resource-value.event-value",
					"span.serviceType":        "type-value",
					"span.serviceURN":         "urn:service:/hostname",
					"sts.service.identifiers": "id-value",
				},
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			OpenTelemetrySpanBuilder(tc.span, tc.kind, tc.event, tc.resource, tc.sType, tc.layer, tc.domain, tc.urn, tc.id)
			assert.EqualValues(t, tc.expected, tc.span)
		})
	}
}
