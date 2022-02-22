package http

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/modules"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"strings"
)

// OpenTelemetryHTTPInterpreter default span interpreter for this data structure
type OpenTelemetryHTTPInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetryHTTPServiceIdentifier The base identifier for this interpreter, Is also used in identifying AWS services
const OpenTelemetryHTTPServiceIdentifier = "Http"

// OpenTelemetryHTTPInterpreterSpan An identifier used to direct Open Telemetry interprets to this Interpreter
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
	log.Debugf("[OTEL] [HTTP] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		httpURL, httpURLOk := modules.RetrieveValidSpanMeta(span, "HTTP", "http.url")
		httpMethod, httpMethodOk := modules.RetrieveValidSpanMeta(span, "HTTP", "http.method")

		if httpURLOk && httpMethodOk && len(*httpURL) > 0 {
			var url = sanitizeURL(*httpURL)
			var urn = t.CreateServiceURN(fmt.Sprintf("lambda-http-request/%s/%s", url, *httpMethod))

			modules.SpanBuilder(span, fmt.Sprintf("HTTP %s", *httpMethod), "Http", "http", "consumer", urn, url)
		} else {
			_ = log.Errorf("[OTEL] [LAMBDA.HTTP]: Unable to map the Lambda HTTP request")
			return nil
		}

		modules.InterpretHTTPError(span)
	}

	return spans
}
