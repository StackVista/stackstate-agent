package aws

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenTelemetryLambdaEntrySpanInterpreter(t *testing.T) {
	interpreter := MakeOpenTelemetryLambdaEntryInterpreter(config.DefaultInterpreterConfig())
	for _, tc := range []struct {
		testCase    string
		interpreter *OpenTelemetryLambdaEntryInterpreter
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
					"faas.id": "0000-0000-0000-0000",
				},
			}},
			expected: []*pb.Span{{
				Service:  "open.telemetry.lambda",
				Resource: "aws.lambda",
				// Type:     "open-telemetry",
				Type: "Lambda Function",
				Meta: map[string]string{
					"domain":                  "test-eu-west-1",
					"layer":                   "Serverless",
					"span.kind":               "producer",
					"faas.id":                 "0000-0000-0000-0000",
					"service":                 "open.telemetry.lambda",
					"span.serviceName":        "open.telemetry.lambda.execute",
					"span.serviceType":        "Lambda Function",
					"span.serviceURN":         "urn:service:/0000-0000-0000-0000",
					"sts.service.identifiers": "0000-0000-0000-0000",
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
					"faas.id": "0000-0000-0000-0000",
				},
			}},
			expected: []*pb.Span{{
				Service: "open.telemetry.lambda",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Resource: "aws.lambda",
				// Type:     "open-telemetry",
				Type: "Lambda Function",
				Meta: map[string]string{
					"domain":                  "test-eu-west-1",
					"layer":                   "Serverless",
					"span.errorClass":         "4xx",
					"span.kind":               "producer",
					"faas.id":                 "0000-0000-0000-0000",
					"service":                 "open.telemetry.lambda",
					"span.serviceName":        "open.telemetry.lambda.execute",
					"span.serviceType":        "Lambda Function",
					"span.serviceURN":         "urn:service:/0000-0000-0000-0000",
					"sts.service.identifiers": "0000-0000-0000-0000",
				},
			}},
		},
		{
			testCase:    "Should interpret 5xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "open.telemetry.lambda",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"faas.id": "0000-0000-0000-0000",
				},
			}},
			expected: []*pb.Span{{
				Service: "open.telemetry.lambda",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Resource: "aws.lambda",
				// Type:     "open-telemetry",
				Type: "Lambda Function",
				Meta: map[string]string{
					"domain":                  "test-eu-west-1",
					"layer":                   "Serverless",
					"span.errorClass":         "5xx",
					"span.kind":               "producer",
					"faas.id":                 "0000-0000-0000-0000",
					"service":                 "open.telemetry.lambda",
					"span.serviceName":        "open.telemetry.lambda.execute",
					"span.serviceType":        "Lambda Function",
					"span.serviceURN":         "urn:service:/0000-0000-0000-0000",
					"sts.service.identifiers": "0000-0000-0000-0000",
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
