package interpreters

import (
	"fmt"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"strings"
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

		// Retrieve the core information required to trace SQS

		// awsService, awsServiceOk := span.Meta["aws.service.api"]
		awsRegion, awsRegionOk := span.Meta["aws.region"]
		awsOperation, awsOperationOk := span.Meta["aws.operation"]
		sqsEndpoint, sqsEndpointOk := span.Meta["messaging.url"]
		sqsQueueName, sqsQueueNameOk := span.Meta["messaging.destination"]

		if sqsQueueNameOk && sqsEndpointOk && awsOperationOk && awsRegionOk {
			sqsEndpointPieces := strings.Split(sqsEndpoint, "/") // Example Input: https://sqs.<region>.amazonaws.com/<account-id>/<queue-name>

			if len(sqsEndpointPieces) >= 3 {
				var accountId = sqsEndpointPieces[3]
				var urn = t.CreateServiceURN(sqsEndpoint)
				var arn = strings.ToLower(
					fmt.Sprintf("https://%s.queue.amazonaws.com/%s/%s", awsRegion, accountId, sqsQueueName))

				OpenTelemetryConsumerMappings(
					span,
					urn,
					arn,
					"sqs",
					OpenTelemetrySQSInterpreterSpan,
					awsOperation,
				)
			}
		}

		t.interpretHTTPError(span)
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
