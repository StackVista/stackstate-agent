package api

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	v12 "github.com/StackVista/stackstate-agent/pkg/trace/pb/open-telemetry/common/v1"
	openTelemetryTrace "github.com/StackVista/stackstate-agent/pkg/trace/pb/open-telemetry/trace/collector"
	v1 "github.com/StackVista/stackstate-agent/pkg/trace/pb/open-telemetry/trace/v1"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"strconv"
)

/**
Open Telemetry Extra Information:

For Open Telemetry we receive the following groupings within the trace (Will differ based on the libraries used)

Groupings [
	Instrumentation Library (AWS-SDK) = [
		S3 Operation,
		SQS Operation,
		etc.
	]
	Instrumentation Library (AWS-LAMBDA) = [
		Lambda that was Invoked,
	]
	Instrumentation Library (HTTP) = [
		S3 Operation (Call to the AWS service),
		SQS Operation (Call to the AWS service),
		Http Request from library, example Axios,
		Http Request from library, example Http,
		Http Request from any other Instrumentation that calls a service,
	]
]

With the structure above we are unable to determine if the aws-sdk operation succeeded because the http library
contains the failure and success states.

We need to attempt and group things that might contain direct influence over each other, IE we need the following

Groupings [
	Instrumentation Library (AWS-SDK) = [
		S3 Operation {
			Contain meta with the success/failure state mapped from the Instrumentation.Library.Http.S3.Operation
		},
		SQS Operation {
			Contain meta with the success/failure state mapped from the Instrumentation.Library.Http.SQS.Operation
		},
		etc.
	]
	Instrumentation Library (AWS-LAMBDA) = [
		Lambda that was Invoked,
	]
	Instrumentation Library (HTTP) = [
		[REMOVE HERE] S3 Operation (Call to the AWS service),
		[REMOVE HERE] SQS Operation (Call to the AWS service),
		Http Request from library, example Axios,
		Http Request from library, example Http,
		[REMOVE HERE] Http Request from any other Instrumentation that calls a service,
	]
]

We removed the HTTP Instrumentation Library items that was merged so that we do no create components for them.
If this is not removed then we will create http components for these states which is incorrect, needs to be part of the
events
**/

// OpenTelemetrySource Source Identifier for Open Telemetry
const OpenTelemetrySource = "openTelemetry"

// mapOtelTraces Converts the Open Telemetry structure into an accepted sts Traces structure
func mapOpenTelemetryTraces(openTelemetryTraces openTelemetryTrace.ExportTraceServiceRequest) pb.Traces {
	var traces = pb.Traces{}

	for _, resourceSpan := range openTelemetryTraces.ResourceSpans {
		// [Graceful] We can continue without awsAccountID, Unable to map module will give warnings
		awsAccountID := lambdaInstrumentationGetAccountID(resourceSpan)

		// [Graceful] We can continue without determining the http status, This will then allow all the relevant information to still display
		remappedInstrumentationLibrarySpans := determineInstrumentationSuccessFromHTTP(resourceSpan.InstrumentationLibrarySpans)

		for _, instrumentationLibrarySpan := range remappedInstrumentationLibrarySpans {
			// When we reach this point then it is safe to start building a trace
			var singleTrace = pb.Trace{}

			// Loop through the instrumentation's library spans
			for _, instrumentationSpan := range instrumentationLibrarySpan.Spans {
				var meta = &map[string]string{
					"instrumentation_library": instrumentationLibrarySpan.InstrumentationLibrary.Name,
					"source":                  OpenTelemetrySource,
				}

				if awsAccountID != nil {
					(*meta)["aws.account.id"] = *awsAccountID
				}

				openTelemetrySpan := pb.Span{
					Name:     instrumentationSpan.Name,
					Start:    int64(instrumentationSpan.StartTimeUnixNano),
					Duration: int64(instrumentationSpan.EndTimeUnixNano) - int64(instrumentationSpan.StartTimeUnixNano),
					Meta:     *meta,
					Service:  OpenTelemetrySource,
					Resource: OpenTelemetrySource,
					Type:     OpenTelemetrySource,
				}

				// Basic attribute[{}] mapping to a single dict {}
				mapAttributesToMeta(instrumentationSpan.Attributes, meta)

				// [Graceful] We can continue without determining the error state.
				// Attempt to use the information determineInstrumentationSuccessFromHTTP mapped to determine the error state
				mapInstrumentationErrors(&openTelemetrySpan)

				// Attempt to extract the parent, span and trace id from the OTEL span.
				// This does need a string to int conversion thus if anything fails we need to exit
				idExtractError := extractTraceSpanAndParentSpanID(instrumentationSpan, instrumentationLibrarySpan, &openTelemetrySpan)

				if idExtractError != nil {
					log.Errorf("Rejecting instrumentation mapping: %v", idExtractError)
					break
				}

				singleTrace = append(singleTrace, &openTelemetrySpan)
			}

			traces = append(traces, singleTrace)
		}
	}

	return traces
}

