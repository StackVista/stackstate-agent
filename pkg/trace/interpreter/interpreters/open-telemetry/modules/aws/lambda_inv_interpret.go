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
	log.Debugf("[OTEL] [LAMBDA] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		functionName, functionNameOk := modules.RetrieveValidSpanMeta(span, "LAMBDA", "aws.request.function.name")
		accountID, accountIDOk := modules.RetrieveValidSpanMeta(span, "LAMBDA", "aws.account.id")
		region, regionOk := modules.RetrieveValidSpanMeta(span, "LAMBDA", "aws.region")
		// awsOperation, awsOperationOk := modules.RetrieveValidSpanMeta(span, "LAMBDA", "aws.operation")

		// Invoke will contain data to another Lambda function being invoked
		if functionNameOk && accountIDOk && regionOk && span.Meta["aws.operation"] == "invoke" {
			var arn = strings.ToLower(fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s", *region, *accountID, *functionName))
			var urn = t.CreateServiceURN(arn)

			modules.SpanBuilder(span, *functionName, "Lambda", "lambda", "consumer", urn, arn)
		} else {
			_ = log.Errorf("[OTEL] [LAMBDA]: Unable to map the invoked Lambda Function")
			return nil
		}

		modules.InterpretHTTPError(span)
	}

	return spans
}
