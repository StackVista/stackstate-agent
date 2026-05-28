// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build test

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"
)

// TestWithTelemetryWrapper_NoopTracer verifies that the handler works correctly
// when the tracer has not been started (no-op spans, no panics).
func TestWithTelemetryWrapper_NoopTracer(t *testing.T) {
	// Intentionally no mocktracer.Start() — tracer returns NoopSpan.
	th := &TelemetryHandler{
		handlerName: "noopHandler",
		handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	th.handle(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestWithTelemetryWrapper_SpanCreation(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	th := &TelemetryHandler{
		handlerName: "getCheckConfigs",
		handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}

	req := httptest.NewRequest("GET", "/clusterchecks/configs/node1", nil)
	rec := httptest.NewRecorder()
	th.handle(rec, req)

	spans := mt.FinishedSpans()
	require.Len(t, spans, 1)
	span := spans[0]

	assert.Equal(t, "cluster_agent.api.request", span.OperationName())
	assert.Equal(t, "getCheckConfigs", span.Tag("resource.name"))
	assert.Equal(t, "GET", span.Tag("http.method"))
	assert.Equal(t, "/clusterchecks/configs/node1", span.Tag("http.url"))
	assert.Equal(t, 200, span.Tag("http.status_code"))
	assert.Equal(t, false, span.Tag("error"))
}

func TestWithTelemetryWrapper_5xxSetsErrorTag(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	th := &TelemetryHandler{
		handlerName: "errorHandler",
		handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	th.handle(rec, req)

	spans := mt.FinishedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, 500, spans[0].Tag("http.status_code"))
	assert.Equal(t, true, spans[0].Tag("error"))
}

// [sts] Regression test for the "superfluous response.WriteHeader call" warning.
// Reproduces the handler pattern (Write before explicit WriteHeader) that, without
// the Write override on telemetryWriterWrapper, would forward two WriteHeader calls
// to the underlying ResponseWriter and trip Go's stdlib double-write detector.
func TestWithTelemetryWrapper_WriteBeforeWriteHeader_NoDuplicateHeader(t *testing.T) {
	th := &TelemetryHandler{
		handlerName: "writeFirstHandler",
		handler: func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
			w.WriteHeader(http.StatusOK)
		},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	counter := &countingResponseWriter{ResponseWriter: rec}
	th.handle(counter, req)

	assert.Equal(t, 1, counter.headerWrites, "wrapper must forward WriteHeader to the underlying ResponseWriter exactly once")
	assert.Equal(t, "ok", rec.Body.String())
}

// countingResponseWriter counts how many times WriteHeader is forwarded to
// the underlying ResponseWriter. Used to detect duplicate-header bugs that
// httptest.ResponseRecorder alone would not flag.
type countingResponseWriter struct {
	http.ResponseWriter
	headerWrites int
}

func (c *countingResponseWriter) WriteHeader(code int) {
	c.headerWrites++
	c.ResponseWriter.WriteHeader(code)
}

func TestWithTelemetryWrapper_4xxSetsErrorTag(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	th := &TelemetryHandler{
		handlerName: "notFoundHandler",
		handler: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	th.handle(rec, req)

	spans := mt.FinishedSpans()
	require.Len(t, spans, 1)
	assert.Equal(t, 404, spans[0].Tag("http.status_code"))
	assert.Equal(t, true, spans[0].Tag("error"))
}
