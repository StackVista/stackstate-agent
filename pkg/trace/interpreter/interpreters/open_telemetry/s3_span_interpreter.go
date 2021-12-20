package interpreters

import (
	"fmt"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"strings"
)

// OpenTelemetryS3Interpreter default span interpreter for this data structure
type OpenTelemetryS3Interpreter struct {
	interpreter.Interpreter
}

// OpenTelemetryS3InterpreterSpan is the name used for matching this interpreter
const OpenTelemetryS3InterpreterSpan = "openTelemetryS3"

// MakeOpenTelemetryS3Interpreter creates an instance of the OpenTelemetryS3 span interpreter
func MakeOpenTelemetryS3Interpreter(config *config.Config) *OpenTelemetryS3Interpreter {
	return &OpenTelemetryS3Interpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetryS3Interpreter
func (t *OpenTelemetryS3Interpreter) Interpret(spans []*pb.Span) []*pb.Span {
	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		span.Meta["span.kind"] = "consumer"

		fmt.Println("Process s3 Span Interpreter")

		// Retrieve the core information required to trace SNS
		awsOperation, awsOperationOk := span.Meta["aws.operation"]
		awsService, awsServiceOk := span.Meta["aws.service.api"]
		s3Bucket, s3BucketOk := span.Meta["aws.request.bucket"]

		if awsServiceOk && awsOperationOk && s3BucketOk {
			var arn = "arn:aws:s3:::" + strings.ToLower(s3Bucket)
			var urn = t.CreateServiceURN(arn)

			span.Type = awsOperation
			span.Service = awsService
			span.Resource = awsService

			span.Meta["sts.service.identifiers"] = arn
			span.Meta["span.serviceURN"] = urn
			span.Meta["span.serviceName"] = awsService // TODO, Change to section in url arn:aws:sns:eu-west-1:965323806078:open-telemetry-dev-OpenTelemetrySNS
			span.Meta["service"] = awsService
		}

		span.Meta["span.serviceType"] = OpenTelemetryS3InterpreterSpan

		t.interpretHTTPError(span)
	}

	return spans
}

func (t *OpenTelemetryS3Interpreter) interpretHTTPError(span *pb.Span) {
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
