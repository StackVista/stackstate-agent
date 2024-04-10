package httpinstrumentation

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/trace/api"
	httpinstrumentationModules "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentations/http/modules"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// InstrumentationIdentifier Identifier for this instrumentation library
// This identifier is based on the one used within the library using within the wild
// https://www.npmjs.com/package/@opentelemetry/instrumentation-http
var InstrumentationIdentifier = "@opentelemetry/instrumentation-http"

// InterpretBuilderForHTTPInstrumentation Mapping for modules used within the Instrumentation library
// Modules are basically sub parts that have certain functionality. Based on that functionality different context
// values needs to be mapped to be part of the span response
func InterpretBuilderForHTTPInstrumentation() string {
	httpInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, httpinstrumentationModules.OpenTelemetryHTTPServiceIdentifier)
	log.Debugf("[OTEL] [INSTRUMENTATION-HTTP] Mapping service: %s", httpInterpreterIdentifier)
	return httpInterpreterIdentifier
}
