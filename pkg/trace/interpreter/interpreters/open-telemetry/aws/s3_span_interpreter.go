package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"strings"
)

// OpenTelemetryS3Interpreter default span interpreter for this data structure
type OpenTelemetryS3Interpreter struct {
	interpreter.Interpreter
}

const OpenTelemetryS3ServiceIdentifier = "S3"

var OpenTelemetryS3InterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryS3ServiceIdentifier)
var OpenTelemetryS3AwsIdentifier = strings.ToLower(OpenTelemetryS3ServiceIdentifier)

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

		// awsService, awsServiceOk := span.Meta["aws.service.api"]
		awsOperation, awsOperationOk := span.Meta["aws.operation"]
		s3Bucket, s3BucketOk := span.Meta["aws.request.bucket"]

		if awsOperationOk && s3BucketOk {
			var arn = strings.ToLower(fmt.Sprintf("arn:aws:s3:::%s", s3Bucket))
			var urn = t.CreateServiceURN(arn)

			OpenTelemetrySpanBuilder(
				span,
				"consumer",
				urn,
				arn,
				"s3",
				OpenTelemetryS3InterpreterSpan,
				awsOperation,
			)
		}

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
