package interpret

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"strings"
)

// OpenTelemetrySQSInterpreter default span interpreter for this data structure
type OpenTelemetrySQSInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetrySQSServiceIdentifier The base identifier for this interpreter, Is also used in identifying AWS services
const OpenTelemetrySQSServiceIdentifier = "SQS"

// OpenTelemetrySQSInterpreterSpan An identifier used to direct Open Telemetry interprets to this Interpreter
var OpenTelemetrySQSInterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetrySQSServiceIdentifier)

// OpenTelemetrySQSAwsIdentifier An identifier used to map the AWS Service to the STS InterpreterServiceIdentifier
var OpenTelemetrySQSAwsIdentifier = strings.ToLower(OpenTelemetrySQSServiceIdentifier)

// MakeOpenTelemetrySQSInterpreter creates an instance of the OpenTelemetrySQS span interpreter
func MakeOpenTelemetrySQSInterpreter(config *config.Config) *OpenTelemetrySQSInterpreter {
	return &OpenTelemetrySQSInterpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetrySQSInterpreter
func (t *OpenTelemetrySQSInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	log.Debugf("[OTEL] [SQS] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		awsRegion, awsRegionOk := instrumentations.RetrieveValidSpanMeta(span, "SQS", "aws.region")
		sqsEndpoint, sqsEndpointOk := instrumentations.RetrieveValidSpanMeta(span, "SQS", "messaging.url")
		sqsQueueName, sqsQueueNameOk := instrumentations.RetrieveValidSpanMeta(span, "SQS", "messaging.destination")

		if sqsQueueNameOk && sqsEndpointOk && awsRegionOk {
			// Example Input: https://sqs.<region>.amazonaws.com/<account-id>/<queue-name>
			sqsEndpointPieces := strings.Split(*sqsEndpoint, "/")

			if len(sqsEndpointPieces) >= 3 {
				var accountID = sqsEndpointPieces[3]
				var urn = t.CreateServiceURN(*sqsEndpoint)
				var arn = strings.ToLower(
					fmt.Sprintf("https://%s.queue.amazonaws.com/%s/%s", *awsRegion, accountID, *sqsQueueName))
				var queueIdentifier = fmt.Sprintf("%s-%s-%s", *sqsQueueName, accountID, *awsRegion)

				instrumentations.AwsSpanBuilder(span, queueIdentifier, "SQS Queue", "sqs.queue", "consumer", urn, arn)
			} else {
				_ = log.Errorf("[OTEL] [SQS]: The SQS Endpoint URL is incorrect, Unable to parse %s.", sqsEndpointPieces)
				return nil
			}
		} else {
			_ = log.Errorf("[OTEL] [SQS]: Unable to map the SQS request")
			return nil
		}

		instrumentations.InterpretHTTPError(span)
	}

	return spans
}
