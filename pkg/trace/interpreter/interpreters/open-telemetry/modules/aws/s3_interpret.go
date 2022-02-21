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

// OpenTelemetryS3Interpreter default span interpreter for this data structure
type OpenTelemetryS3Interpreter struct {
	interpreter.Interpreter
}

// OpenTelemetryS3ServiceIdentifier The base identifier for this interpreter, Is also used in identifying AWS services
const OpenTelemetryS3ServiceIdentifier = "S3"

// OpenTelemetryS3InterpreterSpan An identifier used to direct Open Telemetry interprets to this Interpreter
var OpenTelemetryS3InterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryS3ServiceIdentifier)

// OpenTelemetryS3AwsIdentifier An identifier used to map the AWS Service to the STS InterpreterServiceIdentifier
var OpenTelemetryS3AwsIdentifier = strings.ToLower(OpenTelemetryS3ServiceIdentifier)

// MakeOpenTelemetryS3Interpreter creates an instance of the OpenTelemetryS3 span interpreter
func MakeOpenTelemetryS3Interpreter(config *config.Config) *OpenTelemetryS3Interpreter {
	return &OpenTelemetryS3Interpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetryS3Interpreter
func (t *OpenTelemetryS3Interpreter) Interpret(spans []*pb.Span) []*pb.Span {
	log.Debugf("[OTEL] [S3] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		// awsOperation, awsOperationOk := modules.RetrieveValidSpanMeta(span, "S3", "aws.operation")
		s3Bucket, s3BucketOk := modules.RetrieveValidSpanMeta(span, "S3", "aws.request.bucket")

		if s3BucketOk {
			var arn = strings.ToLower(fmt.Sprintf("arn:aws:s3:::%s", *s3Bucket))
			var urn = t.CreateServiceURN(arn)

			modules.SpanBuilder(span, *s3Bucket, "S3", "s3", "consumer", urn, arn)
		} else {
			_ = log.Errorf("[OTEL] [S3]: Unable to map the S3 request")
			return nil
		}

		modules.InterpretHTTPError(span)
	}

	return spans
}
