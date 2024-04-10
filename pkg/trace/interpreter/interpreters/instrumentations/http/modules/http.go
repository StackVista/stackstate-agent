package modules

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/trace/api"
	config "github.com/DataDog/datadog-agent/pkg/trace/interpreter/config"
	interpreter "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters"
	instrumentationbuilders "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentation-builders"
	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/DataDog/datadog-agent/pkg/util/log"
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

// Interpret performs the interpretation for the OpenTelemetryHTTPInterpreter
func (t *OpenTelemetryHTTPInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	log.Debugf("[OTEL] [HTTP] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		httpURL, httpURLOk := instrumentationbuilders.GetSpanMeta("HTTP", span, "http.url")
		httpMethod, httpMethodOk := instrumentationbuilders.GetSpanMeta("HTTP", span, "http.method")

		if httpURLOk && httpMethodOk && len(*httpURL) > 0 {
			var url = *httpURL
			var urn = t.CreateServiceURN(fmt.Sprintf("lambda-http-request/%s/%s", url, *httpMethod))

			instrumentationbuilders.AwsSpanBuilder(span, fmt.Sprintf("%s - %s", *httpMethod, url), "Http", "http", "consumer", urn, url)
		} else {
			_ = log.Errorf("[OTEL] [LAMBDA.HTTP]: Unable to map the Lambda HTTP request")
			return nil
		}

		instrumentationbuilders.InterpretSpanHTTPError(span)
	}

	return spans
}
