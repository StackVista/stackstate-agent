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

		// && len(*httpURL) > 0
		if httpURLOk && httpMethodOk {
			// Overwrite span testing
			newSpanTest := map[string]string{}

			// var url = sanitizeURL(*httpURL)
			// var urn = t.CreateServiceURN(fmt.Sprintf("lambda-http-request/%s/%s", url, *httpMethod))

			// Name of component displayed below the icon
			newSpanTest["target"] = *httpURL
			newSpanTest["span.serviceName"] = fmt.Sprintf("%s: %s", "Test", fmt.Sprintf("TEST %s", *httpMethod))

			// Name of the trace displayed on the trace graph line
			span.Name = fmt.Sprintf("%s: %s", "Test", fmt.Sprintf("TEST %s", *httpMethod))

			// Displayed on the trace properties
			span.Resource = fmt.Sprintf("aws.%s", "test")
			span.Type = "aws"

			// Mapping inside StackPack for capturing certain metrics
			newSpanTest["span.serviceType"] = "open-telemetry"
			newSpanTest["source"] = "open-telemetry"

			// Unknown
			span.Service = fmt.Sprintf("aws.%s", "test")
			newSpanTest["service"] = fmt.Sprintf("aws.%s", "test")
			newSpanTest["sts.origin"] = "open-telemetry"

			// General mapping
			newSpanTest["span.kind"] = "producer"
			// span.Meta["span.serviceURN"] = urn
			newSpanTest["sts.service.identifiers"] = "url:::request:test"

			span.Meta = newSpanTest

			span.ParentID = 0
		} else {
			_ = log.Errorf("[OTEL] [LAMBDA.HTTP]: Unable to map the Lambda HTTP request")
			return nil
		}

		modules.InterpretHTTPError(span)
	}

	return spans
}
