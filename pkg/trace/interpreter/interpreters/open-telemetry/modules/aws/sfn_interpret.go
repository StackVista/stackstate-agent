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

		// awsService, awsServiceOk := span.Meta["aws.service.api"]
		// awsOperation, awsOperationOk := modules.RetrieveValidSpanMeta(span, "SFN", "aws.operation")
		stateMachineArn, stateMachineArnOk := modules.RetrieveValidSpanMeta(span, "SFN", "aws.request.state.machine.arn")

		if stateMachineArnOk {
			var arn = strings.ToLower(*stateMachineArn)
			var urn = t.CreateServiceURN(arn)

			modules.SpanBuilder(span, "SFN Name Required", "State Machine", "step.function", "consumer", urn, arn)
		} else {
			_ = log.Errorf("[OTEL] [SFN]: Unable to map the Step Functions request")
			return nil
		}

		modules.InterpretHTTPError(span)
	}

	return spans
}
