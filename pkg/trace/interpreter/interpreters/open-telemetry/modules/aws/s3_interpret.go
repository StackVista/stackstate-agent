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

	/*
			lambda -> root span
			aws-sdk sqs -> parentId === lambda
			http -> parentId === aws-sdk

			lambda -> root span
			http -> parentId === lambda


			[SQS] <- Lambda -> SQS -> (http.post.event removed from span)
			[SQS] <- Lambda -> SQS (With http status)


		404 -> sqs does not exist
		400 -> access denied Permission
	*/

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		// TODO: => interpretHTTPError
		// span.Error = 400

		// awsService, awsServiceOk := span.Meta["aws.service.api"]
		awsOperation, awsOperationOk := span.Meta["aws.operation"]
		s3Bucket, s3BucketOk := span.Meta["aws.request.bucket"]

		if awsOperationOk && s3BucketOk {
			var arn = strings.ToLower(fmt.Sprintf("arn:aws:s3:::%s", s3Bucket))
			var urn = t.CreateServiceURN(arn)

			modules.OpenTelemetrySpanBuilder(
				span,
				"consumer",
				awsOperation,
				"s3",
				"S3 Bucket",
				"Storage",
				"test-eu-west-1",
				urn,
				arn,
			)

			//  modules.SpanBuilder(
			//  	span,
			//  	"consumer",
			//  	"s3",
			//  	awsOperation,
			//  	urn,
			//  	arn,
			//  )
		} else {
			_ = log.Errorf("[OTEL] [S3]: Unable to map the S3 request")

			if !awsOperationOk {
				_ = log.Errorf("[OTEL] [S3]: 'aws.operation' is not found in the span meta data, this value is required.")
			}
			if !s3BucketOk {
				_ = log.Errorf("[OTEL] [S3]: 'aws.request.bucket' is not found in the span meta data, this value is required.")
			}

			return nil
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
