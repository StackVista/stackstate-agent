package modules

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpanBuilderInterpreterEngine(t *testing.T) {
	for _, tc := range []struct {
		testCase    string
		span        *pb.Span
		expected    *pb.Span
		namePrefix  string
		service     string
		serviceName string
		kind        string
		urn         string
		arn         string
	}{
		{
			testCase:    "A empty span ran through the AWS Span Builder should supply meta data",
			namePrefix:  "Name Prefix",
			service:     "service-item",
			serviceName: "Service",
			kind:        "consumer",
			urn:         "urn:service:/hostname",
			arn:         "arn:/urn:service:/hostname",
			span: &pb.Span{
				Meta: map[string]string{},
			},
			expected: &pb.Span{
				Name:     "Name Prefix: Service",
				Service:  "aws.service-item",
				Resource: "aws.service-item",
				Type:     "aws",
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"service":                 "aws.service-item",
					"span.kind":               "consumer",
					"span.serviceName":        "Service",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/hostname",
					"sts.service.identifiers": "arn:/urn:service:/hostname",
				},
			},
		},
		{
			testCase:    "Span containing meta data should contain the new and original meta data through the AWS Span Builder",
			namePrefix:  "Name Prefix",
			service:     "service-item",
			serviceName: "Service",
			kind:        "consumer",
			urn:         "urn:service:/hostname",
			arn:         "arn:/urn:service:/hostname",
			span: &pb.Span{
				Meta: map[string]string{
					"extra-item-a": "value-a",
					"extra-item-b": "value-b",
				},
			},
			expected: &pb.Span{
				Name:     "Name Prefix: Service",
				Service:  "aws.service-item",
				Resource: "aws.service-item",
				Type:     "aws",
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"extra-item-a":            "value-a",
					"extra-item-b":            "value-b",
					"service":                 "aws.service-item",
					"span.kind":               "consumer",
					"span.serviceName":        "Service",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/hostname",
					"sts.service.identifiers": "arn:/urn:service:/hostname",
				},
			},
		},
		{
			testCase:    "AWS Span Builder should overwrite existing data if it was predefined",
			namePrefix:  "Name Prefix",
			service:     "service-item",
			serviceName: "Service",
			kind:        "consumer",
			urn:         "urn:service:/hostname",
			arn:         "arn:/urn:service:/hostname",
			span: &pb.Span{
				Meta: map[string]string{
					"service":                 "value-should-be-ignored",
					"span.kind":               "value-should-be-ignored",
					"span.serviceName":        "value-should-be-ignored",
					"span.serviceType":        "value-should-be-ignored",
					"span.serviceURN":         "value-should-be-ignored",
					"sts.service.identifiers": "value-should-be-ignored",
				},
			},
			expected: &pb.Span{
				Name:     "Name Prefix: Service",
				Service:  "aws.service-item",
				Resource: "aws.service-item",
				Type:     "aws",
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"service":                 "aws.service-item",
					"span.kind":               "consumer",
					"span.serviceName":        "Service",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/hostname",
					"sts.service.identifiers": "arn:/urn:service:/hostname",
				},
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			SpanBuilder(tc.span, tc.serviceName, tc.namePrefix, tc.service, tc.kind, tc.urn, tc.arn)
			assert.EqualValues(t, tc.expected, tc.span)
		})
	}
}

func TestRetrieveValidSpanMeta(t *testing.T) {
	validData, validOk := RetrieveValidSpanMeta(&pb.Span{
		Meta: map[string]string{
			"extra-item-a": "value-a",
			"extra-item-b": "value-b",
		},
	}, "Valid Data", "extra-item-a")

	assert.EqualValues(t, "value-a", *validData)
	assert.EqualValues(t, true, validOk)

	invalidData, invalidOk := RetrieveValidSpanMeta(&pb.Span{
		Meta: map[string]string{
			"extra-item-a": "value-a",
			"extra-item-b": "value-b",
		},
	}, "Invalid Data", "extra-item-c")

	assert.Nil(t, invalidData)
	assert.EqualValues(t, false, invalidOk)
}

func TestInterpretHTTPError(t *testing.T) {
	span := pb.Span{
		Meta: map[string]string{
			"extra-item-a": "value-a",
			"extra-item-b": "value-b",
		},
	}
	InterpretHTTPError(&span)
	assert.EqualValues(t, pb.Span{
		Meta: map[string]string{
			"extra-item-a": "value-a",
			"extra-item-b": "value-b",
		},
	}, span)

	spanWithError := pb.Span{
		Error: 404,
		Metrics: map[string]float64{
			"http.status_code": 404,
		},
		Meta: map[string]string{
			"extra-item-a": "value-a",
			"extra-item-b": "value-b",
		},
	}
	InterpretHTTPError(&spanWithError)
	assert.EqualValues(t, pb.Span{
		Error: 404,
		Meta: map[string]string{
			"extra-item-a":    "value-a",
			"extra-item-b":    "value-b",
			"span.errorClass": "4xx",
		},
		Metrics: map[string]float64{
			"http.status_code": 404,
		},
	}, spanWithError)
}
