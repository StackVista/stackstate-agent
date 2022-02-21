package modules

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// SpanBuilder An generic function to map Open Telemetry AWS service traces to a format that STS understands
func SpanBuilder(span *pb.Span, kind string, resource string, event string, urn string, arn string) {
	span.Meta["span.kind"] = kind
	span.Meta["span.serviceName"] = fmt.Sprintf("%s.%s.%s", "open.telemetry", resource, event)
	span.Type = "open-telemetry"
	span.Meta["span.serviceType"] = "open-telemetry"
	span.Resource = fmt.Sprintf("aws.%s", resource)
	span.Meta["service"] = fmt.Sprintf("%s.%s", "open.telemetry", resource)
	span.Service = fmt.Sprintf("%s.%s", "open.telemetry", resource)
	span.Meta["span.serviceURN"] = urn
	span.Meta["sts.service.identifiers"] = arn
}

// RetrieveValidSpanMeta TODO:
func RetrieveValidSpanMeta(span *pb.Span, logName string, target string) (*string, bool) {
	value, ok := span.Meta[target]
	if ok && len(value) > 0 {
		return &value, true
	}

	if !ok {
		_ = log.Errorf("[OTEL] [%s]: '%s' is not found in the span meta data, this value is required.", logName, target)
	}

	return nil, false
}

// InterpretHTTPError TODO:
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
