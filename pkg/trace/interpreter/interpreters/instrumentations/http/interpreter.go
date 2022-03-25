package httpInstrumentation

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	httpInstrumentationModules "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentations/http/modules"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// InstrumentationIdentifier Identifier for this instrumentation library
// This identifier is based on the one used within the library using within the wild
// https://www.npmjs.com/package/@opentelemetry/instrumentation-http
var InstrumentationIdentifier = "@opentelemetry/instrumentation-http"

// InterpretBuilderForHttpInstrumentation Mapping for modules used within the Instrumentation library
// Modules are basically sub parts that have certain functionality. Based on that functionality different context
// values needs to be mapped to be part of the span response
func InterpretBuilderForHttpInstrumentation() string {
	// HTTP Instrumentation has no sub routing as HTTP is top level

	httpInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, httpInstrumentationModules.OpenTelemetryHTTPServiceIdentifier)
	log.Debugf("[OTEL] [INSTRUMENTATION-HTTP] Mapping service: %s", httpInterpreterIdentifier)
	return httpInterpreterIdentifier
}
