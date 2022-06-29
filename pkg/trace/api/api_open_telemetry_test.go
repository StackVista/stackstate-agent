package api

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	v11 "github.com/StackVista/stackstate-agent/pkg/trace/pb/open-telemetry/common/v1"
	openTelemetryTrace "github.com/StackVista/stackstate-agent/pkg/trace/pb/open-telemetry/trace/collector"
	v1 "github.com/StackVista/stackstate-agent/pkg/trace/pb/open-telemetry/trace/v1"
	"github.com/stretchr/testify/assert"
	"math"
	"strconv"
	"testing"
)

func BenchmarkDetermineInstrumentationStatus(b *testing.B) {
	benchmarkTotal := 100

	var instrumentationAwsSdkLibraries []*v1.InstrumentationLibrarySpans
	var instrumentationStackStateLibraries []*v1.InstrumentationLibrarySpans
	var instrumentationHTTPLibraries []*v1.InstrumentationLibrarySpans

	for i := 1; i < benchmarkTotal/3; i++ {
		id := strconv.Itoa(i)

		instrumentationAwsSdkLibraries = append(instrumentationAwsSdkLibraries, &v1.InstrumentationLibrarySpans{
			InstrumentationLibrary: &v11.InstrumentationLibrary{
				Name:    "@opentelemetry/instrumentation-aws-sdk",
				Version: "0.1.0",
			},
			Spans: []*v1.Span{
				{
					TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
					SpanId:            []byte("yjXK+2eLD+s=" + id),
					ParentSpanId:      []byte("Y3OrG+/srMM=" + id),
					Name:              "SQS Name " + id,
					Kind:              4,
					StartTimeUnixNano: uint64(1637684210743088640 + i),
					EndTimeUnixNano:   uint64(1637684210827280128 + i),
					Attributes: []*v11.KeyValue{
						{
							Key: "aws.operation",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "sendMessage",
								},
							},
						},
						{
							Key: "messaging.url",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "https://sqs.eu-west-1.amazonaws.com/120431062118/ENTRY_A_SQS_QUEUE" + id,
								},
							},
						},
					},
				},
			},
		})
	}

	for i := 1; i < benchmarkTotal/3; i++ {
		id := strconv.Itoa(i)

		instrumentationStackStateLibraries = append(instrumentationStackStateLibraries, &v1.InstrumentationLibrarySpans{
			InstrumentationLibrary: &v11.InstrumentationLibrary{
				Name:    "@opentelemetry/instrumentation-stackstate",
				Version: "0.1.0",
			},
			Spans: []*v1.Span{
				{
					TraceId:           []byte("SADAD3423423nasdnsd=="),
					SpanId:            []byte("dsfjkn3m234njkjas=" + id),
					ParentSpanId:      []byte("asdjknkj234bj=" + id),
					Name:              "Custom Span " + id,
					Kind:              4,
					StartTimeUnixNano: uint64(1637684210743088640 + i),
					EndTimeUnixNano:   uint64(1637684210827280128 + i),
					Attributes: []*v11.KeyValue{
						{
							Key: "trace.perspective.name",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "RDS Table: Users" + id,
								},
							},
						},
						{
							Key: "service.name",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "RDS Table" + id,
								},
							},
						},
						{
							Key: "service.type",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "Database Tables" + id,
								},
							},
						},
						{
							Key: "service.identifier",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "rds:database:table:users" + id,
								},
							},
						},
						{
							Key: "resource.name",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "AWS RDS" + id,
								},
							},
						},
					},
				},
			},
		})
	}

	for i := 1; i < benchmarkTotal/3; i++ {
		id := strconv.Itoa(i)

		// Combine every odd number with an existing component
		var parentSpanID []byte = nil
		if i%2 == 0 {
			parentSpanID = []byte("yjXK+2eLD+s=" + id)
		}

		instrumentationHTTPLibraries = append(instrumentationHTTPLibraries, &v1.InstrumentationLibrarySpans{
			InstrumentationLibrary: &v11.InstrumentationLibrary{
				Name:    "@opentelemetry/instrumentation-http",
				Version: "0.1.0",
			},
			Spans: []*v1.Span{
				{
					TraceId:           []byte("SADAD3423423nasdnsd=="),
					SpanId:            []byte("sdajkn4234oinksjdfb=" + id),
					ParentSpanId:      parentSpanID,
					Name:              "HTTP " + id,
					Kind:              4,
					StartTimeUnixNano: uint64(1637684210743088640 + i),
					EndTimeUnixNano:   uint64(1637684210827280128 + i),
					Attributes: []*v11.KeyValue{
						{
							Key: "http.url",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "https://url.com" + id,
								},
							},
						},
						{
							Key: "http.method",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "POST",
								},
							},
						},
						{
							Key: "http.status_code",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_IntValue{
									IntValue: 200,
								},
							},
						},
						{
							Key: "http.status_text",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "OK",
								},
							},
						},
					},
				},
			},
		})
	}

	var instrumentationLibrarySpans []*v1.InstrumentationLibrarySpans

	instrumentationLibrarySpans = append(instrumentationLibrarySpans, instrumentationAwsSdkLibraries...)
	instrumentationLibrarySpans = append(instrumentationLibrarySpans, instrumentationStackStateLibraries...)
	instrumentationLibrarySpans = append(instrumentationLibrarySpans, instrumentationHTTPLibraries...)

	for i := 0; i < 50000; i++ {
		determineInstrumentationStatus(instrumentationLibrarySpans)
	}
}

