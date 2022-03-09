package aws

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenTelemetryStepFunctionsSpanInterpreter(t *testing.T) {
	interpreter := MakeOpenTelemetryStepFunctionsInterpreter(config.DefaultInterpreterConfig())
	for _, tc := range []struct {
		testCase    string
		interpreter *OpenTelemetryStepFunctionsInterpreter
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
					"aws.operation":                 "execute",
					"aws.request.state.machine.arn": "arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
				},
			}},
			expected: []*pb.Span{{
				Name:     "State Machine: state-machine",
				Service:  "aws.step.function",
				Resource: "aws.step.function",
				Type:     "aws",
				Meta: map[string]string{
					"sts.origin":                    "open-telemetry",
					"source":                        "open-telemetry",
					"aws.operation":                 "execute",
					"aws.request.state.machine.arn": "arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
					"service":                       "aws.step.function",
					"span.kind":                     "consumer",
					"span.serviceName":              "State Machine: state-machine",
					"span.serviceType":              "open-telemetry",
					"span.serviceURN":               "urn:service:/arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
					"sts.service.identifiers":       "arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
				},
			}},
		},
		{
			testCase:    "Should interpret 4xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "aws.step.function",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"aws.operation":                 "execute",
					"aws.request.state.machine.arn": "arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
				},
			}},
			expected: []*pb.Span{{
				Name:     "State Machine: state-machine",
				Service:  "aws.step.function",
				Resource: "aws.step.function",
				Type:     "aws",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"sts.origin":                    "open-telemetry",
					"source":                        "open-telemetry",
					"span.errorClass":               "4xx",
					"aws.operation":                 "execute",
					"aws.request.state.machine.arn": "arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
					"service":                       "aws.step.function",
					"span.kind":                     "consumer",
					"span.serviceName":              "State Machine: state-machine",
					"span.serviceType":              "open-telemetry",
					"span.serviceURN":               "urn:service:/arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
					"sts.service.identifiers":       "arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
				},
			}},
		},
		{
			testCase:    "Should interpret 5xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "aws.step.function",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"aws.operation":                 "execute",
					"aws.request.state.machine.arn": "arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
				},
			}},
			expected: []*pb.Span{{
				Name:     "State Machine: state-machine",
				Service:  "aws.step.function",
				Resource: "aws.step.function",
				Type:     "aws",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"sts.origin":                    "open-telemetry",
					"source":                        "open-telemetry",
					"span.errorClass":               "5xx",
					"aws.operation":                 "execute",
					"aws.request.state.machine.arn": "arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
					"service":                       "aws.step.function",
					"span.kind":                     "consumer",
					"span.serviceName":              "State Machine: state-machine",
					"span.serviceType":              "open-telemetry",
					"span.serviceURN":               "urn:service:/arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
					"sts.service.identifiers":       "arn:aws:lambda:us-east-1:9876543210:sfn:state-machine",
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
