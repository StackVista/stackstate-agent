package aws

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenTelemetrySQSSpanInterpreter(t *testing.T) {
	interpreter := MakeOpenTelemetrySQSInterpreter(config.DefaultInterpreterConfig())
	for _, tc := range []struct {
		testCase    string
		interpreter *OpenTelemetrySQSInterpreter
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
					"aws.region":            "eu-west-1",
					"aws.operation":         "publishMessage",
					"messaging.url":         "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"messaging.destination": "target-queue-destination",
				},
			}},
			expected: []*pb.Span{{
				Service:  "open.telemetry.sqs",
				Resource: "aws.sqs",
				Type:     "SQS Queue",
				Meta: map[string]string{
					"domain":                  "test-eu-west-1",
					"layer":                   "Messaging",
					"aws.operation":           "publishMessage",
					"aws.region":              "eu-west-1",
					"messaging.destination":   "target-queue-destination",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"service":                 "open.telemetry.sqs",
					"span.kind":               "consumer",
					"span.serviceName":        "open.telemetry.sqs.publishMessage",
					"span.serviceType":        "SQS Queue",
					"span.serviceURN":         "urn:service:/https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"sts.service.identifiers": "https://eu-west-1.queue.amazonaws.com/9876543210/target-queue-destination",
				},
			}},
		},
		{
			testCase:    "Should interpret 4xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "open.telemetry.sqs",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"aws.region":            "eu-west-1",
					"aws.operation":         "publishMessage",
					"messaging.url":         "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"messaging.destination": "target-queue-destination",
				},
			}},
			expected: []*pb.Span{{
				Service:  "open.telemetry.sqs",
				Resource: "aws.sqs",
				Type:     "SQS Queue",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"domain":                  "test-eu-west-1",
					"layer":                   "Messaging",
					"span.errorClass":         "4xx",
					"aws.operation":           "publishMessage",
					"aws.region":              "eu-west-1",
					"messaging.destination":   "target-queue-destination",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"service":                 "open.telemetry.sqs",
					"span.kind":               "consumer",
					"span.serviceName":        "open.telemetry.sqs.publishMessage",
					"span.serviceType":        "SQS Queue",
					"span.serviceURN":         "urn:service:/https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"sts.service.identifiers": "https://eu-west-1.queue.amazonaws.com/9876543210/target-queue-destination",
				},
			}},
		},
		{
			testCase:    "Should interpret 5xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "open.telemetry.sns",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"aws.region":            "eu-west-1",
					"aws.operation":         "publishMessage",
					"messaging.url":         "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"messaging.destination": "target-queue-destination",
				},
			}},
			expected: []*pb.Span{{
				Service:  "open.telemetry.sqs",
				Resource: "aws.sqs",
				Type:     "SQS Queue",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"domain":                  "test-eu-west-1",
					"layer":                   "Messaging",
					"span.errorClass":         "5xx",
					"aws.operation":           "publishMessage",
					"aws.region":              "eu-west-1",
					"messaging.destination":   "target-queue-destination",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"service":                 "open.telemetry.sqs",
					"span.kind":               "consumer",
					"span.serviceName":        "open.telemetry.sqs.publishMessage",
					"span.serviceType":        "SQS Queue",
					"span.serviceURN":         "urn:service:/https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"sts.service.identifiers": "https://eu-west-1.queue.amazonaws.com/9876543210/target-queue-destination",
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
