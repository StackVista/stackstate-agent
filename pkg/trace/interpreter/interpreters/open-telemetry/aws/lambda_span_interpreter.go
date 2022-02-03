package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"strings"
)

// OpenTelemetryLambdaInterpreter default span interpreter for this data structure
type OpenTelemetryLambdaInterpreter struct {
	interpreter.Interpreter
}

const OpenTelemetryLambdaServiceIdentifier = "Lambda"

var OpenTelemetryLambdaInterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryLambdaServiceIdentifier)
var OpenTelemetryLambdaAwsIdentifier = strings.ToLower(OpenTelemetryLambdaServiceIdentifier)

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

		// Invoke will contain data to another Lambda function being invoked
		if lambdaName := span.Name; span.Meta["aws.operation"] == "invoke" && lambdaName != "" {
			var functionName = span.Meta["aws.request.function.name"]
			var accountID = span.Meta["aws.account.id"]
			var region = span.Meta["aws.region"]

			var arn = strings.ToLower(fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", region, accountID, functionName))
			var urn = t.CreateServiceURN(arn)

			OpenTelemetrySpanBuilder(
				span,
				"consumer",
				urn,
				arn,
				"lambda.function",
				OpenTelemetryLambdaInterpreterSpan,
				"invoke",
			)

		} else if arn, ok := span.Meta["faas.id"]; arn != "" && ok {
			var urn = t.CreateServiceURN(strings.ToLower(arn))
			arn = strings.ToLower(arn)

			OpenTelemetrySpanBuilder(
				span,
				"producer",
				urn,
				arn,
				"lambda.function",
				OpenTelemetryLambdaInterpreterSpan,
				"execute",
			)
		}

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
