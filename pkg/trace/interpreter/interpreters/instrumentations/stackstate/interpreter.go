package aws_sdk

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	stackStateInstrumentationModules "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentations/stackstate/modules"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// InstrumentationIdentifier Identifier for this instrumentation library
// This identifier is based on the one used within the library using within the wild
// External link for this does not exist yet and might only be internal mappings
var InstrumentationIdentifier = "@opentelemetry/instrumentation-stackstate"

// InterpretBuilderForStackStateInstrumentation Mapping for modules used within the Instrumentation library
// Modules are basically sub parts that have certain functionality. Based on that functionality different context
// values needs to be mapped to be part of the span response
func InterpretBuilderForStackStateInstrumentation() string {
	// StackState Instrumentation has no sub routing as StackState is top level

	stackStateInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, stackStateInstrumentationModules.OpenTelemetryStackStateServiceIdentifier)
	log.Debugf("[OTEL] [STACKSTATE] Mapping service: %s", stackStateInterpreterIdentifier)
	return stackStateInterpreterIdentifier
}
