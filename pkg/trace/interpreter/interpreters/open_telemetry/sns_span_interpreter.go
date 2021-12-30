package interpreters

import (
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"strings"
)

// OpenTelemetrySNSInterpreter default span interpreter for this data structure
type OpenTelemetrySNSInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetrySNSInterpreterSpan is the name used for matching this interpreter
const OpenTelemetrySNSInterpreterSpan = "openTelemetrySNS"

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

		// Retrieve the core information required to trace SNS

		// awsService, awsServiceOk := span.Meta["aws.service.api"]
		awsOperation, awsOperationOk := span.Meta["aws.operation"]
		topicArn, topicArnOk := span.Meta["aws.request.topic.arn"]

		if awsOperationOk && topicArnOk {
			var urn = t.CreateServiceURN(strings.ToLower(topicArn))

			OpenTelemetryConsumerMappings(
				span,
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