// determineInstrumentationSuccessFromHTTP We attempt to separate the http and other instrumentation's from each other
// We then use the http to determine if the other instrumentation calls failed or succeeded by matching up parentSpanIds
// from the http instrumentation and the other instrumentation spanId.
// If the http parentSpanIds does exist in the trace then we remove the http span and merge the attributes we found
// with the relevant parent attributes. This allows the parent to contain the state for if the call failed or succeeded
// whilst we do not create a useless http component by removing it
func determineInstrumentationSuccessFromHTTP(librarySpans []*v1.InstrumentationLibrarySpans) []v1.InstrumentationLibrarySpans {
	var httpInstrumentation []v1.InstrumentationLibrarySpans
	var instrumentation []v1.InstrumentationLibrarySpans

	// We separate the http instrumentation from the other ones
	// This is done because if they do have a parent, they'll be mapped into the parent
	for _, library := range librarySpans {
		switch library.InstrumentationLibrary.Name {
		case "@opentelemetry/instrumentation-http":
			httpInstrumentation = append(httpInstrumentation, *library)
		default:
			instrumentation = append(instrumentation, *library)
		}
	}

	var httpWithParentSpans = make([]*v1.Span, 0)
	var httpLibraryNoParentSpans []v1.InstrumentationLibrarySpans

	// We map through the http instrumentation and separate out the ones that has instrumentation parents
	// and those who do not
	// The ones that does have parents we will be merging them and the ones that does not have to go back into the
	// list of spans to be mapped as c component
	for _, httpLibrary := range httpInstrumentation {
		libraryReference := httpLibrary
		httpWithNoParentSpans := make([]*v1.Span, 0)

		for _, httpSpan := range httpLibrary.Spans {
			// This variable is flagged if any span has this http request as a child
			hasParentSpan := false

			for _, library := range instrumentation {
				for _, span := range library.Spans {
					if httpSpan.ParentSpanId != nil && span.SpanId != nil && string(httpSpan.ParentSpanId) == string(span.SpanId) {
						hasParentSpan = true
					}
				}
			}

			// We can then divide it into a group that is separate
			if hasParentSpan {
				httpWithParentSpans = append(httpWithParentSpans, httpSpan)
			} else {
				httpWithNoParentSpans = append(httpWithNoParentSpans, httpSpan)
			}
		}

		// For the reference that had not parents we need to map them back into a object that can merge back
		libraryReference.Spans = httpWithNoParentSpans
		httpLibraryNoParentSpans = append(httpLibraryNoParentSpans, libraryReference)
	}

	// For the libraries that had http children lets merge the http child with this instrumentation
	for _, library := range instrumentation {
		for _, span := range library.Spans {
			for _, httpSpan := range httpWithParentSpans {
				if httpSpan.ParentSpanId != nil && span.SpanId != nil && string(httpSpan.ParentSpanId) == string(span.SpanId) {
					span.Attributes = append(span.Attributes, httpSpan.Attributes...)
				}
			}
		}
	}

	// Now that we merged the above we can merge the two separate groups back into one
	return append(httpLibraryNoParentSpans, instrumentation...)
}

// lambdaInstrumentationGetAccountID We attempt to extract the aws account id from the instrumentation-aws-lambda
// library this is the root entry for the main lambda calling the script
// This is not a requirement and will only trigger with the aws-lambda library
func lambdaInstrumentationGetAccountID(resourceSpan *v1.ResourceSpans) *string {
	var awsAccountID *string

	// Attempt to extract information from the lambda library to enhance the sdk library
	// We need the account id for sections where it is not defined for example lambda to lambda
	for _, library := range resourceSpan.InstrumentationLibrarySpans {
		if library.InstrumentationLibrary.Name == "@opentelemetry/instrumentation-aws-lambda" {
			for _, span := range library.Spans {
				for _, attribute := range span.Attributes {
					if attribute.Key == "cloud.account.id" {
						var accountID = attribute.Value.GetStringValue()
						awsAccountID = &accountID
					}
				}
			}
		}
	}

	return awsAccountID
}

