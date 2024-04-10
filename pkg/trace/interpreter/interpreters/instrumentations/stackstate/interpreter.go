package stackstateinstrumentation

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/trace/api"
	stackstateinstrumentationModules "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentations/stackstate/modules"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// InstrumentationIdentifier Identifier for this instrumentation library
// This identifier is based on the one used within the library using within the wild
// External link for this does not exist yet and might only be internal mappings
var InstrumentationIdentifier = "@opentelemetry/instrumentation-stackstate"

// InterpretBuilderForStackStateInstrumentation Mapping for modules used within the Instrumentation library
// Modules are basically sub parts that have certain functionality. Based on that functionality different context
// values needs to be mapped to be part of the span response
func InterpretBuilderForStackStateInstrumentation() string {
	stackStateInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, stackstateinstrumentationModules.OpenTelemetryStackStateServiceIdentifier)
	log.Debugf("[OTEL] [STACKSTATE] Mapping service: %s", stackStateInterpreterIdentifier)
	return stackStateInterpreterIdentifier
}
