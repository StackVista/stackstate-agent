package modules

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/trace/api"
	config "github.com/DataDog/datadog-agent/pkg/trace/interpreter/config"
	interpreter "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters"
	instrumentationbuilders "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentation-builders"
	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/DataDog/datadog-agent/pkg/util/log"
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
	log.Debugf("[OTEL] [SNS] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		topicArn, topicArnOk := instrumentationbuilders.GetSpanMeta("SNS", span, "aws.request.topic.arn")

		if topicArnOk {
			var urn = t.CreateServiceURN(strings.ToLower(*topicArn))
			var arn = strings.ToLower(*topicArn)

			arnParts := strings.Split(arn, ":")

			if len(arnParts) >= 6 {
				topicName := arnParts[5]
				instrumentationbuilders.AwsSpanBuilder(span, topicName, "SNS", "sns", "consumer", urn, arn)
			} else {
				_ = log.Errorf("[OTEL] [SNS]: 'arn' invalid structure supplied '%s'", arn)
				return nil
			}
		} else {
			_ = log.Errorf("[OTEL] [SNS]: Unable to map the SNS request")
			return nil
		}

		instrumentationbuilders.InterpretSpanHTTPError(span)
	}

	return spans
}
