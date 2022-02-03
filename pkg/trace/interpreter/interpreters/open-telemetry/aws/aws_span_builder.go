package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// OpenTelemetrySpanBuilder An generic function to map Open Telemetry AWS service traces to a format that STS understands
func OpenTelemetrySpanBuilder(span *pb.Span,
	kind string,
	urn string,
	arn string,
	awsService string,
	serviceType string,
	awsOperation string) {
	_ = log.Warnf(fmt.Sprintf("OTEL Map: '%s'", awsService))

	var mappingKey = "open.telemetry"

	span.Type = fmt.Sprintf("aws.%s", awsService)
	span.Service = fmt.Sprintf("%s.%s", mappingKey, awsService)
	span.Resource = fmt.Sprintf("aws.%s.%s", awsService, awsOperation)

	span.Meta["service"] = span.Service
	span.Meta["span.kind"] = kind
	span.Meta["span.serviceName"] = span.Service
	span.Meta["span.serviceType"] = serviceType
	span.Meta["span.serviceURN"] = urn
	span.Meta["sts.service.identifiers"] = arn
}
