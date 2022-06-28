package interpreter

/*

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpanInterpreterEngine(t *testing.T) {
	sie := NewSpanInterpreterEngine(config.New())

	for _, tc := range []struct {
		testCase string
		span     pb.Span
		expected pb.Span
	}{
		{
			testCase: "Should run the default span interpreter if we have no metadata on the span",
			span:     pb.Span{Service: "SpanServiceName"},
			expected: pb.Span{
				Service: "SpanServiceName",
				Meta: map[string]string{
					"span.serviceName": "SpanServiceName",
					"span.serviceURN":  "urn:service:/SpanServiceName",
				},
			},
		},
		{
			testCase: "Should run the sql span interpreter if we have metadata and the type is 'sql'",
			span: pb.Span{
				Service: "Postgresql",
				Type:    "sql",
				Meta: map[string]string{
					"span.starttime": "1586441095", //Thursday, 9 April 2020 14:04:55
					"span.hostname":  "hostname",
					"span.pid":       "10",
					"span.kind":      "some-kind",
					"db.type":        "postgresql",
					"db.instance":    "Instance",
				},
			},
			expected: pb.Span{
				Service: "Postgresql",
				Type:    "sql",
				Meta: map[string]string{
					"span.serviceName": "Postgresql:Instance",
					"span.starttime":   "1586441095", //Thursday, 9 April 2020 14:04:55
					"span.hostname":    "hostname",
					"span.pid":         "10",
					"span.kind":        "some-kind",
					"db.type":          "postgresql",
					"db.instance":      "Instance",
					"span.serviceType": "postgresql",
					"span.serviceURN":  "urn:service:/Postgresql:Instance",
				},
			},
		},
		{
			testCase: "Should run the process span interpreter if we have metadata and the type is 'web'",
			span: pb.Span{
				Service: "WebServiceName",
				Type:    "web",
				Meta: map[string]string{
					"span.starttime": "1586441095", //Thursday, 9 April 2020 14:04:55
					"span.hostname":  "hostname",
					"span.pid":       "10",
					"span.kind":      "some-kind",
				},
			},
			expected: pb.Span{
				Service: "WebServiceName",
				Type:    "web",
				Meta: map[string]string{
					"span.serviceName":        "WebServiceName",
					"span.starttime":          "1586441095", //Thursday, 9 April 2020 14:04:55
					"span.hostname":           "hostname",
					"span.pid":                "10",
					"span.kind":               "some-kind",
					"span.serviceType":        "service",
					"span.serviceURN":         "urn:service:/WebServiceName",
					"span.serviceInstanceURN": "urn:service-instance:/WebServiceName:/hostname:10:1586441095",
				},
			},
		},
		{
			testCase: "Should run the process span interpreter if we have metadata and the type is 'server'",
			span: pb.Span{
				Service: "JavaServiceName",
				Type:    "server",
				Meta: map[string]string{
					"span.starttime": "1586441095", //Thursday, 9 April 2020 14:04:55
					"span.hostname":  "hostname",
					"span.pid":       "10",
					"span.kind":      "some-kind",
					"language":       "jvm",
				},
			},
			expected: pb.Span{
				Service: "JavaServiceName",
				Type:    "server",
				Meta: map[string]string{
					"span.serviceName":        "JavaServiceName",
					"span.starttime":          "1586441095", //Thursday, 9 April 2020 14:04:55
					"span.hostname":           "hostname",
					"span.pid":                "10",
					"span.kind":               "some-kind",
					"language":                "jvm",
					"span.serviceType":        "java",
					"span.serviceURN":         "urn:service:/JavaServiceName",
					"span.serviceInstanceURN": "urn:service-instance:/JavaServiceName:/hostname:10:1586441095",
				},
			},
		},
		{
			testCase: "Should run the traefik span interpreter if the meta source is 'traefik'",
			span: pb.Span{
				Service: "TraefikServiceName",
				Meta: map[string]string{
					"source":    "traefik",
					"http.host": "hostname",
					"span.kind": "server",
				},
			},
			expected: pb.Span{
				Service: "TraefikServiceName",
				Meta: map[string]string{
					"span.serviceName": "hostname",
					"source":           "traefik",
					"http.host":        "hostname",
					"span.kind":        "server",
					"span.serviceType": "traefik",
					"span.serviceURN":  "urn:service:/hostname",
				},
			},
		},
		{
			testCase: "Should interpret internal Traefik span",
			span: pb.Span{
				Name:    "TLSClientHeaders",
				Service: "TraefikService",
				Meta: map[string]string{
					"source": "traefik",
				},
			},
			expected: pb.Span{
				Name:    "TLSClientHeaders",
				Service: "TraefikService",
				Meta: map[string]string{
					"source":           "traefik",
					"span.serviceType": "traefik",
					"span.serviceName": "TraefikService",
					"span.serviceURN":  "urn:service:/TraefikService",
				},
			},
		},
		{
			testCase: "Should not interpret an already interpreted span",
			span: pb.Span{
				Service: "TraefikServiceName",
				Meta: map[string]string{
					"source":           "traefik",
					"http.host":        "hostname",
					"span.kind":        "server",
					"span.serviceType": "traefik",
					"span.serviceURN":  "some-different-external-urn-format",
				},
			},
			expected: pb.Span{
				Service: "TraefikServiceName",
				Meta: map[string]string{
					"source":           "traefik",
					"http.host":        "hostname",
					"span.kind":        "server",
					"span.serviceType": "traefik",
					"span.serviceURN":  "some-different-external-urn-format",
				},
			},
		},
		{
			testCase: "Open Telemetry interpret @opentelemetry/instrumentation-aws-lambda",
			span: pb.Span{
				Name:     "random-name",
				Start:    1586441095,
				Duration: 10000000,
				Resource: api.OpenTelemetrySource,
				Type:     api.OpenTelemetrySource,
				Service:  api.OpenTelemetrySource,
				Meta: map[string]string{
					"cloud.account.id":        "987654123",
					"instrumentation_library": "@opentelemetry/instrumentation-aws-lambda",
					"source":                  api.OpenTelemetrySource,
					"span.hostname":           api.OpenTelemetrySource,
					"faas.id":                 "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
				},
			},
			expected: pb.Span{
				Name:     "Lambda: function-name",
				Service:  "aws.lambda",
				Resource: "aws.lambda",
				Start:    1586441095,
				Error:    0,
				Duration: 10000000,
				Type:     "aws",
				Meta: map[string]string{
					"cloud.account.id":        "987654123",
					"faas.id":                 "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
					"instrumentation_library": "@opentelemetry/instrumentation-aws-lambda",
					"service":                 "aws.lambda",
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"span.hostname":           "openTelemetry",
					"span.kind":               "producer",
					"span.serviceName":        "Lambda: function-name",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/arn:aws:lambda:eu-west-1:965323806078:function:function-name",
					"sts.service.identifiers": "arn:aws:lambda:eu-west-1:965323806078:function:function-name",
				},
			},
		},
		{
			testCase: "Open Telemetry interpret @opentelemetry/instrumentation-http",
			span: pb.Span{
				Name:     "random-name",
				Start:    1586441095,
				Duration: 10000000,
				Resource: api.OpenTelemetrySource,
				Type:     api.OpenTelemetrySource,
				Service:  api.OpenTelemetrySource,
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-http",
					"source":                  api.OpenTelemetrySource,
					"span.hostname":           api.OpenTelemetrySource,
					"http.url":                "http://www.example.com/user/testing?queue=1#random-data",
					"http.method":             "GET",
				},
			},
			expected: pb.Span{
				Name:     "Http: GET - http://www.example.com/user/testing?queue=1#random-data",
				Service:  "aws.http",
				Resource: "aws.http",
				Start:    1586441095,
				Duration: 10000000,
				Error:    0,
				Type:     "aws",
				Meta: map[string]string{
					"http.method":             "GET",
					"http.url":                "http://www.example.com/user/testing?queue=1#random-data",
					"service":                 "aws.http",
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"span.hostname":           "openTelemetry",
					"span.kind":               "consumer",
					"span.serviceName":        "Http: GET - http://www.example.com/user/testing?queue=1#random-data",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/lambda-http-request/http://www.example.com/user/testing?queue=1#random-data/GET",
					"sts.service.identifiers": "http://www.example.com/user/testing?queue=1#random-data",
					"instrumentation_library": "@opentelemetry/instrumentation-http",
				},
			},
		},
		{
			testCase: "Open Telemetry interpret @opentelemetry/instrumentation-aws-sdk with service SQS",
			span: pb.Span{
				Name:     "random-name",
				Start:    1586441095,
				Duration: 10000000,
				Resource: api.OpenTelemetrySource,
				Type:     api.OpenTelemetrySource,
				Service:  api.OpenTelemetrySource,
				Meta: map[string]string{
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
					"source":                  api.OpenTelemetrySource,
					"span.hostname":           api.OpenTelemetrySource,
					"aws.service.identifier":  "sqs",
					"aws.region":              "eu-west-1",
					"aws.operation":           "publishMessage",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"messaging.destination":   "target-queue-destination",
				},
			},
			expected: pb.Span{
				Name:     "SQS Queue: target-queue-destination-9876543210-eu-west-1",
				Service:  "aws.sqs.queue",
				Resource: "aws.sqs.queue",
				Type:     "aws",
				Error:    0,
				Start:    1586441095,
				Duration: 10000000,
				Meta: map[string]string{
					"aws.operation":           "publishMessage",
					"aws.region":              "eu-west-1",
					"aws.service.identifier":  "sqs",
					"instrumentation_library": "@opentelemetry/instrumentation-aws-sdk",
					"messaging.destination":   "target-queue-destination",
					"messaging.url":           "https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"service":                 "aws.sqs.queue",
					"sts.origin":              "open-telemetry",
					"source":                  "open-telemetry",
					"span.hostname":           "openTelemetry",
					"span.kind":               "consumer",
					"span.serviceName":        "SQS Queue: target-queue-destination-9876543210-eu-west-1",
					"span.serviceType":        "open-telemetry",
					"span.serviceURN":         "urn:service:/https://sqs.eu-west-1.amazonaws.com/9876543210/target-queue-destination",
					"sts.service.identifiers": "https://eu-west-1.queue.amazonaws.com/9876543210/target-queue-destination",
				},
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			trace := []*pb.Span{&tc.span}
			actual := sie.Interpret(trace)
			assert.EqualValues(t, tc.expected, *actual[0])
		})
	}
}

*/
