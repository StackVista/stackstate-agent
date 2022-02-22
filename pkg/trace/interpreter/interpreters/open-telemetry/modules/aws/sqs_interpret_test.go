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
				Name:     "SQS Queue: target-queue-destination-9876543210-eu-west-1",
				Service:  "aws.sqs.queue",
				Resource: "aws.sqs.queue",
				Type:     "aws",
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"aws.operation":           "publishMessage",
					"aws.region":              "eu-west-1",
					"messaging.destination":   "target-queue-destination",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"service":                 "aws.sqs.queue",
					"span.kind":               "consumer",
					"span.serviceName":        "SQS Queue: target-queue-destination-9876543210-eu-west-1",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"sts.service.identifiers": "https://eu-west-1.queue.amazonaws.com/9876543210/target-queue-destination",
				},
			}},
		},
		{
			testCase:    "Should interpret 4xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Error: 1,
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
				Name:     "SQS Queue: target-queue-destination-9876543210-eu-west-1",
				Service:  "aws.sqs.queue",
				Resource: "aws.sqs.queue",
				Type:     "aws",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"span.errorClass":         "4xx",
					"aws.operation":           "publishMessage",
					"aws.region":              "eu-west-1",
					"messaging.destination":   "target-queue-destination",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"service":                 "aws.sqs.queue",
					"span.kind":               "consumer",
					"span.serviceName":        "SQS Queue: target-queue-destination-9876543210-eu-west-1",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"sts.service.identifiers": "https://eu-west-1.queue.amazonaws.com/9876543210/target-queue-destination",
				},
			}},
		},
		{
			testCase:    "Should interpret 5xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Error: 1,
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
				Name:     "SQS Queue: target-queue-destination-9876543210-eu-west-1",
				Service:  "aws.sqs.queue",
				Resource: "aws.sqs.queue",
				Type:     "aws",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"span.errorClass":         "5xx",
					"aws.operation":           "publishMessage",
					"aws.region":              "eu-west-1",
					"messaging.destination":   "target-queue-destination",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"service":                 "aws.sqs.queue",
					"span.kind":               "consumer",
					"span.serviceName":        "SQS Queue: target-queue-destination-9876543210-eu-west-1",
					"span.serviceType":        "open-telemetry",
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
