package instrumentations

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// SpanBuilder An generic function to map Open Telemetry AWS service traces to a format that STS understands
func SpanBuilder(span *pb.Span, kind string, resource string, event string, urn string, arn string) {
	// Producer or Consumer
	span.Meta["span.kind"] = kind

	// Event, Example: open.telemetry.s3.
	span.Meta["span.serviceName"] = fmt.Sprintf("%s.%s.%s", "open.telemetry", resource, event)

	// Service Types
	span.Type = "open-telemetry"
	span.Meta["span.serviceType"] = "open-telemetry"

	span.Resource = fmt.Sprintf("aws.%s", resource)

	span.Meta["service"] = fmt.Sprintf("%s.%s", "open.telemetry", resource)
	span.Service = fmt.Sprintf("%s.%s", "open.telemetry", resource)

	span.Meta["span.serviceURN"] = urn
	span.Meta["sts.service.identifiers"] = arn
}
