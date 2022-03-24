package instrumentation_aws_sdk

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	instrumentationAwsSdkModules "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations/instrumentation-aws-sdk/modules"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

var InterpreterInstrumentationIdentifier = "@opentelemetry/instrumentation-aws-lambda"

// ComposeInstrumentationIdentifier an Open Telemetry mapping specific for aws services
// It is separate to the one above as these services might grow to a massive list of items
func ComposeInstrumentationIdentifier() string {
	lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, instrumentationAwsSdkModules.OpenTelemetryLambdaEntryServiceIdentifier)
	log.Debugf("[OTEL] [INSTRUMENTATION-AWS-LAMBDA] Mapping service: %s", lambdaInterpreterIdentifier)
	return lambdaInterpreterIdentifier
}
