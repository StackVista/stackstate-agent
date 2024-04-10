package modules

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/trace/api"
	config "github.com/DataDog/datadog-agent/pkg/trace/interpreter/config"
	interpreter "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters"
	instrumentationbuilders "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentation-builders"
	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"strings"
)

// OpenTelemetryStepFunctionsInterpreter default span interpreter for this data structure
type OpenTelemetryStepFunctionsInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetrySFNServiceIdentifier The base identifier for this interpreter, Is also used in identifying AWS services
const OpenTelemetrySFNServiceIdentifier = "StepFunctions"

// OpenTelemetrySFNInterpreterSpan An identifier used to direct Open Telemetry interprets to this Interpreter
var OpenTelemetrySFNInterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetrySFNServiceIdentifier)

// OpenTelemetrySFNAwsIdentifier An identifier used to map the AWS Service to the STS InterpreterServiceIdentifier
var OpenTelemetrySFNAwsIdentifier = strings.ToLower(OpenTelemetrySFNServiceIdentifier)

// MakeOpenTelemetryStepFunctionsInterpreter creates an instance of the OpenTelemetryStepFunctions span interpreter
func MakeOpenTelemetryStepFunctionsInterpreter(config *config.Config) *OpenTelemetryStepFunctionsInterpreter {
	return &OpenTelemetryStepFunctionsInterpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetryStepFunctionsInterpreter
func (t *OpenTelemetryStepFunctionsInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	log.Debugf("[OTEL] [SFN] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		stateMachineArn, stateMachineArnOk := instrumentationbuilders.GetSpanMeta("SFN", span, "aws.request.state.machine.arn")

		if stateMachineArnOk {
			var arn = strings.ToLower(*stateMachineArn)
			var urn = t.CreateServiceURN(arn)

			arnParts := strings.Split(arn, ":")

			if len(arnParts) >= 7 {
				stateFunctionName := arnParts[6]
				instrumentationbuilders.AwsSpanBuilder(span, stateFunctionName, "State Machine", "step.function", "consumer", urn, arn)
			} else {
				_ = log.Errorf("[OTEL] [SFN]: 'arn' invalid structure supplied '%s'", arn)
				return nil
			}
		} else {
			_ = log.Errorf("[OTEL] [SFN]: Unable to map the Step Functions request")
			return nil
		}

		instrumentationbuilders.InterpretSpanHTTPError(span)
	}

	return spans
}