func TestBulkDetermineInstrumentationStatus(t *testing.T) {
	benchmarkTotal := 1000
	amountPerBlock := int(math.Floor(float64(benchmarkTotal / 3)))

	var instrumentationAwsSdkLibraries []*v1.InstrumentationLibrarySpans
	var instrumentationStackStateLibraries []*v1.InstrumentationLibrarySpans
	var instrumentationHTTPLibraries []*v1.InstrumentationLibrarySpans

	for i := 0; i < amountPerBlock; i++ {
		id := strconv.Itoa(i)

		instrumentationAwsSdkLibraries = append(instrumentationAwsSdkLibraries, &v1.InstrumentationLibrarySpans{
			InstrumentationLibrary: &v11.InstrumentationLibrary{
				Name:    "@opentelemetry/instrumentation-aws-sdk",
				Version: "0.1.0",
			},
			Spans: []*v1.Span{
				{
					TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
					SpanId:            []byte("yjXK+2eLD+s=" + id),
					ParentSpanId:      []byte("Y3OrG+/srMM=" + id),
					Name:              "SQS Name " + id,
					Kind:              4,
					StartTimeUnixNano: uint64(1637684210743088640 + i),
					EndTimeUnixNano:   uint64(1637684210827280128 + i),
					Attributes: []*v11.KeyValue{
						{
							Key: "aws.operation",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "sendMessage",
								},
							},
						},
						{
							Key: "messaging.url",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "https://sqs.eu-west-1.amazonaws.com/120431062118/ENTRY_A_SQS_QUEUE" + id,
								},
							},
						},
					},
				},
				{
					TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
					SpanId:            []byte("asd2332+2eLD+s=" + id),
					ParentSpanId:      nil,
					Name:              "SNS Name " + id,
					Kind:              4,
					StartTimeUnixNano: uint64(1637684210743088644 + i),
					EndTimeUnixNano:   uint64(1637684210827280125 + i),
					Attributes: []*v11.KeyValue{
						{
							Key: "aws.operation",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "topicTest",
								},
							},
						},
					},
				},
			},
		})
	}

	for i := 0; i < amountPerBlock; i++ {
		id := strconv.Itoa(i)

		instrumentationStackStateLibraries = append(instrumentationStackStateLibraries, &v1.InstrumentationLibrarySpans{
			InstrumentationLibrary: &v11.InstrumentationLibrary{
				Name:    "@opentelemetry/instrumentation-stackstate",
				Version: "0.1.0",
			},
			Spans: []*v1.Span{
				{
					TraceId:           []byte("SADAD3423423nasdnsd=="),
					SpanId:            []byte("dsfjkn3m234njkjas=" + id),
					ParentSpanId:      []byte("asdjknkj234bj=" + id),
					Name:              "Custom Span " + id,
					Kind:              4,
					StartTimeUnixNano: uint64(1637684210743088640 + i),
					EndTimeUnixNano:   uint64(1637684210827280128 + i),
					Attributes: []*v11.KeyValue{
						{
							Key: "trace.perspective.name",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "RDS Table: Users" + id,
								},
							},
						},
						{
							Key: "service.name",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "RDS Table" + id,
								},
							},
						},
						{
							Key: "service.type",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "Database Tables" + id,
								},
							},
						},
						{
							Key: "service.identifier",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "rds:database:table:users" + id,
								},
							},
						},
						{
							Key: "resource.name",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "AWS RDS" + id,
								},
							},
						},
					},
				},
			},
		})
	}

	mergedHTTPSpanName := "HTTP Merged"
	unmergedHTTPSpanName := "HTTP"

	for i := 0; i < amountPerBlock; i++ {
		id := strconv.Itoa(i)

		// Combine every odd number with an existing component
		var parentSpanID []byte = nil
		name := unmergedHTTPSpanName
		if i%2 == 0 {
			parentSpanID = []byte("yjXK+2eLD+s=" + id)
			name = mergedHTTPSpanName
		}

		instrumentationHTTPLibraries = append(instrumentationHTTPLibraries, &v1.InstrumentationLibrarySpans{
			InstrumentationLibrary: &v11.InstrumentationLibrary{
				Name:    "@opentelemetry/instrumentation-http",
				Version: "0.1.0",
			},
			Spans: []*v1.Span{
				{
					TraceId:           []byte("SADAD3423423nasdnsd=="),
					SpanId:            []byte("sdajkn4234oinksjdfb=" + id),
					ParentSpanId:      parentSpanID,
					Name:              name,
					Kind:              4,
					StartTimeUnixNano: uint64(1637684210743088640 + i),
					EndTimeUnixNano:   uint64(1637684210827280128 + i),
					Attributes: []*v11.KeyValue{
						{
							Key: "http.url",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "https://url.com" + id,
								},
							},
						},
						{
							Key: "http.method",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "POST",
								},
							},
						},
						{
							Key: "http.status_code",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_IntValue{
									IntValue: 200,
								},
							},
						},
						{
							Key: "http.status_text",
							Value: &v11.AnyValue{
								Value: &v11.AnyValue_StringValue{
									StringValue: "OK",
								},
							},
						},
					},
				},
			},
		})
	}

	var instrumentationLibrarySpans []*v1.InstrumentationLibrarySpans

	instrumentationLibrarySpans = append(instrumentationLibrarySpans, instrumentationAwsSdkLibraries...)
	instrumentationLibrarySpans = append(instrumentationLibrarySpans, instrumentationStackStateLibraries...)
	instrumentationLibrarySpans = append(instrumentationLibrarySpans, instrumentationHTTPLibraries...)

	totalSpansBeforeInstrumentationStatus := 0
	for _, instrumentation := range instrumentationLibrarySpans {
		totalSpansBeforeInstrumentationStatus += len(instrumentation.Spans)
	}

	afterInstrumentationStatus := determineInstrumentationStatus(instrumentationLibrarySpans)

	totalSpansAfterInstrumentationStatus := 0
	for _, instrumentation := range afterInstrumentationStatus {
		totalSpansAfterInstrumentationStatus += len(instrumentation.Spans)
	}

	unmergedHTTPSpans := 0
	mergedHTTPSpans := 0
	for _, instrumentation := range afterInstrumentationStatus {
		for _, span := range instrumentation.Spans {
			if span.Name == unmergedHTTPSpanName {
				unmergedHTTPSpans++
			} else if span.Name == mergedHTTPSpanName {
				mergedHTTPSpans++
			}
		}
	}

	assert.Equal(t, amountPerBlock*3, len(instrumentationLibrarySpans), "The total instrumentation's before determineInstrumentationStatus should match the total we started with")
	assert.Equal(t, amountPerBlock*3, len(afterInstrumentationStatus), "The total instrumentation's after determineInstrumentationStatus should match the total we started with")
	assert.Equal(t, amountPerBlock*4, totalSpansBeforeInstrumentationStatus, "The total instrumentation spans before determineInstrumentationStatus")
	// We will deduct 50 of amountPerBlock as only the odd numbers will merge with existing components with the above spans
	// We have a one offset atm
	assert.Equal(t, (amountPerBlock*4)-(amountPerBlock/2)-1, totalSpansAfterInstrumentationStatus, "The total instrumentation spans after determineInstrumentationStatus")
	// We need to make sure that none of the spans that should have merged stayed behind, if they did then the merger failed
	assert.Equal(t, mergedHTTPSpans, 0, "All the spans merged successfully, Testing merged components")
	// Next we make sure that all the alternative http spans did actually not merge
	assert.Equal(t, unmergedHTTPSpans, amountPerBlock/2, "All the spans merged successfully, Testing unmerged components")

}

