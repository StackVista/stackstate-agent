package interpreters

import (
	"fmt"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"strings"
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

		fmt.Println("Process Lambda Span Interpreter")

		if span.Meta["aws.operation"] == "invoke" {
			fmt.Println("A")
			span.Meta["span.kind"] = "consumer"

			var functionName = span.Meta["aws.request.function.name"]
			var accountId = span.Meta["aws.account.id"]
			var region = span.Meta["aws.region"]

			var arn = strings.ToLower(fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", region, accountId, functionName))
			var urn = t.CreateServiceURN(arn)

			span.Meta["span.serviceURN"] = urn
			span.Meta["sts.service.identifiers"] = arn

		} else {
			fmt.Println("B")
			span.Meta["span.kind"] = "producer"

			if arn, ok := span.Meta["faas.id"]; arn != "" && ok {
				var urn = t.CreateServiceURN(strings.ToLower(arn))
				span.Meta["span.serviceURN"] = urn
				span.Meta["sts.service.identifiers"] = strings.ToLower(arn)
			}
		}

		if lambdaName := span.Name; lambdaName != "" {
			span.Meta["span.serviceName"] = lambdaName
			span.Meta["service"] = lambdaName
			span.Service = "Lambda"
			span.Resource = "Lambda"
			span.Meta["aws.service.api"] = "lambda"
			span.Meta["aws.operation"] = "invoke"
			span.Type = "invoke"
		}

		span.Meta["span.serviceType"] = OpenTelemetryLambdaInterpreterSpan

		t.interpretHTTPError(span)
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
