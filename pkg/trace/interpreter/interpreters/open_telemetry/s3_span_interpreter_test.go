package interpreters

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenTelemetryS3Interpreter(t *testing.T) {
	openTelemetryInterpreter := MakeOpenTelemetryS3Interpreter(config.DefaultInterpreterConfig())

	for _, tc := range []struct {
		testCase    string
		interpreter *OpenTelemetryS3Interpreter
		trace       []*pb.Span
		expected    []*pb.Span
	}{
		{
			testCase:    "Should set span.serviceType to 'openTelemetryS3' when no span.kind metadata exists",
			interpreter: openTelemetryInterpreter,
			trace:       []*pb.Span{{Service: "service-name"}},
			expected: []*pb.Span{{
				Service: "service-name",
				Meta: map[string]string{
					"span.serviceType": "openTelemetryS3",
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
