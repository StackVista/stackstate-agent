package interpreters

import (
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// OpenTelemetrySQSInterpreter default span interpreter for this data structure
type OpenTelemetrySQSInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetrySQSInterpreterSpan is the name used for matching this interpreter
const OpenTelemetrySQSInterpreterSpan = "openTelemetrySQS"

// MakeOpenTelemetrySQSInterpreter creates an instance of the OpenTelemetrySQS span interpreter
func MakeOpenTelemetrySQSInterpreter(config *config.Config) *OpenTelemetrySQSInterpreter {
	return &OpenTelemetrySQSInterpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetrySQSInterpreter
func (t *OpenTelemetrySQSInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		// SQS Name for example: SQSQueueName
		if queueName, ok := span.Meta["messaging.destination"]; queueName != "" && ok {
			span.Meta["sts.service.name"] = queueName
			span.Service = queueName
		}

		if url, ok := span.Meta["messaging.url"]; url != "" && ok {
			var urn = t.CreateServiceURN(url)
			span.Meta["sts.service.URN"] = urn
			span.Meta["sts.service.identifiers"] = urn // TODO: Possible arn
		}

		// AWS Service used like SQS, SNS etc
		if service, ok := span.Meta["aws.service.api"]; service != "" && ok {
			span.Meta["span.kind"] = "consumer"
			span.Resource = service
		}

		// AWS Action taken for example MessageSend
		if action, ok := span.Meta["aws.operation"]; action != "" && ok {
			span.Type = action
		}

		t.interpretHTTPError(span)

		span.Meta["span.serviceType"] = OpenTelemetrySQSInterpreterSpan
	}

	return spans
}

func (t *OpenTelemetrySQSInterpreter) interpretHTTPError(span *pb.Span) {
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
