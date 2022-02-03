package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"strings"
)

// OpenTelemetrySNSInterpreter default span interpreter for this data structure
type OpenTelemetrySNSInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetrySNSServiceIdentifier The base identifier for this interpreter, Is also used in identifying AWS services
const OpenTelemetrySNSServiceIdentifier = "SNS"

// OpenTelemetrySNSInterpreterSpan An identifier used to direct Open Telemetry interprets to this Interpreter
var OpenTelemetrySNSInterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetrySNSServiceIdentifier)

// OpenTelemetrySNSAwsIdentifier An identifier used to map the AWS Service to the STS InterpreterServiceIdentifier
var OpenTelemetrySNSAwsIdentifier = strings.ToLower(OpenTelemetrySNSServiceIdentifier)

// MakeOpenTelemetrySNSInterpreter creates an instance of the OpenTelemetrySNS span interpreter
func MakeOpenTelemetrySNSInterpreter(config *config.Config) *OpenTelemetrySNSInterpreter {
	return &OpenTelemetrySNSInterpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetrySNSInterpreter
func (t *OpenTelemetrySNSInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		// awsService, awsServiceOk := span.Meta["aws.service.api"]
		awsOperation, awsOperationOk := span.Meta["aws.operation"]
		topicArn, topicArnOk := span.Meta["aws.request.topic.arn"]

		if awsOperationOk && topicArnOk {
			var urn = t.CreateServiceURN(strings.ToLower(topicArn))

			OpenTelemetrySpanBuilder(
				span,
				"consumer",
				urn,
				strings.ToLower(topicArn),
				"sns",
				OpenTelemetrySNSInterpreterSpan,
				awsOperation,
			)
		}

		t.interpretHTTPError(span)
	}

	return spans
}

func (t *OpenTelemetrySNSInterpreter) interpretHTTPError(span *pb.Span) {
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
