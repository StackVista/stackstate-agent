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

		arn, arnOk := modules.RetrieveValidSpanMeta(span, "LAMBDA", "faas.id")
		_, awsAccountIDOk := modules.RetrieveValidSpanMeta(span, "LAMBDA", "cloud.account.id")

		if arnOk && awsAccountIDOk {
			// Example Arn:
			// arn:aws:lambda:eu-west-1:965323806078:function:otel-example-nodejs-dev-success-and-failure
			arnParts := strings.Split(*arn, ":")

			if len(arnParts) >= 7 {
				functionName := arnParts[6]

				var urn = t.CreateServiceURN(strings.ToLower(*arn))

				// Name of component displayed below the icon
				span.Meta["span.serviceName"] = functionName

				// Name of the trace displayed on the trace graph line
				span.Name = fmt.Sprintf("%s: %s", "Lambda", functionName)

				// Displayed on the trace properties
				span.Resource = "aws.lambda"
				span.Type = "aws"

				// Mapping inside StackPack for capturing certain metrics
				span.Meta["span.serviceType"] = "open-telemetry"
				span.Meta["source"] = "open-telemetry"

				// Unknown
				span.Service = "aws.lambda"
				span.Meta["service"] = "aws.lambda"
				span.Meta["sts.origin"] = "open-telemetry"

				// General mapping
				span.Meta["span.kind"] = "server"
				span.Meta["span.serviceURN"] = urn
				span.Meta["sts.service.identifiers"] = *arn
			} else {
				_ = log.Errorf("[OTEL] [LAMBDA]: 'faas.id' invalid structure supplied '%s'", *arn)
				return nil
			}
		} else {
			_ = log.Errorf("[OTEL] [LAMBDA]: Unable to map the LAMBDA request")
			return nil
		}

		modules.InterpretHTTPError(span)
	}

	return spans
}
