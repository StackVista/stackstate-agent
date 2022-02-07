package opentelemetry

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// OpenTelemetrySpanBuilder An generic function to map Open Telemetry AWS service traces to a format that STS understands
func OpenTelemetrySpanBuilder(span *pb.Span, spanKind string, spanEvent string, serviceID string, serviceType string, serviceLayer string, serviceDomain string, urn string, identifier string) {
	var mappingKey = "open.telemetry"

	span.Type = serviceType
	span.Resource = fmt.Sprintf("aws.%s", serviceID)
	span.Service = fmt.Sprintf("%s.%s", mappingKey, serviceID)

	span.Meta["sts.service.identifiers"] = identifier
	span.Meta["span.serviceURN"] = urn

	span.Meta["service"] = span.Service
	span.Meta["layer"] = serviceLayer
	span.Meta["domain"] = serviceDomain

	span.Meta["span.serviceType"] = serviceType
	span.Meta["span.serviceName"] = fmt.Sprintf("%s.%s.%s", mappingKey, serviceID, spanEvent)
	span.Meta["span.kind"] = spanKind
	span.Meta["span.layer"] = serviceLayer
	span.Meta["span.domain"] = serviceDomain
}
