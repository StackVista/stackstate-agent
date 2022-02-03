package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"strings"
)

// OpenTelemetryHTTPInterpreter default span interpreter for this data structure
type OpenTelemetryHTTPInterpreter struct {
	interpreter.Interpreter
}

const OpenTelemetryHTTPServiceIdentifier = "Http"

var OpenTelemetryHTTPInterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryHTTPServiceIdentifier)

// MakeOpenTelemetryHTTPInterpreter creates an instance of the OpenTelemetry HTTP span interpreter
func MakeOpenTelemetryHTTPInterpreter(config *config.Config) *OpenTelemetryHTTPInterpreter {
	return &OpenTelemetryHTTPInterpreter{interpreter.Interpreter{Config: config}}
}

func sanitizeURL(url string) string {
	urlLowerCase := strings.ToLower(url)
	hashFiltered := strings.Split(urlLowerCase, "#")
	queryFiltered := strings.Split(hashFiltered[0], "?")
	httpFiltered := strings.Replace(queryFiltered[0], "http://", "", 1)
	httpsFiltered := strings.Replace(httpFiltered, "https://", "", 1)
	return httpsFiltered
}

// Interpret performs the interpretation for the OpenTelemetryHTTPInterpreter
func (t *OpenTelemetryHTTPInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		httpURL, httpURLOk := span.Meta["http.url"]
		httpMethod, httpMethodOk := span.Meta["http.method"]

		if httpURLOk && httpMethodOk {
			var url = sanitizeURL(httpURL)
			var urn = t.CreateServiceURN(fmt.Sprintf("lambda-http-request/%s/%s", url, httpMethod))

			OpenTelemetrySpanBuilder(
				span,
				"consumer",
				urn,
				url,
				"lambda.http",
				OpenTelemetryHTTPInterpreterSpan,
				httpMethod,
			)
		}

		t.interpretHTTPError(span)
	}

	return spans
}

func (t *OpenTelemetryHTTPInterpreter) interpretHTTPError(span *pb.Span) {
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
