package api

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	v11 "github.com/StackVista/stackstate-agent/pkg/trace/pb/open-telemetry/common/v1"
	v1 "github.com/StackVista/stackstate-agent/pkg/trace/pb/open-telemetry/trace/v1"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
		},
	}

	instrumentationLibrarySpans := []*v1.InstrumentationLibrarySpans{
		&instrumentationAwsSdkLibrary,
		&instrumentationOtherLibrary,
		&instrumentationHTTPLibrary,
	}

	newRemappedInstrumentationLibraries := determineInstrumentationSuccessFromHTTP(instrumentationLibrarySpans)

	// Total Counts
	assert.Equal(t, 3, len(instrumentationLibrarySpans), "Total instrumentation library spans count")
	assert.Equal(t, 3, len(newRemappedInstrumentationLibraries), "We should still have the same amount of instrumentationLibraries even if HTTP spans was remapped")

	// Span Lengths
	assert.Equal(t, 1, len(instrumentationLibrarySpans[0].Spans), "[INSTRUMENTATION-AWS-SDK] Total library spans count")
	assert.Equal(t, 1, len(instrumentationLibrarySpans[1].Spans), "[INSTRUMENTATION-*] Total library spans count")
	assert.Equal(t, 2, len(instrumentationLibrarySpans[2].Spans), "[INSTRUMENTATION-HTTP] Total library spans count")

	assert.Equal(t, 1, len(newRemappedInstrumentationLibraries[1].Spans), "[INSTRUMENTATION-AWS-SDK] should have the same amount of spans it started with")
	assert.Equal(t, 1, len(newRemappedInstrumentationLibraries[2].Spans), "[INSTRUMENTATION-*] should have the same amount of spans it started with")
	assert.Equal(t, 1, len(newRemappedInstrumentationLibraries[0].Spans), "[INSTRUMENTATION-HTTP] should have 1 less span, Only one should be mapped and removed because of a parentSpanId. The res")

	// Total Attributes
	assert.Equal(t, 2, len(instrumentationLibrarySpans[0].Spans[0].Attributes), "[INSTRUMENTATION-AWS-SDK] Original amount of attributes should stay the same after mappings")
	assert.Equal(t, 1, len(instrumentationLibrarySpans[1].Spans[0].Attributes), "[INSTRUMENTATION-*] Original amount of attributes should stay the same after mappings")
	assert.Equal(t, 4, len(instrumentationLibrarySpans[2].Spans[0].Attributes), "[INSTRUMENTATION-HTTP] Original amount of attributes should stay the same after mappings")
	assert.Equal(t, 4, len(instrumentationLibrarySpans[2].Spans[1].Attributes), "[INSTRUMENTATION-HTTP] Original amount of attributes should stay the same after mappings")

	assert.Equal(t, 6, len(newRemappedInstrumentationLibraries[1].Spans[0].Attributes), "[INSTRUMENTATION-AWS-SDK] The aws-sdk should have received more attributes since a http instrumentation would merge with this")
	assert.Equal(t, 1, len(newRemappedInstrumentationLibraries[2].Spans[0].Attributes), "[INSTRUMENTATION-*] Should stay the same as there was no http mapping for this")
	assert.Equal(t, 4, len(newRemappedInstrumentationLibraries[0].Spans[0].Attributes), "[INSTRUMENTATION-HTTP] Should stay the same and present as this http instrumentation span had no parentId")
	assert.Equal(t, 1, len(newRemappedInstrumentationLibraries[0].Spans), "[INSTRUMENTATION-HTTP] Should not exist seeing that this http instrumentation span was suppose to merge")
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

	assert.Nil(t, lambdaInstrumentationGetAccountID(&noAccountID), "Should not be able to extract the account id from aws-lambda instrumentation with no id")
	assert.Equal(t, *lambdaInstrumentationGetAccountID(&accountID), "91234567890", "Should be able to extract the account id from aws-lambda instrumentation")
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

	assert.Equal(t, &pb.Span{
		TraceID:  *convertStringToUint64(traceID),
		ParentID: *convertStringToUint64(parentSpanID),
		SpanID:   *convertStringToUint64(spanID),
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
