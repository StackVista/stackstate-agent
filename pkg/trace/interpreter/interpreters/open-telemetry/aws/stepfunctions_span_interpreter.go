package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
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
	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		// awsService, awsServiceOk := span.Meta["aws.service.api"]
		awsOperation, awsOperationOk := span.Meta["aws.operation"]
		stateMachineArn, stateMachineArnOk := span.Meta["aws.request.state.machine.arn"]

		if awsOperationOk && stateMachineArnOk {
			var arn = strings.ToLower(stateMachineArn)
			var urn = t.CreateServiceURN(arn)

			OpenTelemetrySpanBuilder(
				span,
				"consumer",
				awsOperation,
				"step.function",
				"Step Functions State Machine",
				"Serverless",
				"test-eu-west-1",
				urn,
				arn,
			)
		} else {
			_ = log.Errorf("[OTEL] [SFN]: Unable to map the Step Functions request")

			if !awsOperationOk {
				_ = log.Errorf("[OTEL] [SFN]: 'aws.operation' is not found in the span meta data, this value is required.")
			}
			if !stateMachineArnOk {
				_ = log.Errorf("[OTEL] [SFN]: 'aws.request.state.machine.arn' is not found in the span meta data, this value is required.")
			}
		}

		t.interpretHTTPError(span)
	}

	return spans
}

func (t *OpenTelemetryStepFunctionsInterpreter) interpretHTTPError(span *pb.Span) {
	if span.Error != 0 {
		if httpStatus, found := span.Metrics["http.status_code"]; found {
			if httpStatus >= 400 && httpStatus < 500 {
				span.Meta["span.errorClass"] = "4xx"
			} else if httpStatus >= 500 {
				span.Meta["span.errorClass"] = "5xx"
			}
		}
	}
}
