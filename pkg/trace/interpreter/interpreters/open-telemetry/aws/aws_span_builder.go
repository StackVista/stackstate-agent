package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// OpenTelemetrySpanBuilder An generic function to map Open Telemetry AWS service traces to a format that STS understands
func OpenTelemetrySpanBuilder(span *pb.Span, spanKind string, spanEvent string, serviceID string, serviceType string, serviceLayer string, serviceDomain string, _ string, arn string) {
	var mappingKey = "open.telemetry"

	span.Type = serviceType
	span.Resource = fmt.Sprintf("%s.%s", serviceID, spanEvent)
	span.Service = fmt.Sprintf("%s.%s", mappingKey, serviceID)

	span.Meta["span.serviceName"] = fmt.Sprintf("%s.%s.%s", mappingKey, serviceID, spanEvent)
	span.Meta["span.kind"] = spanKind
	span.Meta["sts.service.identifiers"] = arn

	span.Meta["span.layer"] = serviceLayer
	span.Meta["layer"] = serviceLayer

	span.Meta["span.domain"] = serviceDomain
	span.Meta["domain"] = serviceDomain

	// span.Meta["service"] = span.Service
	// span.Meta["span.serviceType"] = serviceType
	// span.Meta["span.serviceURN"] = urn
}
