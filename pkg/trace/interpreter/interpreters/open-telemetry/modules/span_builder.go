package modules

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// SpanBuilder An generic function to map Open Telemetry AWS service traces to a format that STS understands
//  func SpanBuilder(span *pb.Span, kind string, resource string, event string, urn string, arn string) {
//  	// Producer or Consumer
//  	span.Meta["span.kind"] = kind
//
//  	// Event, Example: open.telemetry.s3.
//  	span.Meta["span.serviceName"] = fmt.Sprintf("%s.%s.%s", "open.telemetry", resource, event)
//
//  	// Service Types
//  	span.Type = "open-telemetry"
//  	span.Meta["span.serviceType"] = "open-telemetry"
//
//  	span.Resource = fmt.Sprintf("aws.%s", resource)
//
//  	span.Meta["service"] = fmt.Sprintf("%s.%s", "open.telemetry", resource)
//  	span.Service = fmt.Sprintf("%s.%s", "open.telemetry", resource)
//
//  	span.Meta["span.serviceURN"] = urn
//  	span.Meta["sts.service.identifiers"] = arn
//  }

// OpenTelemetrySpanBuilder An generic function to map Open Telemetry AWS service traces to a format that STS understands
func OpenTelemetrySpanBuilder(span *pb.Span, kind string, event string, resource string, sType string, layer string, domain string, urn string, id string) {
	var mappingKey = "open.telemetry"

	span.Type = sType
	span.Resource = fmt.Sprintf("aws.%s", resource)
	span.Service = fmt.Sprintf("%s.%s", mappingKey, resource) // aws.

	span.Meta["service"] = span.Service
	span.Meta["sts.service.identifiers"] = id

	span.Meta["span.serviceURN"] = urn
	span.Meta["span.serviceType"] = sType // TODO: General for all services -> Agent service mapping extend case to capture this type
	span.Meta["span.serviceName"] = fmt.Sprintf("%s.%s.%s", mappingKey, resource, event)
	span.Meta["span.kind"] = kind

	span.Meta["layer"] = layer
	span.Meta["domain"] = domain
}