// extractTraceSpanAndParentSpanID Open telemetry gives us ids that do not correspond to int number but contains string value
// Thus we need to take those and generate a number from it that will always stay the same as long as the seed/string stays the same
// we should receive the same int number
func extractTraceSpanAndParentSpanID(instrumentationSpan *v1.Span, instrumentationLibrarySpan v1.InstrumentationLibrarySpans, openTelemetrySpan *pb.Span) *error {
	if instrumentationSpan.TraceId != nil && instrumentationSpan.TraceId[:] != nil && len(string(instrumentationSpan.TraceId[:])) > 0 {
		traceID, err := convertStringToUint64(string(instrumentationSpan.TraceId[:]))
		if err != nil {
			return &err
		}
		openTelemetrySpan.TraceID = *traceID
	}

	if instrumentationSpan.SpanId != nil && instrumentationSpan.SpanId[:] != nil && len(string(instrumentationSpan.SpanId[:])) > 0 {
		spanID, err := convertStringToUint64(string(instrumentationSpan.SpanId[:]))
		if err != nil {
			return &err
		}
		openTelemetrySpan.SpanID = *spanID
	}

	if instrumentationSpan.ParentSpanId != nil &&
		instrumentationSpan.ParentSpanId[:] != nil &&
		len(string(instrumentationSpan.ParentSpanId[:])) > 0 &&
		instrumentationLibrarySpan.InstrumentationLibrary.Name != "@opentelemetry/instrumentation-aws-lambda" {
		parentSpanID, err := convertStringToUint64(string(instrumentationSpan.ParentSpanId[:]))
		if err != nil {
			return &err
		}
		openTelemetrySpan.ParentID = *parentSpanID
	}

	return nil
}

// mapAttributesToMeta The open telemetry meta attributes' comes in a form of array if (dict type items)
// We can obv combine this to create one dict with a simple key value pair mapping
func mapAttributesToMeta(attributes []*v12.KeyValue, meta *map[string]string) {
	for _, attribute := range attributes {
		attributeValue := attribute.Value.GetValue()

		switch attributeValue.(type) {
		case *v12.AnyValue_StringValue:
			var stringValue = attributeValue.(*v12.AnyValue_StringValue).StringValue
			(*meta)[attribute.Key] = stringValue

		case *v12.AnyValue_BoolValue:
			var boolValue = attributeValue.(*v12.AnyValue_BoolValue).BoolValue
			(*meta)[attribute.Key] = strconv.FormatBool(boolValue)

		case *v12.AnyValue_IntValue:
			var intValue = attributeValue.(*v12.AnyValue_IntValue).IntValue
			(*meta)[attribute.Key] = strconv.FormatInt(intValue, 10)

		case *v12.AnyValue_DoubleValue:
			var doubleValue = attributeValue.(*v12.AnyValue_DoubleValue).DoubleValue
			(*meta)[attribute.Key] = fmt.Sprintf("%f", doubleValue)

		default:
			log.Warnf("Open Telemetry, Unable to map the value '%v' of type '%T' into the meta struct.", attribute, attribute)
		}
	}
}

// mapInstrumentationErrors Determine if the http-instrumentation contains error
// If it does it will be mapped into the http and error states for the span
func mapInstrumentationErrors(span *pb.Span) {
	if statusCode, statusCodeOk := span.Meta["http.status_code"]; statusCodeOk && len(statusCode) > 0 {
		statusCodeInt64, err := strconv.ParseInt(statusCode, 10, 64)
		if err == nil {
			if statusCodeInt64 >= 400 {
				span.Error = int32(statusCodeInt64)
				span.Metrics = map[string]float64{
					"http.status_code": float64(statusCodeInt64),
				}
			}
		}
	}
}

// convertStringToUint64 Current solution for convert Open Telemetry strings to integer id values
// These strings will contain characters, numbers and strings. We need a solid uint at the end
// Do note there is a change that we might get the same id but that will not matter as this is used in tracing and
// Will be temp
func convertStringToUint64(input string) (*uint64, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("unable to convert the string identifier to a uint64 representation: %v", input)
	}

	id := uint64(0)

	// Convert the input to a list of runes that we can add together
	runes := []rune(input)

	// Attempt to create a multiplier using the last and first number
	// This should randomize things a bit more
	multiplier := runes[0] + runes[len(runes)-1]
	for _, r := range runes {
		id += uint64(r) * uint64(multiplier)
	}

	return &id, nil
}
