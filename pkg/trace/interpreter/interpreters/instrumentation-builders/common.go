package instrumentationbuilders

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// GetSpanMeta Retrieve span data or display a message saying what is missing in the agent logs and returning false
func GetSpanMeta(logName string, span *pb.Span, spanMetaTarget string) (*string, bool) {
	value, ok := span.Meta[spanMetaTarget]
	if ok && len(value) > 0 {
		log.Debugf("[OTEL] [%s]: '%s' was found for this module, value content: %s", logName, spanMetaTarget, value)
		return &value, true
	}

	_ = log.Errorf("[OTEL] [%s]: '%s' is not found in the span meta data, this value is required.", logName, spanMetaTarget)
	return nil, false
}

// InterpretSpanHTTPError Maps a proper error class if the instrumentation contains a error
func InterpretSpanHTTPError(span *pb.Span) {
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
