package modules

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// SpanBuilder Map span data for the Open Telemetry service.
// This allows us to more easily change the data mapping for all the open telemetry services
func SpanBuilder(span *pb.Span, serviceName string, namePrefix string, service string, kind string, urn string, arn string) {
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

// RetrieveValidSpanMeta Retrieve span data or display a message saying what is missing in the agent logs and returning false
func RetrieveValidSpanMeta(span *pb.Span, logName string, target string) (*string, bool) {
	value, ok := span.Meta[target]
	if ok && len(value) > 0 {
		log.Debugf("[OTEL] [%s]: '%s' was found for this module, value content: %s", logName, target, value)

		return &value, true
	}

	_ = log.Errorf("[OTEL] [%s]: '%s' is not found in the span meta data, this value is required.", logName, target)

	return nil, false
}

// InterpretHTTPError Maps a proper error class if the instrumentation contains a error
func InterpretHTTPError(span *pb.Span) {
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
