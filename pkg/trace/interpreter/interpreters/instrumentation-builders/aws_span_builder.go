package instrumentationbuilders

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

// AwsSpanBuilder Map span data for the Open Telemetry service.
// This allows us to more easily change the data mapping for all the open telemetry services
func AwsSpanBuilder(span *pb.Span, serviceName string, namePrefix string, service string, kind string, urn string, arn string) {
	// Name of component displayed below the icon
	span.Meta["span.serviceName"] = fmt.Sprintf("%s: %s", namePrefix, serviceName)

	// Name of the trace displayed on the trace graph line
	span.Name = fmt.Sprintf("%s: %s", namePrefix, serviceName)

	// Displayed on the trace properties
	span.Resource = fmt.Sprintf("aws.%s", service)
	span.Type = "aws"

	// Mapping inside StackPack for capturing certain metrics
	span.Meta["span.serviceType"] = "open-telemetry"
	span.Meta["source"] = "open-telemetry"

	// Unknown
	span.Service = fmt.Sprintf("aws.%s", service)
	span.Meta["service"] = fmt.Sprintf("aws.%s", service)
	span.Meta["sts.origin"] = "open-telemetry"

	// General mapping
	span.Meta["span.kind"] = kind
	span.Meta["span.serviceURN"] = urn
	span.Meta["sts.service.identifiers"] = arn
}
