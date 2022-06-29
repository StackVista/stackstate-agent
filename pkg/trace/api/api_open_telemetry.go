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
		remappedInstrumentationLibrarySpans := determineInstrumentationStatus(resourceSpan.InstrumentationLibrarySpans)

		for _, instrumentationLibrarySpan := range remappedInstrumentationLibrarySpans {
			// When we reach this point then it is safe to start building a trace
			var singleTrace = pb.Trace{}

			// Loop through the instrumentation's library spans
			for _, instrumentationSpan := range instrumentationLibrarySpan.Spans {
				var meta = map[string]string{
					"instrumentation_library": instrumentationLibrarySpan.InstrumentationLibrary.Name,
					"instrumentation_version": instrumentationLibrarySpan.InstrumentationLibrary.Version,
					"source":                  OpenTelemetrySource,
				}

				if awsAccountID != "" {
					meta["aws.account.id"] = awsAccountID
				}

				openTelemetrySpan := pb.Span{
					Name:     instrumentationSpan.Name,
					Start:    int64(instrumentationSpan.StartTimeUnixNano),
					Duration: int64(instrumentationSpan.EndTimeUnixNano) - int64(instrumentationSpan.StartTimeUnixNano),
					Meta:     meta,
					// We set the Service, Resource, and Type to a default openTelemetry string, This allows us to
					// use the interpreter to identify if this Span is OpenTelemetry and allow us to use internal data
					// like the service type IE sqs sns and redirect it to the correct interpreter for OpenTelemetry
					// If these are not defined then it will never even reach the interpreter.
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

// determineInstrumentationSuccessFromHTTP
// We are faced with the current problem that the function below should solve
// - A librarySpans[] contains instrumentations with the instrumentation identifier
// - The librarySpan[] also contains a span[] which has a potential span type with a http status
// Now the problem is the following:
// If the librarySpan[:].name === "certain http instrumentation name" then check if the librarySpan[:].span[:]. has a parentId
// if it does then it can potentially be a http status child, then next step will be to check if any of the librarySpan range -> spans range
// contains the same spanId === parentSpanId. If it does then this http span should be removed and the attributes merged.
// If not then it should stay as a separate span and not be moved. Yes the http span can have a parent that does not exist if we
// are looking at lambda, so you can not remove it based on parent != nil
func determineInstrumentationStatus(librarySpans []*v1.InstrumentationLibrarySpans) []v1.InstrumentationLibrarySpans {
	// Index for HTTPStatusSpans
	type HTTPStatusSpans struct {
		index int
		span  *v1.Span
	}

	// Instead of creating an entire index of the librarySpans to know if the parent exists that the Span needs to merge with
	// let's take out the span and also take note of the index. The librarySpans index stays untouched meaning that we
	// can insert the http spans back into the same InstrumentationLibrarySpans if it has no parent
	// Thus we reduce the memory usage by not building up a useless dictionary
	// int == index
	httpStatusSpans := make(map[string]HTTPStatusSpans)
	var standAloneLibrarySpans []v1.InstrumentationLibrarySpans

	for libraryIndex, library := range librarySpans {
		// We only have to rebuild the instrumentation span lists for spans we want to remove as components
		// These will be merged in, Thus we only remove instrumentation that contains http statuses and that has a parent
		if library.InstrumentationLibrary.Name == "@opentelemetry/instrumentation-http" {
			// Let's rebuild the http status instrumentation to only contain http requests that contains no parents
			// The ones with parents we want to merge
			var rebuildLibrary = v1.InstrumentationLibrarySpans{
				InstrumentationLibrary: library.InstrumentationLibrary,
				Spans:                  []*v1.Span{},
				SchemaUrl:              library.SchemaUrl,
				XXX_NoUnkeyedLiteral:   library.XXX_NoUnkeyedLiteral,
				XXX_unrecognized:       library.XXX_unrecognized,
				XXX_sizecache:          library.XXX_sizecache,
			}

			// If an HTTP span contains a parent then it will merge with the parent to bring its own health state into the parent
			// If there is no parent then the http component is standalone
			for _, span := range library.Spans {
				if span.ParentSpanId == nil {
					rebuildLibrary.Spans = append(rebuildLibrary.Spans, span)
				} else {
					// The span that we want to merge with the parent can be saved in an index to improve look up times
					// If there is duplicate http status that needs to merge with the same component then there is already
					// something wrong, and we will only use the latest one, the map key will overwrite
					httpStatusSpans[string(span.ParentSpanId)] = HTTPStatusSpans{
						index: libraryIndex,
						span:  span,
					}
				}
			}

			standAloneLibrarySpans = append(standAloneLibrarySpans, rebuildLibrary)
		} else {
			standAloneLibrarySpans = append(standAloneLibrarySpans, *library)
		}
	}

	// We are not removing any librarySpans only span inside the library span
	// Thus if the index of the standAloneLibrarySpans does not match then something was done incorrectly
	// We want to also reuse indexes to inject spans that was unmerged
	if len(standAloneLibrarySpans) != len(librarySpans) {
		_ = fmt.Errorf("attempting to determine instrumentation http status failed, a mismatch occurred leaving the determination invalid. Mapping the trace will continue but the health states will be skipped")

		// Let's not bring back a error and break everything but rather allow the spans to still be created without a health state and log that there is a problem
		var alternativeLibrarySpans []v1.InstrumentationLibrarySpans
		for _, library := range librarySpans {
			alternativeLibrarySpans = append(alternativeLibrarySpans, *library)
		}
		return alternativeLibrarySpans
	}

	// Now lets loop through all the stand-alone library span and determine if there is a http status that can merge
	// with this span to create a complete status span
	// standAloneLibrarySpans will contain all the non http spans that can merge with http spans
	for _, library := range standAloneLibrarySpans {
		for _, span := range library.Spans {
			if httpStatus, ok := httpStatusSpans[string(span.SpanId)]; ok {
				// Append the HTTP status attributes into the span attributes
				span.Attributes = append(span.Attributes, httpStatus.span.Attributes...)

				// We can now delete this key as a parentSpanId will only be used once reducing the map size increase lookup speeds
				// Afterwards we can also determine if there is any left that had no parent
				delete(httpStatusSpans, string(span.SpanId))
			}
		}
	}

	// We can now append back the http status spans that was not used as they had no parents and might be best to act as
	// an individual span
	for _, remainingHTTPStatusSpan := range httpStatusSpans {
		standAloneLibrarySpans[remainingHTTPStatusSpan.index].Spans = append(standAloneLibrarySpans[remainingHTTPStatusSpan.index].Spans, remainingHTTPStatusSpan.span)
	}

	// Return the new merged spans
	return standAloneLibrarySpans
}

// lambdaInstrumentationGetAccountID We attempt to extract the aws account id from the instrumentation-aws-lambda
// library this is the root entry for the main lambda calling the script
// This is not a requirement and will only trigger with the aws-lambda library
func lambdaInstrumentationGetAccountID(resourceSpan *v1.ResourceSpans) string {
	// Attempt to extract information from the lambda library to enhance the sdk library
	// We need the account id for sections where it is not defined for example lambda to lambda
	for _, library := range resourceSpan.InstrumentationLibrarySpans {
		if library.InstrumentationLibrary.Name == "@opentelemetry/instrumentation-aws-lambda" {
			for _, span := range library.Spans {
				for _, attribute := range span.Attributes {
					if attribute.Key == "cloud.account.id" {
						return attribute.Value.GetStringValue()
					}
				}
			}
		}
	}

	return ""
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
func mapAttributesToMeta(attributes []*v12.KeyValue, meta map[string]string) {
	for _, attribute := range attributes {
		attributeValue := attribute.Value.GetValue()

		switch attributeValue.(type) {
		case *v12.AnyValue_StringValue:
			var stringValue = attributeValue.(*v12.AnyValue_StringValue).StringValue
			meta[attribute.Key] = stringValue

		case *v12.AnyValue_BoolValue:
			var boolValue = attributeValue.(*v12.AnyValue_BoolValue).BoolValue
			meta[attribute.Key] = strconv.FormatBool(boolValue)

		case *v12.AnyValue_IntValue:
			var intValue = attributeValue.(*v12.AnyValue_IntValue).IntValue
			meta[attribute.Key] = strconv.FormatInt(intValue, 10)

		case *v12.AnyValue_DoubleValue:
			var doubleValue = attributeValue.(*v12.AnyValue_DoubleValue).DoubleValue
			meta[attribute.Key] = fmt.Sprintf("%f", doubleValue)

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
