package aws

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenTelemetryS3SpanInterpreter(t *testing.T) {
	interpreter := MakeOpenTelemetryS3Interpreter(config.DefaultInterpreterConfig())
	for _, tc := range []struct {
		testCase    string
		interpreter *OpenTelemetryS3Interpreter
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
					"aws.request.bucket": "bucket-name",
					"aws.operation":      "putObject",
				},
			}},
			expected: []*pb.Span{{
				Service:  "open.telemetry.s3",
				Resource: "aws.s3",
				Type:     "S3 Bucket",
				Meta: map[string]string{
					"aws.operation":           "putObject",
					"aws.request.bucket":      "bucket-name",
					"domain":                  "test-eu-west-1",
					"layer":                   "Storage",
					"service":                 "open.telemetry.s3",
					"span.kind":               "consumer",
					"span.serviceName":        "open.telemetry.s3.putObject",
					"span.serviceType":        "S3 Bucket",
					"span.serviceURN":         "urn:service:/arn:aws:s3:::bucket-name",
					"sts.service.identifiers": "arn:aws:s3:::bucket-name",
				},
			}},
		},
		{
			testCase:    "Should interpret 4xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "open.telemetry.s3",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"aws.request.bucket": "bucket-name",
					"aws.operation":      "putObject",
				},
			}},
			expected: []*pb.Span{{
				Service:  "open.telemetry.s3",
				Resource: "aws.s3",
				Type:     "S3 Bucket",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"span.errorClass":         "4xx",
					"aws.operation":           "putObject",
					"aws.request.bucket":      "bucket-name",
					"domain":                  "test-eu-west-1",
					"layer":                   "Storage",
					"service":                 "open.telemetry.s3",
					"span.kind":               "consumer",
					"span.serviceName":        "open.telemetry.s3.putObject",
					"span.serviceType":        "S3 Bucket",
					"span.serviceURN":         "urn:service:/arn:aws:s3:::bucket-name",
					"sts.service.identifiers": "arn:aws:s3:::bucket-name",
				},
			}},
		},
		{
			testCase:    "Should interpret 5xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "open.telemetry.S3",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"aws.request.bucket": "bucket-name",
					"aws.operation":      "putObject",
				},
			}},
			expected: []*pb.Span{{
				Service:  "open.telemetry.s3",
				Resource: "aws.s3",
				Type:     "S3 Bucket",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"span.errorClass":         "5xx",
					"aws.operation":           "putObject",
					"aws.request.bucket":      "bucket-name",
					"domain":                  "test-eu-west-1",
					"layer":                   "Storage",
					"service":                 "open.telemetry.s3",
					"span.kind":               "consumer",
					"span.serviceName":        "open.telemetry.s3.putObject",
					"span.serviceType":        "S3 Bucket",
					"span.serviceURN":         "urn:service:/arn:aws:s3:::bucket-name",
					"sts.service.identifiers": "arn:aws:s3:::bucket-name",
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
