package opentelemetry

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpanInterpreterEngine(t *testing.T) {
	for _, tc := range []struct {
		testCase string
		source   string
		span     pb.Span
		expected string
	}{
		{
			testCase: "Unknown instrumentation libraries should default back to the original source id",
			source:   api.OpenTelemetrySource,
			span: pb.Span{
				Meta: map[string]string{
					"instrumentation_library": "invalid",
				},
			},
			expected: api.OpenTelemetrySource,
		},
		{
			testCase: "Instrumentation library 'instrumentation-aws-lambda' should map to a specific interpreter id.",
			source:   api.OpenTelemetrySource,
			span: pb.Span{
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-aws-lambda",
				},
			},
			expected: "openTelemetryLambdaEntry",
		},
		{
			testCase: "Instrumentation library 'instrumentation-http' should map to a specific interpreter id.",
			source:   api.OpenTelemetrySource,
			span: pb.Span{
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-http",
				},
			},
			expected: "openTelemetryHttp",
		},
		{
			testCase: "Instrumentation library 'instrumentation-aws-sdk' with the service identifier sqs should map to a sqs interpreter id.",
			source:   api.OpenTelemetrySource,
			span: pb.Span{
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
					"aws.service.identifier":  "sqs",
				},
			},
			expected: "openTelemetrySQS",
		},
		{
			testCase: "Instrumentation library 'instrumentation-aws-sdk' with the service identifier sns should map to a sns interpreter id.",
			source:   api.OpenTelemetrySource,
			span: pb.Span{
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
					"aws.service.identifier":  "sns",
				},
			},
			expected: "openTelemetrySNS",
		},
		{
			testCase: "Instrumentation library 'instrumentation-aws-sdk' with the service identifier lambda should map to a lambda interpreter id.",
			source:   api.OpenTelemetrySource,
			span: pb.Span{
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
					"aws.service.identifier":  "lambda",
				},
			},
			expected: "openTelemetryLambda",
		},
		{
			testCase: "Instrumentation library 'instrumentation-aws-sdk' with the service identifier Lambda entry should map to a Lambda entry interpreter id.",
			source:   api.OpenTelemetrySource,
			span: pb.Span{
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
					"aws.service.identifier":  "lambdaentry",
				},
			},
			expected: "openTelemetryLambdaEntry",
		},
		{
			testCase: "Instrumentation library 'instrumentation-aws-sdk' with the service identifier s3 should map to a s3 interpreter id.",
			source:   api.OpenTelemetrySource,
			span: pb.Span{
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
					"aws.service.identifier":  "s3",
				},
			},
			expected: "openTelemetryS3",
		},
		{
			testCase: "Instrumentation library 'instrumentation-aws-sdk' with the service identifier step function should map to a step function interpreter id.",
			source:   api.OpenTelemetrySource,
			span: pb.Span{
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
					"aws.service.identifier":  "stepfunctions",
				},
			},
			expected: "openTelemetryStepFunctions",
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			trace := &tc.span
			source := tc.source
			actual := InterpretBasedOnInstrumentationLibrary(trace, source)
			assert.EqualValues(t, tc.expected, actual)
		})
	}
}
