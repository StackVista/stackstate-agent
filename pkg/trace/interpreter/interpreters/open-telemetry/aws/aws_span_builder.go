package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// OpenTelemetrySpanBuilder An generic function to map Open Telemetry AWS service traces to a format that STS understands
func OpenTelemetrySpanBuilder(span *pb.Span, kind string, event string, resource string, sType string, layer string, domain string, urn string, id string) {
	var mappingKey = "open.telemetry"

	span.Type = sType
	span.Resource = fmt.Sprintf("aws.%s", resource)
	span.Service = fmt.Sprintf("%s.%s", mappingKey, resource)
	span.Meta["service"] = span.Service

	span.Meta["sts.service.identifiers"] = id
	span.Meta["span.serviceURN"] = urn

	span.Meta["span.serviceType"] = sType
	span.Meta["span.serviceName"] = fmt.Sprintf("%s.%s.%s", mappingKey, resource, event)
	span.Meta["span.kind"] = kind

	span.Meta["layer"] = layer
	span.Meta["domain"] = domain
}
