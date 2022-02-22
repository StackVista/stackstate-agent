package http

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOpenTelemetryHTTPSpanInterpreter(t *testing.T) {
	interpreter := MakeOpenTelemetryHTTPInterpreter(config.DefaultInterpreterConfig())

	for _, tc := range []struct {
		testCase    string
		interpreter *OpenTelemetryHTTPInterpreter
		trace       []*pb.Span
		expected    []*pb.Span
	}{
		{
			testCase:    "Span should not be filled in if the Open Telemetry data is invalid or missing",
			interpreter: interpreter,
			trace:       []*pb.Span{},
			expected:    []*pb.Span{},
		},
		{
			testCase:    "Open Telemetry data should be mapped if all the correct meta data has been passed",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Meta: map[string]string{
					"http.url":    "http://www.example.com/user/testing?queue=1#random-data",
					"http.method": "GET",
				},
			}},
			expected: []*pb.Span{{
				Name:     "Http: HTTP GET",
				Service:  "aws.http",
				Resource: "aws.http",
				Type:     "aws",
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"http.method":             "GET",
					"http.url":                "http://www.example.com/user/testing?queue=1#random-data",
					"service":                 "aws.http",
					"span.kind":               "consumer",
					"span.serviceName":        "HTTP GET",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/lambda-http-request/www.example.com/user/testing/GET",
					"sts.service.identifiers": "www.example.com/user/testing",
				},
			}},
		},
		{
			testCase:    "Should interpret 4xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "aws.http",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"http.url":    "http://www.example.com/user/testing?queue=1#random-data",
					"http.method": "GET",
				},
			}},
			expected: []*pb.Span{{
				Name:     "Http: HTTP GET",
				Service:  "aws.http",
				Resource: "aws.http",
				Type:     "aws",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 404.0,
				},
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"span.errorClass":         "4xx",
					"http.method":             "GET",
					"http.url":                "http://www.example.com/user/testing?queue=1#random-data",
					"service":                 "aws.http",
					"span.kind":               "consumer",
					"span.serviceName":        "HTTP GET",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/lambda-http-request/www.example.com/user/testing/GET",
					"sts.service.identifiers": "www.example.com/user/testing",
				},
			}},
		},
		{
			testCase:    "Should interpret 5xx http errors",
			interpreter: interpreter,
			trace: []*pb.Span{{
				Service: "aws.http",
				Error:   1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"http.url":    "http://www.example.com/user/testing?queue=1#random-data",
					"http.method": "GET",
				},
			}},
			expected: []*pb.Span{{
				Name:     "Http: HTTP GET",
				Service:  "aws.http",
				Resource: "aws.http",
				Type:     "aws",
				Error:    1,
				Metrics: map[string]float64{
					"http.status_code": 503.0,
				},
				Meta: map[string]string{
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"span.errorClass":         "5xx",
					"http.method":             "GET",
					"http.url":                "http://www.example.com/user/testing?queue=1#random-data",
					"service":                 "aws.http",
					"span.kind":               "consumer",
					"span.serviceName":        "HTTP GET",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/lambda-http-request/www.example.com/user/testing/GET",
					"sts.service.identifiers": "www.example.com/user/testing",
				},
			}},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			actual := tc.interpreter.Interpret(tc.trace)
			assert.EqualValues(t, tc.expected, actual)
		})
	}
}
