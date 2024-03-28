package modules

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	instrumentationbuilders "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentation-builders"
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

// MakeOpenTelemetryStackStateInterpreter creates an instance of the OpenTelemetry StackState span interpreter
func MakeOpenTelemetryStackStateInterpreter(config *config.Config) *OpenTelemetryStackStateInterpreter {
	return &OpenTelemetryStackStateInterpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetryStackStateInterpreter
func (t *OpenTelemetryStackStateInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	log.Debugf("[OTEL] [CUSTOM-METRIC] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		tracePerspectiveName, tracePerspectiveNameOk := instrumentationbuilders.GetSpanMeta("CUSTOM-METRIC", span, "trace.perspective.name")
		serviceName, serviceNameOk := instrumentationbuilders.GetSpanMeta("CUSTOM-METRIC", span, "service.name")
		serviceType, serviceTypeOk := instrumentationbuilders.GetSpanMeta("CUSTOM-METRIC", span, "service.type")
		serviceIdentifier, serviceIdentifierOk := instrumentationbuilders.GetSpanMeta("CUSTOM-METRIC", span, "service.identifier")
		resourceName, resourceNameOk := instrumentationbuilders.GetSpanMeta("CUSTOM-METRIC", span, "resource.name")

		if tracePerspectiveNameOk && serviceNameOk && serviceTypeOk && serviceIdentifierOk && resourceNameOk {
			var kind = "consumer"
			var urn = t.CreateServiceURN(*serviceIdentifier)

			instrumentationbuilders.StackStateSpanBuilder(span, *tracePerspectiveName, *serviceType, *serviceName, *serviceIdentifier, *resourceName, kind, urn)
		} else {
			_ = log.Errorf("[OTEL] [CUSTOM-METRIC]: Unable to map the custom metric request")
			return nil
		}

		instrumentationbuilders.InterpretSpanHTTPError(span)
	}

	return spans
}
