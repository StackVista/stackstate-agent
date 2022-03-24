package instrumentation_aws_sdk

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	instrumentationHttpModules "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations/instrumentation-http/modules"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

var InterpreterInstrumentationIdentifier = "@opentelemetry/instrumentation-http"

// ComposeInstrumentationIdentifier an Open Telemetry mapping specific for aws services
// It is separate to the one above as these services might grow to a massive list of items
func ComposeInstrumentationIdentifier() string {
	httpInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, instrumentationHttpModules.OpenTelemetryHTTPServiceIdentifier)
	log.Debugf("[OTEL] [INSTRUMENTATION-HTTP] Mapping service: %s", httpInterpreterIdentifier)
	return httpInterpreterIdentifier
}
