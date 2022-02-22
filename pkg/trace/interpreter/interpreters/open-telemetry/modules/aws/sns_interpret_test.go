package aws

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenTelemetrySNSSpanInterpreter(t *testing.T) {
	interpreter := MakeOpenTelemetrySNSInterpreter(config.DefaultInterpreterConfig())
	for _, tc := range []struct {
		testCase    string
		interpreter *OpenTelemetrySNSInterpreter
		trace       []*pb.Span
		expected    []*pb.Span
	}{
		{
			testCase:    "Span should not be filled in if the Open Telemetry data is invalid or missing",
			interpreter: interpreter,
			trace:       []*pb.Span{},
			expected:    []*pb.Span{},
		},
		{
			testCase:    "Open Telemetry data should be mapped if all the correct meta data has been passed",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Meta: map[string]string{
					"aws.operation":         "postMessage",
					"aws.request.topic.arn": "arn:aws:sns:eu-west-1:965323806078:topic-name",
				},
			}},
			expected: []*pb.Span{{
				Name:     "SNS: topic-name",
				Service:  "aws.sns",
				Resource: "aws.sns",
				Type:     "aws",
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"aws.operation":           "postMessage",
					"aws.request.topic.arn":   "arn:aws:sns:eu-west-1:965323806078:topic-name",
					"service":                 "aws.sns",
					"span.kind":               "consumer",
					"span.serviceName":        "topic-name",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/arn:aws:sns:eu-west-1:965323806078:topic-name",
					"sts.service.identifiers": "arn:aws:sns:eu-west-1:965323806078:topic-name",
				},
			}},
		},
		{
			testCase:    "Should interpret 4xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "service-name",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"aws.operation":         "postMessage",
					"aws.request.topic.arn": "arn:aws:sns:eu-west-1:965323806078:topic-name",
				},
			}},
			expected: []*pb.Span{{
				Name:     "SNS: topic-name",
				Service:  "aws.sns",
				Resource: "aws.sns",
				Type:     "aws",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"span.errorClass":         "4xx",
					"aws.operation":           "postMessage",
					"aws.request.topic.arn":   "arn:aws:sns:eu-west-1:965323806078:topic-name",
					"service":                 "aws.sns",
					"span.kind":               "consumer",
					"span.serviceName":        "topic-name",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/arn:aws:sns:eu-west-1:965323806078:topic-name",
					"sts.service.identifiers": "arn:aws:sns:eu-west-1:965323806078:topic-name",
				},
			}},
		},
		{
			testCase:    "Should interpret 5xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "aws.sns",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"aws.operation":         "postMessage",
					"aws.request.topic.arn": "arn:aws:sns:eu-west-1:965323806078:topic-name",
				},
			}},
			expected: []*pb.Span{{
				Name:     "SNS: topic-name",
				Service:  "aws.sns",
				Resource: "aws.sns",
				Type:     "aws",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"span.errorClass":         "5xx",
					"aws.operation":           "postMessage",
					"aws.request.topic.arn":   "arn:aws:sns:eu-west-1:965323806078:topic-name",
					"service":                 "aws.sns",
					"span.kind":               "consumer",
					"span.serviceName":        "topic-name",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/arn:aws:sns:eu-west-1:965323806078:topic-name",
					"sts.service.identifiers": "arn:aws:sns:eu-west-1:965323806078:topic-name",
				},
			}},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			actual := tc.interpreter.Interpret(tc.trace)
			assert.EqualValues(t, tc.expected, actual)
		})
	}
}
