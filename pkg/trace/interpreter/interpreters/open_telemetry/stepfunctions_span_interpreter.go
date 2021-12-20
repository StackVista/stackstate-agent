package interpreters

import (
	"fmt"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"strings"
)

// OpenTelemetryStepFunctionsInterpreter default span interpreter for this data structure
type OpenTelemetryStepFunctionsInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetryStepFunctionsInterpreterSpan is the name used for matching this interpreter
const OpenTelemetryStepFunctionsInterpreterSpan = "openTelemetryStepFunctions"

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

		fmt.Println("Process Step Functions Span Interpreter")

		span.Meta["span.kind"] = "consumer"

		// Retrieve the core information required to trace SNS
		awsOperation, awsOperationOk := span.Meta["aws.operation"]
		awsService, awsServiceOk := span.Meta["aws.service.api"]
		stateMachineArn, stateMachineArnOk := span.Meta["aws.request.state.machine.arn"]

		if awsServiceOk && awsOperationOk && stateMachineArnOk {
			var arn = strings.ToLower(stateMachineArn)
			var urn = t.CreateServiceURN(arn)

			span.Type = awsOperation
			span.Service = awsService
			span.Resource = awsService

			span.Meta["sts.service.identifiers"] = arn
			span.Meta["span.serviceURN"] = urn
			span.Meta["span.serviceName"] = awsService
			span.Meta["service"] = awsService
		}

		span.Meta["span.serviceType"] = OpenTelemetryStepFunctionsInterpreterSpan

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
