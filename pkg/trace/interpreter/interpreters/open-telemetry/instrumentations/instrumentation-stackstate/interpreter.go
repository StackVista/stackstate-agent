package instrumentation_aws_sdk

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	instrumentationStackStateModules "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations/instrumentation-stackstate/modules"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

var InterpreterInstrumentationIdentifier = "@opentelemetry/instrumentation-stackstate"

// ComposeInstrumentationIdentifier an Open Telemetry mapping specific for aws services
// It is separate to the one above as these services might grow to a massive list of items
func ComposeInstrumentationIdentifier() string {
	stackStateInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, instrumentationStackStateModules.OpenTelemetryStackStateServiceIdentifier)
	log.Debugf("[OTEL] [INSTRUMENTATION-STACKSTATE] Mapping service: %s", stackStateInterpreterIdentifier)
	return stackStateInterpreterIdentifier
}
