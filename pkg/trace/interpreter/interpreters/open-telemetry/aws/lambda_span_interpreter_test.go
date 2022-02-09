package aws

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenTelemetryLambdaSpanInterpreter(t *testing.T) {
	interpreter := MakeOpenTelemetryLambdaInterpreter(config.DefaultInterpreterConfig())
	for _, tc := range []struct {
		testCase    string
		interpreter *OpenTelemetryLambdaInterpreter
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
					"aws.request.function.name": "example-function-name",
					"aws.account.id":            "9876543210",
					"aws.region":                "us-east-1",
					"aws.operation":             "invoke",
				},
			}},
			expected: []*pb.Span{{
				Service:  "open.telemetry.lambda",
				Resource: "aws.lambda",
				Type:     "Lambda Function",
				Meta: map[string]string{
					"aws.account.id":            "9876543210",
					"aws.operation":             "invoke",
					"aws.region":                "us-east-1",
					"aws.request.function.name": "example-function-name",
					"domain":                    "test-eu-west-1",
					"layer":                     "Serverless",
					"service":                   "open.telemetry.lambda",
					"span.kind":                 "consumer",
					"span.serviceName":          "open.telemetry.lambda.invoke",
					"span.serviceType":          "Lambda Function",
					"span.serviceURN":           "urn:service:/arn:aws:lambda:us-east-1:9876543210:function:example-function-name",
					"sts.service.identifiers":   "arn:aws:lambda:us-east-1:9876543210:function:example-function-name",
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
					"aws.request.function.name": "example-function-name",
					"aws.account.id":            "9876543210",
					"aws.region":                "us-east-1",
					"aws.operation":             "invoke",
				},
			}},
			expected: []*pb.Span{{
				Service: "open.telemetry.lambda",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Resource: "aws.lambda",
				Type:     "Lambda Function",
				Meta: map[string]string{
					"span.errorClass":           "4xx",
					"aws.account.id":            "9876543210",
					"aws.operation":             "invoke",
					"aws.region":                "us-east-1",
					"aws.request.function.name": "example-function-name",
					"domain":                    "test-eu-west-1",
					"layer":                     "Serverless",
					"service":                   "open.telemetry.lambda",
					"span.kind":                 "consumer",
					"span.serviceName":          "open.telemetry.lambda.invoke",
					"span.serviceType":          "Lambda Function",
					"span.serviceURN":           "urn:service:/arn:aws:lambda:us-east-1:9876543210:function:example-function-name",
					"sts.service.identifiers":   "arn:aws:lambda:us-east-1:9876543210:function:example-function-name",
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
					"aws.request.function.name": "example-function-name",
					"aws.account.id":            "9876543210",
					"aws.region":                "us-east-1",
					"aws.operation":             "invoke",
				},
			}},
			expected: []*pb.Span{{
				Service: "open.telemetry.lambda",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Resource: "aws.lambda",
				Type:     "Lambda Function",
				Meta: map[string]string{
					"span.errorClass":           "5xx",
					"aws.account.id":            "9876543210",
					"aws.operation":             "invoke",
					"aws.region":                "us-east-1",
					"aws.request.function.name": "example-function-name",
					"domain":                    "test-eu-west-1",
					"layer":                     "Serverless",
					"service":                   "open.telemetry.lambda",
					"span.kind":                 "consumer",
					"span.serviceName":          "open.telemetry.lambda.invoke",
					"span.serviceType":          "Lambda Function",
					"span.serviceURN":           "urn:service:/arn:aws:lambda:us-east-1:9876543210:function:example-function-name",
					"sts.service.identifiers":   "arn:aws:lambda:us-east-1:9876543210:function:example-function-name",
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
