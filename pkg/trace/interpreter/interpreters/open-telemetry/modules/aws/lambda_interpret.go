package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/modules"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"strings"
)

// OpenTelemetryLambdaEntryInterpreter default span interpreter for this data structure
type OpenTelemetryLambdaEntryInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetryLambdaEntryServiceIdentifier The base identifier for this interpreter, Is also used in identifying AWS services
const OpenTelemetryLambdaEntryServiceIdentifier = "LambdaEntry"

// OpenTelemetryLambdaEntryInterpreterSpan An identifier used to direct Open Telemetry interprets to this Interpreter
var OpenTelemetryLambdaEntryInterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryLambdaEntryServiceIdentifier)

// OpenTelemetryLambdaEntryAwsIdentifier An identifier used to map the AWS Service to the STS InterpreterServiceIdentifier
var OpenTelemetryLambdaEntryAwsIdentifier = strings.ToLower(OpenTelemetryLambdaEntryServiceIdentifier)

// MakeOpenTelemetryLambdaEntryInterpreter creates an instance of the OpenTelemetry Lambda Entry span interpreter
func MakeOpenTelemetryLambdaEntryInterpreter(config *config.Config) *OpenTelemetryLambdaEntryInterpreter {
	return &OpenTelemetryLambdaEntryInterpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetryLambdaEntryInterpreter
func (t *OpenTelemetryLambdaEntryInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	log.Debugf("[OTEL] [LAMBDA] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		// Extract meta information
		arn, arnOk := modules.RetrieveValidSpanMeta(span, "LAMBDA", "faas.id")
		_, awsAccountIDOk := modules.RetrieveValidSpanMeta(span, "LAMBDA", "cloud.account.id")

		// Only continue if there is a cloud.account.id
		// We do not use it but is important in mappings for the other aws-sdk libraries
		// If this is not found then we attempt to leave out the parent thus causing no trace link
		if arnOk && awsAccountIDOk {
			var urn = t.CreateServiceURN(strings.ToLower(*arn))

			// To reduce the amount of data mapped to the request we can extract the requirements from the ARN
			arnParts := strings.Split(*arn, ":")

			// Valid ARN will always have 7 parts
			if len(arnParts) >= 7 {
				functionName := arnParts[6]

				// Map information for this span
				modules.SpanBuilder(span, functionName, "Lambda", "lambda", "producer", urn, *arn)
			} else {
				_ = log.Errorf("[OTEL] [LAMBDA]: Unable to map LAMBDA because of a invalid arn, %s", *arn)
				return nil
			}
		} else {
			_ = log.Errorf("[OTEL] [LAMBDA]: Unable to map LAMBDA because of invalid data")
			return nil
		}

		modules.InterpretHTTPError(span)
	}

	return spans
}
