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
					"cloud.account.id": "9876543210",
					"faas.id":          "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
				},
			}},

			expected: []*pb.Span{{
				Service:  "aws.lambda",
				Name:     "Lambda: function-name",
				Resource: "aws.lambda",
				Type:     "aws",
				Meta: map[string]string{
					"source":                  "open-telemetry",
					"sts.origin":              "open-telemetry",
					"span.kind":               "producer",
					"cloud.account.id":        "9876543210",
					"faas.id":                 "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
					"service":                 "aws.lambda",
					"span.serviceName":        "function-name",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/arn:aws:lambda:eu-west-1:965323806078:function:function-name",
					"sts.service.identifiers": "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
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
					"cloud.account.id": "9876543210",
					"faas.id":          "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
				},
			}},
			expected: []*pb.Span{{
				Service: "aws.lambda",
				Name:    "Lambda: function-name",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Resource: "aws.lambda",
				Type:     "aws",
				Meta: map[string]string{
					"source":                  "open-telemetry",
					"span.errorClass":         "4xx",
					"sts.origin":              "open-telemetry",
					"span.kind":               "producer",
					"cloud.account.id":        "9876543210",
					"faas.id":                 "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
					"service":                 "aws.lambda",
					"span.serviceName":        "function-name",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/arn:aws:lambda:eu-west-1:965323806078:function:function-name",
					"sts.service.identifiers": "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
				},
			}},
		},
		{
			testCase:    "Should interpret 5xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "aws.lambda",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"cloud.account.id": "9876543210",
					"faas.id":          "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
				},
			}},
			expected: []*pb.Span{{
				Service: "aws.lambda",
				Name:    "Lambda: function-name",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Resource: "aws.lambda",
				Type:     "aws",
				Meta: map[string]string{
					"source":                  "open-telemetry",
					"span.errorClass":         "5xx",
					"sts.origin":              "open-telemetry",
					"span.kind":               "producer",
					"cloud.account.id":        "9876543210",
					"faas.id":                 "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
					"service":                 "aws.lambda",
					"span.serviceName":        "function-name",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/arn:aws:lambda:eu-west-1:965323806078:function:function-name",
					"sts.service.identifiers": "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
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
