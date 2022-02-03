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

// OpenTelemetryLambdaInterpreter default span interpreter for this data structure
type OpenTelemetryLambdaInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetryLambdaServiceIdentifier The base identifier for this interpreter, Is also used in identifying AWS services
const OpenTelemetryLambdaServiceIdentifier = "Lambda"

// OpenTelemetryLambdaInterpreterSpan An identifier used to direct Open Telemetry interprets to this Interpreter
var OpenTelemetryLambdaInterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryLambdaServiceIdentifier)

// OpenTelemetryLambdaAwsIdentifier An identifier used to map the AWS Service to the STS InterpreterServiceIdentifier
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
			functionName, functionNameOk := span.Meta["aws.request.function.name"]
			accountID, accountIDOk := span.Meta["aws.account.id"]
			region, regionOk := span.Meta["aws.region"]

			if functionNameOk && accountIDOk && regionOk {
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

				t.interpretHTTPError(span)

				return nil
			}

			_ = log.Errorf("[OTEL] [LAMBDA]: Unable to map the invoked Lambda Function")

			if !functionNameOk {
				_ = log.Errorf("[OTEL] [LAMBDA]: 'aws.request.function.name' is not found in the span meta data, this value is required.")
			}
			if !accountIDOk {
				_ = log.Errorf("[OTEL] [LAMBDA]: 'aws.account.id' is not found in the span meta data, this value is required.")
			}
			if !regionOk {
				_ = log.Errorf("[OTEL] [LAMBDA]: 'aws.region' is not found in the span meta data, this value is required.")
			}
		}

		if arn, ok := span.Meta["faas.id"]; arn != "" && ok {
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
		} else {
			_ = log.Errorf("[OTEL] [LAMBDA-CORE]: Unable to determine the root Lambda Span")

			if !ok {
				_ = log.Errorf("[OTEL] [LAMBDA-CORE]: 'faas.id' is not found in the span meta data, this value is required.")
			}
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
