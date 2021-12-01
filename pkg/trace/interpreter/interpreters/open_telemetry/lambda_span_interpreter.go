package interpreters

import (
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// OpenTelemetryLambdaInterpreter default span interpreter for this data structure
type OpenTelemetryLambdaInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetryLambdaInterpreterSpan is the name used for matching this interpreter
const OpenTelemetryLambdaInterpreterSpan = "openTelemetryLambda"

// MakeOpenTelemetryLambdaInterpreter creates an instance of the OpenTelemetry Lambda span interpreter
func MakeOpenTelemetryLambdaInterpreter(config *config.Config) *OpenTelemetryLambdaInterpreter {
	return &OpenTelemetryLambdaInterpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetryLambdaInterpreter
func (t *OpenTelemetryLambdaInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		span.Meta["span.kind"] = "producer"

		// Lambda Function Name
		if lambdaName := span.Name; lambdaName != "" {
			span.Meta["sts.service.name"] = lambdaName
			span.Service = lambdaName
		}

		if arn, ok := span.Meta["faas.id"]; arn != "" && ok {
			// Service
			span.Meta["aws.service.api"] = "Lambda"
			span.Resource = "Lambda"

			// Action
			span.Meta["aws.operation"] = "Invoke"
			span.Type = "Invoke"

			// URN & ARN
			var urn = t.CreateServiceURN(arn)
			span.Meta["sts.service.URN"] = urn
			span.Meta["sts.service.identifiers"] = urn // TODO: Possible arn
		}

		t.interpretHTTPError(span)

		span.Meta["span.serviceType"] = OpenTelemetryLambdaInterpreterSpan
	}

	return spans
}

func (t *OpenTelemetryLambdaInterpreter) interpretHTTPError(span *pb.Span) {
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
