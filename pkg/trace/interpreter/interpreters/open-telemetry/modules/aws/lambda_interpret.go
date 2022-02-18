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
	log.Debugf("[OTEL] [LAMBDA-ENTRY] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		if arn, ok := span.Meta["faas.id"]; arn != "" && ok {
			var urn = t.CreateServiceURN(strings.ToLower(arn))
			arn = strings.ToLower(arn)

			modules.OpenTelemetrySpanBuilder(
				span,
				"producer",
				"execute",
				"lambda",
				"Lambda Function",
				"Serverless",
				"test-eu-west-1",
				urn,
				arn,
			)

			//  modules.SpanBuilder(
			//  	span,
			//  	"producer",
			//  	"lambda",
			//  	"execute",
			//  	urn,
			//  	arn,
			//  )
		} else {
			_ = log.Errorf("[OTEL] [LAMBDA-ENTRY]: Unable to determine the root Lambda Span")

			if !ok {
				_ = log.Errorf("[OTEL] [LAMBDA-ENTRY]: 'faas.id' is not found in the span meta data, this value is required.")
			}

			return nil
		}

		t.interpretHTTPError(span)
	}

	return spans
}

func (t *OpenTelemetryLambdaEntryInterpreter) interpretHTTPError(span *pb.Span) {
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
