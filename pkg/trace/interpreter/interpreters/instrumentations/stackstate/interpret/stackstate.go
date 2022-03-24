package interpret

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentations"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// OpenTelemetryStackStateInterpreter default span interpreter for this data structure
type OpenTelemetryStackStateInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetryStackStateServiceIdentifier The base identifier for this interpreter, Is also used in identifying AWS services
const OpenTelemetryStackStateServiceIdentifier = "StackState"

// OpenTelemetryStackStateInterpreterSpan An identifier used to direct Open Telemetry interprets to this Interpreter
var OpenTelemetryStackStateInterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryStackStateServiceIdentifier)

// MakeOpenTelemetryStackStateInterpreter creates an instance of the OpenTelemetry HTTP span interpreter
func MakeOpenTelemetryStackStateInterpreter(config *config.Config) *OpenTelemetryStackStateInterpreter {
	return &OpenTelemetryStackStateInterpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetryHTTPInterpreter
func (t *OpenTelemetryStackStateInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	log.Debugf("[OTEL] [STACKSTATE] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		log.Debugf("[OTEL] [INSTRUMENTATION-STACKSTATE] Entry into instrumentation")

		instrumentations.InterpretHTTPError(span)
	}

	return spans
}