func TestMapOpenTelemetryTraces(t *testing.T) {
	instrumentationAwsSdkLibrary := v1.InstrumentationLibrarySpans{
		InstrumentationLibrary: &v11.InstrumentationLibrary{
			Name:    "@opentelemetry/instrumentation-aws-sdk",
			Version: "0.1.0",
		},
		Spans: []*v1.Span{
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("12389ybsad32"),
				ParentSpanId:      []byte("234/dsfs234=sd"),
				Name:              "SQS Success",
				Kind:              4,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "aws.operation",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "sendMessage",
							},
						},
					},
					{
						Key: "messaging.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "https://sqs.eu-west-1.amazonaws.com/120431062118/ENTRY_A_SQS_QUEUE",
							},
						},
					},
				},
			},
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("sadkjnas832434"),
				ParentSpanId:      []byte("234/dsfs234=sd"),
				Name:              "SQS Failure",
				Kind:              4,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "aws.operation",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "sendMessage",
							},
						},
					},
					{
						Key: "messaging.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "https://sqs.eu-west-1.amazonaws.com/120431062118/RANDOM",
							},
						},
					},
				},
			},
		},
	}

	instrumentationOtherLibrary := v1.InstrumentationLibrarySpans{
		InstrumentationLibrary: &v11.InstrumentationLibrary{
			Name:    "@opentelemetry/other-library",
			Version: "0.1.0",
		},
		Spans: []*v1.Span{
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("asdasd8324298"),
				ParentSpanId:      []byte("234/dsfs234=sd"),
				Name:              "Other Name",
				Kind:              4,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "random.value",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "text",
							},
						},
					},
				},
			},
		},
	}

	instrumentationHTTPLibrary := v1.InstrumentationLibrarySpans{
		InstrumentationLibrary: &v11.InstrumentationLibrary{
			Name:    "@opentelemetry/instrumentation-http",
			Version: "0.1.0",
		},
		Spans: []*v1.Span{
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("3423hbiusdf9a"),
				ParentSpanId:      []byte("12389ybsad32"),
				Name:              "HTTPS A",
				Kind:              3,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "http.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "https://otel-example-nodejs-dev-s3-965323806078-eu-west-1.s3.eu-west-1.amazonaws.com/filename",
							},
						},
					},
					{
						Key: "http.method",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "PUT",
							},
						},
					},
					{
						Key: "http.status_code",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_IntValue{
								IntValue: 200,
							},
						},
					},
					{
						Key: "http.status_text",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "OK",
							},
						},
					},
				},
			},
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("asd234213sd"),
				ParentSpanId:      []byte("sadkjnas832434"),
				Name:              "HTTPS B",
				Kind:              3,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "http.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "https://otel-example-nodejs-dev-s3-965323806078-eu-west-1.s3.eu-west-1.amazonaws.com/filename",
							},
						},
					},
					{
						Key: "http.method",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "POST",
							},
						},
					},
					{
						Key: "http.status_code",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_IntValue{
								IntValue: 404,
							},
						},
					},
					{
						Key: "http.status_text",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "NOT FOUND",
							},
						},
					},
				},
			},
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("asdkuh2349hbdasd"),
				ParentSpanId:      []byte("234/dsfs234=sd"),
				Name:              "HTTPS C",
				Kind:              3,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "http.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "https://random/filename",
							},
						},
					},
					{
						Key: "http.method",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "POST",
							},
						},
					},
					{
						Key: "http.status_code",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_IntValue{
								IntValue: 404,
							},
						},
					},
					{
						Key: "http.status_text",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "Not Found - This has no parent span",
							},
						},
					},
				},
			},
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("zxcasdasdn3nk3mk32"),
				ParentSpanId:      nil,
				Name:              "HTTPS D",
				Kind:              3,
				StartTimeUnixNano: 1637684210743088641,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "http.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "random.url",
							},
						},
					},
					{
						Key: "http.method",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "POST",
							},
						},
					},
					{
						Key: "http.status_code",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_IntValue{
								IntValue: 404,
							},
						},
					},
					{
						Key: "http.status_text",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "This has a nil parent",
							},
						},
					},
				},
			},
		},
	}

	traces := mapOpenTelemetryTraces(openTelemetryTrace.ExportTraceServiceRequest{
		ResourceSpans: []*v1.ResourceSpans{
			{
				InstrumentationLibrarySpans: []*v1.InstrumentationLibrarySpans{
					&instrumentationAwsSdkLibrary,
					&instrumentationOtherLibrary,
					&instrumentationHTTPLibrary,
				},
			},
		},
	})

	expected := pb.Traces{
		pb.Trace{
			&pb.Span{
				Service:  OpenTelemetrySource,
				Name:     "SQS Success",
				Resource: OpenTelemetrySource,
				TraceID:  280050,
				SpanID:   88605,
				ParentID: 159150,
				Start:    1637684210743088640,
				Duration: 84191488,
				Meta: map[string]string{
					"aws.operation":           "sendMessage",
					"http.method":             "PUT",
					"instrumentation_version": "0.1.0",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/120431062118/ENTRY_A_SQS_QUEUE",
					"source":                  OpenTelemetrySource,
					"http.status_code":        "200",
					"http.status_text":        "OK",
					"http.url":                "https://otel-example-nodejs-dev-s3-965323806078-eu-west-1.s3.eu-west-1.amazonaws.com/filename",
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
				},
				Metrics: nil,
				Type:    OpenTelemetrySource,
			},
			&pb.Span{
				Service:  OpenTelemetrySource,
				Name:     "SQS Failure",
				Resource: OpenTelemetrySource,
				TraceID:  280050,
				SpanID:   193553,
				ParentID: 159150,
				Start:    1637684210743088640,
				Duration: 84191488,
				Error:    404,
				Meta: map[string]string{
					"aws.operation":           "sendMessage",
					"http.method":             "POST",
					"instrumentation_version": "0.1.0",
					"source":                  OpenTelemetrySource,
					"http.status_code":        "404",
					"http.status_text":        "NOT FOUND",
					"http.url":                "https://otel-example-nodejs-dev-s3-965323806078-eu-west-1.s3.eu-west-1.amazonaws.com/filename",
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/120431062118/RANDOM",
				},
				Metrics: map[string]float64{
					"http.status_code": 404,
				},
				Type: OpenTelemetrySource,
			},
		},
		pb.Trace{
			&pb.Span{
				Service:  OpenTelemetrySource,
				Name:     "Other Name",
				Resource: OpenTelemetrySource,
				TraceID:  280050,
				SpanID:   152388,
				ParentID: 159150,
				Start:    1637684210743088640,
				Duration: 84191488,
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/other-library",
					"instrumentation_version": "0.1.0",
					"random.value":            "text",
					"source":                  OpenTelemetrySource,
				},
				Metrics: nil,
				Type:    OpenTelemetrySource,
			},
		},
		{
			&pb.Span{
				Service:  OpenTelemetrySource,
				Name:     "HTTPS D",
				Resource: OpenTelemetrySource,
				TraceID:  280050,
				SpanID:   294292,
				ParentID: 0,
				Start:    1637684210743088641,
				Duration: 84191487,
				Error:    404,
				Meta: map[string]string{
					"http.method":             "POST",
					"http.status_code":        "404",
					"http.status_text":        "This has a nil parent",
					"http.url":                "random.url",
					"instrumentation_library": "@opentelemetry/instrumentation-http",
					"instrumentation_version": "0.1.0",
					"source":                  OpenTelemetrySource,
				},
				Metrics: map[string]float64{
					"http.status_code": 404,
				},
				Type: OpenTelemetrySource,
			},
			&pb.Span{
				Service:  OpenTelemetrySource,
				Name:     "HTTPS C",
				Resource: OpenTelemetrySource,
				TraceID:  280050,
				SpanID:   288408,
				ParentID: 159150,
				Start:    1637684210743088640,
				Duration: 84191488,
				Error:    404,
				Meta: map[string]string{
					"http.method":             "POST",
					"http.status_code":        "404",
					"http.status_text":        "Not Found - This has no parent span",
					"http.url":                "https://random/filename",
					"instrumentation_library": "@opentelemetry/instrumentation-http",
					"instrumentation_version": "0.1.0",
					"source":                  OpenTelemetrySource,
				},
				Metrics: map[string]float64{
					"http.status_code": 404,
				},
				Type: OpenTelemetrySource,
			},
		},
	}

	assert.Equal(t, &expected, &traces, "Total instrumentation library spans count")
}

