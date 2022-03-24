package modules

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenTelemetryStackStateSpanInterpreter(t *testing.T) {
	// interpreter := MakeOpenTelemetryStackStateInterpreter(config.DefaultInterpreterConfig())

	for _, tc := range []struct {
		testCase    string
		interpreter *OpenTelemetryStackStateInterpreter
		trace       []*pb.Span
		expected    []*pb.Span
	}{} {
		t.Run(tc.testCase, func(t *testing.T) {
			actual := tc.interpreter.Interpret(tc.trace)
			assert.EqualValues(t, tc.expected, actual)
		})
	}
}
