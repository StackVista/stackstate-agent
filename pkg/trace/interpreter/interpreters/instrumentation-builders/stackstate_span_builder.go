package instrumentationbuilders

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// StackStateSpanBuilder Map span data for the Open Telemetry service.
// This allows us to more easily change the data mapping for all the open telemetry services
func StackStateSpanBuilder(span *pb.Span, tracePerspectiveName string, serviceType string, serviceName string, serviceIdentifier string, resourceName string, kind string, urn string) {
	span.Meta["span.serviceName"] = serviceName

	// Name of the trace displayed on the trace graph line
	span.Name = tracePerspectiveName

	// Displayed on the trace properties
	span.Resource = resourceName
	span.Type = "custom-trace"

	// Mapping inside StackPack for capturing certain metrics
	span.Meta["span.serviceType"] = "open-telemetry"
	span.Meta["source"] = "open-telemetry"

	// Unknown
	span.Meta["sts.origin"] = "open-telemetry"
	span.Service = serviceType
	span.Meta["service"] = serviceType

	// General mapping
	span.Meta["span.kind"] = kind
	span.Meta["span.serviceURN"] = urn
	span.Meta["sts.service.identifiers"] = serviceIdentifier
}