func TestRemapOtelHttpLibraryStatusMappers(t *testing.T) {
	instrumentationAwsSdkLibrary := v1.InstrumentationLibrarySpans{
		InstrumentationLibrary: &v11.InstrumentationLibrary{
			Name:    "@opentelemetry/instrumentation-aws-sdk",
			Version: "0.1.0",
		},
		Spans: []*v1.Span{
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("yjXK+2eLD+s="),
				ParentSpanId:      []byte("Y3OrG+/srMM="),
				Name:              "SQS",
				Kind:              4,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "aws.operation",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "sendMessage",
							},
						},
					},
					{
						Key: "messaging.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "https://sqs.eu-west-1.amazonaws.com/120431062118/ENTRY_A_SQS_QUEUE",
							},
						},
					},
				},
			},
		},
	}

	instrumentationOtherLibrary := v1.InstrumentationLibrarySpans{
		InstrumentationLibrary: &v11.InstrumentationLibrary{
			Name:    "@opentelemetry/other-library",
			Version: "0.1.0",
		},
		Spans: []*v1.Span{
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("sdf334%dasd"),
				ParentSpanId:      []byte("Y3OrG+/srMM="),
				Name:              "SQS",
				Kind:              4,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "random.value",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "text",
							},
						},
					},
				},
			},
		},
	}

	instrumentationHTTPLibrary := v1.InstrumentationLibrarySpans{
		InstrumentationLibrary: &v11.InstrumentationLibrary{
			Name:    "@opentelemetry/instrumentation-http",
			Version: "0.1.0",
		},
		Spans: []*v1.Span{
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("edsf+2eLD+s="),
				ParentSpanId:      []byte("yjXK+2eLD+s="),
				Name:              "HTTPS PUT",
				Kind:              3,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "http.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "https://otel-example-nodejs-dev-s3-965323806078-eu-west-1.s3.eu-west-1.amazonaws.com/filename",
							},
						},
					},
					{
						Key: "http.method",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "PUT",
							},
						},
					},
					{
						Key: "http.status_code",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_IntValue{
								IntValue: 200,
							},
						},
					},
					{
						Key: "http.status_text",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "OK",
							},
						},
					},
				},
			},
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("zxcxc+2eLD+s="),
				ParentSpanId:      []byte("this-has-no-parent"),
				Name:              "HTTPS PUT",
				Kind:              3,
				StartTimeUnixNano: 1637684210743088640,
				EndTimeUnixNano:   1637684210827280128,
				Attributes: []*v11.KeyValue{
					{
						Key: "http.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "https://random/filename",
							},
						},
					},
					{
						Key: "http.method",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "POST",
							},
						},
					},
					{
						Key: "http.status_code",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_IntValue{
								IntValue: 404,
							},
						},
					},
					{
						Key: "http.status_text",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "Not Found",
							},
						},
					},
				},
			},
			{
				TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
				SpanId:            []byte("ex2xe+1eLD+s="),
				ParentSpanId:      nil,
				Name:              "HTTPS POST - Nil parent",
				Kind:              3,
				StartTimeUnixNano: 1637684210743088643,
				EndTimeUnixNano:   1637684210827280122,
				Attributes: []*v11.KeyValue{
					{
						Key: "http.url",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "https://www.unknown-loc.com",
							},
						},
					},
					{
						Key: "http.method",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "POST",
							},
						},
					},
					{
						Key: "http.status_code",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_IntValue{
								IntValue: 200,
							},
						},
					},
					{
						Key: "http.status_text",
						Value: &v11.AnyValue{
							Value: &v11.AnyValue_StringValue{
								StringValue: "OK",
							},
						},
					},
				},
			},
		},
	}

	instrumentationLibrarySpans := []*v1.InstrumentationLibrarySpans{
		&instrumentationAwsSdkLibrary,
		&instrumentationOtherLibrary,
		&instrumentationHTTPLibrary,
	}

	newRemappedInstrumentationLibraries := determineInstrumentationStatus(instrumentationLibrarySpans)

	assert.Equal(t, 3, len(newRemappedInstrumentationLibraries), "We should still have the same amount of instrumentationLibraries even if HTTP spans was remapped")
	assert.Equal(t, 1, len(newRemappedInstrumentationLibraries[0].Spans), "[INSTRUMENTATION-HTTP] should have 1 less span, Only one should be mapped and removed because of a parentSpanId. The res")
	assert.Equal(t, 1, len(newRemappedInstrumentationLibraries[1].Spans), "[INSTRUMENTATION-AWS-SDK] should have the same amount of spans it started with")
	assert.Equal(t, 2, len(newRemappedInstrumentationLibraries[2].Spans), "[INSTRUMENTATION-*] should have the same amount of spans it started with")
	assert.Equal(t, 6, len(newRemappedInstrumentationLibraries[0].Spans[0].Attributes), "[INSTRUMENTATION-AWS-SDK] The aws-sdk should have received more attributes since a http instrumentation would merge with this")
	assert.Equal(t, 1, len(newRemappedInstrumentationLibraries[1].Spans[0].Attributes), "[INSTRUMENTATION-*] Should stay the same as there was no http mapping for this")
	assert.Equal(t, 4, len(newRemappedInstrumentationLibraries[2].Spans[0].Attributes), "[INSTRUMENTATION-HTTP] Should stay the same and present as this http instrumentation span had no parentId")
	assert.Equal(t, 4, len(newRemappedInstrumentationLibraries[2].Spans[1].Attributes), "[INSTRUMENTATION-HTTP] Should stay the same and present as this http instrumentation span had no parentId")
}

func TestLambdaOtelInstrumentationGetAccountID(t *testing.T) {
	noAccountID := v1.ResourceSpans{
		InstrumentationLibrarySpans: []*v1.InstrumentationLibrarySpans{
			{
				InstrumentationLibrary: &v11.InstrumentationLibrary{
					Name:    "@opentelemetry/instrumentation-aws-lambda",
					Version: "0.27.0",
				},
				Spans: []*v1.Span{
					{
						TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
						SpanId:            []byte("Y3OrG+/srMM="),
						ParentSpanId:      []byte("RK3KTmkP93g="),
						Name:              "nn-observability-stack-dev-EntryLambdaToSQS",
						Kind:              2,
						StartTimeUnixNano: 1637684210732307968,
						EndTimeUnixNano:   1637684210827808768,
						Attributes: []*v11.KeyValue{
							{
								Key: "faas.execution",
								Value: &v11.AnyValue{
									Value: &v11.AnyValue_StringValue{
										StringValue: "2ef7e384-cda2-46cc-bcf7-2268671e2cf5",
									},
								},
							},
							{
								Key: "faas.id",
								Value: &v11.AnyValue{
									Value: &v11.AnyValue_StringValue{
										StringValue: "arn:aws:lambda:eu-west-1:120431062118:function:nn-observability-stack-dev-EntryLambdaToSQS",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	accountID := v1.ResourceSpans{
		InstrumentationLibrarySpans: []*v1.InstrumentationLibrarySpans{
			{
				InstrumentationLibrary: &v11.InstrumentationLibrary{
					Name:    "@opentelemetry/instrumentation-aws-lambda",
					Version: "0.27.0",
				},
				Spans: []*v1.Span{
					{
						TraceId:           []byte("YZ0T8B2Ll8IIzMv3EfFIqQ=="),
						SpanId:            []byte("Y3OrG+/srMM="),
						ParentSpanId:      []byte("RK3KTmkP93g="),
						Name:              "nn-observability-stack-dev-EntryLambdaToSQS",
						Kind:              2,
						StartTimeUnixNano: 1637684210732307968,
						EndTimeUnixNano:   1637684210827808768,
						Attributes: []*v11.KeyValue{
							{
								Key: "faas.execution",
								Value: &v11.AnyValue{
									Value: &v11.AnyValue_StringValue{
										StringValue: "2ef7e384-cda2-46cc-bcf7-2268671e2cf5",
									},
								},
							},
							{
								Key: "faas.id",
								Value: &v11.AnyValue{
									Value: &v11.AnyValue_StringValue{
										StringValue: "arn:aws:lambda:eu-west-1:120431062118:function:nn-observability-stack-dev-EntryLambdaToSQS",
									},
								},
							},
							{
								Key: "cloud.account.id",
								Value: &v11.AnyValue{
									Value: &v11.AnyValue_StringValue{
										StringValue: "91234567890",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	assert.Equal(t, lambdaInstrumentationGetAccountID(&noAccountID), "", "A blank string will be extract for the account id from aws-lambda instrumentation with no id")
	assert.Equal(t, lambdaInstrumentationGetAccountID(&accountID), "91234567890", "Should be able to extract the account id from aws-lambda instrumentation")
}

func TestConvertOtelIdentifiersToStsIdentifiers(t *testing.T) {
	traceID := "YZ0T8B2Ll8IIzMv3EfFIqQ=="
	spanID := "Y3OrG+/srMM="
	parentSpanID := "RK3KTmkP93g="

	resourceSpan := v1.ResourceSpans{
		InstrumentationLibrarySpans: []*v1.InstrumentationLibrarySpans{
			{
				InstrumentationLibrary: &v11.InstrumentationLibrary{
					Name:    "@opentelemetry/instrumentation-aws-sdk",
					Version: "0.27.0",
				},
				Spans: []*v1.Span{
					{
						TraceId:           []byte(traceID),
						SpanId:            []byte(spanID),
						ParentSpanId:      []byte(parentSpanID),
						Name:              "nn-observability-stack-dev-EntryLambdaToSQS",
						Kind:              2,
						StartTimeUnixNano: 1637684210732307968,
						EndTimeUnixNano:   1637684210827808768,
						Attributes: []*v11.KeyValue{
							{
								Key: "faas.execution",
								Value: &v11.AnyValue{
									Value: &v11.AnyValue_StringValue{
										StringValue: "2ef7e384-cda2-46cc-bcf7-2268671e2cf5",
									},
								},
							},
							{
								Key: "faas.id",
								Value: &v11.AnyValue{
									Value: &v11.AnyValue_StringValue{
										StringValue: "arn:aws:lambda:eu-west-1:120431062118:function:nn-observability-stack-dev-EntryLambdaToSQS",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	captureSpan := pb.Span{}
	selectedSpan := resourceSpan.InstrumentationLibrarySpans[0].Spans[0]

	extractTraceSpanAndParentSpanID(selectedSpan, *resourceSpan.InstrumentationLibrarySpans[0], &captureSpan)

	traceIDIntValue, _ := convertStringToUint64(traceID)
	parentSpanIDIntValue, _ := convertStringToUint64(parentSpanID)
	spanIDIntValue, _ := convertStringToUint64(spanID)

	assert.Equal(t, &pb.Span{
		TraceID:  *traceIDIntValue,
		ParentID: *parentSpanIDIntValue,
		SpanID:   *spanIDIntValue,
	}, &captureSpan, "Extract ids from Open Telemetry span, convert to a uint64 and push into the main span")
}

func TestMapInstrumentationErrors(t *testing.T) {
	span200 := pb.Span{
		Service:  "span-service",
		Name:     "span-name",
		Resource: "span-resource",
		TraceID:  1000,
		SpanID:   2000,
		ParentID: 3000,
		Start:    400000,
		Duration: 500,
		Meta: map[string]string{
			"http.status_code": "200",
		},
		Metrics: map[string]float64{},
		Type:    "span-type",
	}

	span404 := pb.Span{
		Service:  "span-service",
		Name:     "span-name",
		Resource: "span-resource",
		TraceID:  1000,
		SpanID:   2000,
		ParentID: 3000,
		Start:    400000,
		Duration: 500,
		Meta: map[string]string{
			"http.status_code": "404",
			"http.status_text": "NOT FOUND",
		},
		Metrics: map[string]float64{},
		Type:    "span-type",
	}

	span500 := pb.Span{
		Service:  "span-service",
		Name:     "span-name",
		Resource: "span-resource",
		TraceID:  1000,
		SpanID:   2000,
		ParentID: 3000,
		Start:    400000,
		Duration: 500,
		Meta: map[string]string{
			"http.status_code": "500",
			"http.status_text": "ERROR",
		},
		Metrics: map[string]float64{},
		Type:    "span-type",
	}

	mapInstrumentationErrors(&span200)
	assert.Equal(t, int32(0), span200.Error, "[200] The status 200 does not require mapping")
	assert.Equal(t, "200", span200.Meta["http.status_code"], "[200] The status code still needs to be in the Meta dict")

	mapInstrumentationErrors(&span404)
	assert.Equal(t, int32(404), span404.Error, "[404] The status code meta needs to be mapped into the correct top level error value")
	assert.Equal(t, "404", span404.Meta["http.status_code"], "[404] The status code still needs to be in the Meta dict")
	assert.Equal(t, "NOT FOUND", span404.Meta["http.status_text"], "[404] The status text still needs to be in the Meta dict")

	mapInstrumentationErrors(&span500)
	assert.Equal(t, int32(500), span500.Error, "[500] The status code meta needs to be mapped into the correct top level error value")
	assert.Equal(t, "500", span500.Meta["http.status_code"], "[500] The status code still needs to be in the Meta dict")
	assert.Equal(t, "ERROR", span500.Meta["http.status_text"], "[500] The status text still needs to be in the Meta dict")
}

func TestConvertStringToUint64(t *testing.T) {
	sampleA, _ := convertStringToUint64("random-text-a")
	assert.Equal(t, uint64(270291), *sampleA, "String to Int sample a should always have the same value")

	sampleB, _ := convertStringToUint64("random-text-b")
	assert.Equal(t, uint64(271784), *sampleB, "String to Int sample b should always have the same value")

	sampleC, _ := convertStringToUint64("hello-world")
	assert.Equal(t, uint64(230316), *sampleC, "String to Int sample c should always have the same value")

	sampleD, _ := convertStringToUint64("012393498324789")
	assert.Equal(t, uint64(83160), *sampleD, "String to Int sample d should always have the same value")

	sampleE, _ := convertStringToUint64("ZXsadkjnjkASDsad")
	assert.Equal(t, uint64(295260), *sampleE, "String to Int sample e should always have the same value")

	sampleF, _ := convertStringToUint64("123SADbhsad372343")
	assert.Equal(t, uint64(119000), *sampleF, "String to Int sample f should always have the same value")

	sampleG, _ := convertStringToUint64("$%^348y2349987fjskdfh")
	assert.Equal(t, uint64(218540), *sampleG, "String to Int sample g should always have the same value")

	sampleH, _ := convertStringToUint64("random-text")
	assert.Equal(t, uint64(261970), *sampleH, "String to Int sample h should always have the same value")

	sampleI, _ := convertStringToUint64("ASDxkjchi8y349h234987hgfeiwundfuishf89234yh23uh4iu2rh8hsad")
	assert.Equal(t, uint64(833580), *sampleI, "String to Int sample i should always have the same value")
}
