// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021-present Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package ksm

import (
	"gotest.tools/assert"
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/aggregator/mocksender"
	core "github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	ksmstore "github.com/StackVista/stackstate-agent/pkg/kubestatemetrics/store"
)

var _ metricAggregator = &sumValuesAggregator{}
var _ metricAggregator = &countObjectsAggregator{}

func Test_counterAggregator(t *testing.T) {
	tests := []struct {
		name          string
		ddMetricName  string
		allowedLabels []string
		metrics       []ksmstore.DDMetric
		expected      []metricsExpected
	}{
		{
			name:          "One allowed label",
			ddMetricName:  "my.count",
			allowedLabels: []string{"foo"},
			metrics: []ksmstore.DDMetric{
				{
					Labels: map[string]string{
						"foo": "foo1",
						"bar": "bar1",
					},
					Val: 1,
				},
				{
					Labels: map[string]string{
						"foo": "foo1",
						"bar": "bar2",
					},
					Val: 2,
				},
				{
					Labels: map[string]string{
						"foo": "foo2",
						"bar": "bar1",
					},
					Val: 4,
				},
				{
					Labels: map[string]string{
						"foo": "foo2",
						"bar": "bar2",
					},
					Val: 8,
				},
			},
			expected: []metricsExpected{
				{
					name: "kubernetes_state.my.count",
					val:  1 + 2,
					tags: []string{"foo:foo1"},
				},
				{
					name: "kubernetes_state.my.count",
					val:  4 + 8,
					tags: []string{"foo:foo2"},
				},
			},
		},
		{
			name:          "Two allowed labels",
			ddMetricName:  "my.count",
			allowedLabels: []string{"foo", "bar"},
			metrics: []ksmstore.DDMetric{
				{
					Labels: map[string]string{
						"foo": "foo1",
						"bar": "bar1",
						"baz": "baz1",
					},
					Val: 1,
				},
				{
					Labels: map[string]string{
						"foo": "foo1",
						"bar": "bar1",
						"baz": "baz2",
					},
					Val: 2,
				},
				{
					Labels: map[string]string{
						"foo": "foo1",
						"bar": "bar2",
						"baz": "baz1",
					},
					Val: 4,
				},
				{
					Labels: map[string]string{
						"foo": "foo1",
						"bar": "bar2",
						"baz": "baz2",
					},
					Val: 8,
				},
				{
					Labels: map[string]string{
						"foo": "foo2",
						"bar": "bar1",
						"baz": "baz1",
					},
					Val: 16,
				},
				{
					Labels: map[string]string{
						"foo": "foo2",
						"bar": "bar1",
						"baz": "baz2",
					},
					Val: 32,
				},
				{
					Labels: map[string]string{
						"foo": "foo2",
						"bar": "bar2",
						"baz": "baz1",
					},
					Val: 64,
				},
				{
					Labels: map[string]string{
						"foo": "foo2",
						"bar": "bar2",
						"baz": "baz2",
					},
					Val: 128,
				},
			},
			expected: []metricsExpected{
				{
					name: "kubernetes_state.my.count",
					val:  1 + 2,
					tags: []string{"foo:foo1", "bar:bar1"},
				},
				{
					name: "kubernetes_state.my.count",
					val:  4 + 8,
					tags: []string{"foo:foo1", "bar:bar2"},
				},
				{
					name: "kubernetes_state.my.count",
					val:  16 + 32,
					tags: []string{"foo:foo2", "bar:bar1"},
				},
				{
					name: "kubernetes_state.my.count",
					val:  64 + 128,
					tags: []string{"foo:foo2", "bar:bar2"},
				},
			},
		},
	}

	ksmCheck := newKSMCheck(core.NewCheckBase(kubeStateMetricsCheckName), &KSMConfig{})

	for _, tt := range tests {
		s := mocksender.NewMockSender("ksm")
		s.SetupAcceptAll()

		t.Run(tt.name, func(t *testing.T) {
			agg := newSumValuesAggregator(tt.ddMetricName, "", tt.allowedLabels)
			for _, metric := range tt.metrics {
				agg.accumulate(metric)
			}

			agg.flush(s, ksmCheck, newLabelJoiner(ksmCheck.instance.LabelJoins))

			s.AssertNumberOfCalls(t, "Gauge", len(tt.expected))
			for _, expected := range tt.expected {
				s.AssertMetric(t, "Gauge", expected.name, expected.val, expected.hostname, expected.tags)
			}
		})
	}
}

func Test_aggregateStatusReasonMetrics(t *testing.T) {
	tests := []struct {
		name         string
		metricFamily []ksmstore.DDMetricsFam
		accumulator  map[string]ksmstore.DDMetric
		isZeroValue  bool
		expected     []ksmstore.DDMetricsFam
	}{
		{
			name: "Test A",
			metricFamily: []ksmstore.DDMetricsFam{
				{
					Type: "*v1.Pod",
					Name: "kube_pod_container_info",
					ListMetrics: []ksmstore.DDMetric{
						{
							Labels: map[string]string{
								"container":    "restarts-increment-always-critical",
								"container_id": "docker://abfe1487e744da362ae503e20bf363544db13f1fa2dc491601a92cb3ce1ac3c3",
								"image":        "registry.k8s.io/busybox:latest",
								"image_id":     "docker-pullable://registry.k8s.io/busybox@sha256:d8d3bc2c183ed2f9f10e7258f84971202325ee6011ba137112e01e30f206de67",
								"namespace":    "kubernetes-monitors",
								"pod":          "restarts-increment-always-critical",
								"uid":          "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
						{
							Labels: map[string]string{
								"container":    "this-should-not-exist-in-reasons-and-get-a-zero-state",
								"container_id": "docker://abfe1487e744da362ae503e20bf363544db13f1fa2dc491601a92cb3ce1ac3c3",
								"image":        "registry.k8s.io/busybox:latest",
								"image_id":     "docker-pullable://registry.k8s.io/busybox@sha256:d8d3bc2c183ed2f9f10e7258f84971202325ee6011ba137112e01e30f206de67",
								"namespace":    "kubernetes-monitors",
								"pod":          "restarts-increment-always-critical",
								"uid":          "000000-000000-00000-000000-00000",
							},
							Val: 1,
						},
					},
				},
				{
					Type: "*v1.Pod",
					Name: "kube_pod_container_status_waiting_reason",
					ListMetrics: []ksmstore.DDMetric{
						{
							Labels: map[string]string{
								"container": "restarts-increment-always-critical",
								"namespace": "kubernetes-monitors",
								"pod":       "restarts-increment-always-critical",
								"reason":    "Pending",
								"uid":       "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
						{
							Labels: map[string]string{
								"container": "restarts-increment-always-critical",
								"namespace": "kubernetes-monitors",
								"pod":       "restarts-increment-always-critical",
								"reason":    "CrashLoopBackOff",
								"uid":       "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
						{
							Labels: map[string]string{
								"container": "restarts-increment-always-critical",
								"namespace": "kubernetes-monitors",
								"pod":       "restarts-increment-always-critical",
								"reason":    "Idle",
								"uid":       "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
					},
				},
				{
					Type: "*v1.Pod",
					Name: "kube_pod_container_info",
					ListMetrics: []ksmstore.DDMetric{
						{
							Labels: map[string]string{
								"container":    "restarts-increment-always-critical",
								"container_id": "docker://abfe1487e744da362ae503e20bf363544db13f1fa2dc491601a92cb3ce1ac3c3",
								"image":        "registry.k8s.io/busybox:latest",
								"image_id":     "docker-pullable://registry.k8s.io/busybox@sha256:d8d3bc2c183ed2f9f10e7258f84971202325ee6011ba137112e01e30f206de67",
								"namespace":    "kubernetes-monitors",
								"pod":          "restarts-increment-always-critical",
								"uid":          "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
					},
				},
			},
			expected: []ksmstore.DDMetricsFam{
				{
					Type: "*v1.Pod",
					Name: "kube_pod_container_info",
					ListMetrics: []ksmstore.DDMetric{
						{
							Labels: map[string]string{
								"container":    "restarts-increment-always-critical",
								"container_id": "docker://abfe1487e744da362ae503e20bf363544db13f1fa2dc491601a92cb3ce1ac3c3",
								"image":        "registry.k8s.io/busybox:latest",
								"image_id":     "docker-pullable://registry.k8s.io/busybox@sha256:d8d3bc2c183ed2f9f10e7258f84971202325ee6011ba137112e01e30f206de67",
								"namespace":    "kubernetes-monitors",
								"pod":          "restarts-increment-always-critical",
								"reason":       "Unknown",
								"uid":          "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
						{
							Labels: map[string]string{
								"container":    "this-should-not-exist-in-reasons-and-get-a-zero-state",
								"container_id": "docker://abfe1487e744da362ae503e20bf363544db13f1fa2dc491601a92cb3ce1ac3c3",
								"image":        "registry.k8s.io/busybox:latest",
								"image_id":     "docker-pullable://registry.k8s.io/busybox@sha256:d8d3bc2c183ed2f9f10e7258f84971202325ee6011ba137112e01e30f206de67",
								"namespace":    "kubernetes-monitors",
								"pod":          "restarts-increment-always-critical",
								"reason":       "Unknown",
								"uid":          "000000-000000-00000-000000-00000",
							},
							Val: 1,
						},
					},
				},
				{
					Type: "*v1.Pod",
					Name: "kube_pod_container_status_waiting_reason",
					ListMetrics: []ksmstore.DDMetric{
						{
							Labels: map[string]string{
								"container": "restarts-increment-always-critical",
								"namespace": "kubernetes-monitors",
								"pod":       "restarts-increment-always-critical",
								"reason":    "Pending",
								"uid":       "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
						{
							Labels: map[string]string{
								"container": "restarts-increment-always-critical",
								"namespace": "kubernetes-monitors",
								"pod":       "restarts-increment-always-critical",
								"reason":    "CrashLoopBackOff",
								"uid":       "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
						{
							Labels: map[string]string{
								"container": "restarts-increment-always-critical",
								"namespace": "kubernetes-monitors",
								"pod":       "restarts-increment-always-critical",
								"reason":    "Idle",
								"uid":       "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
					},
				},
				{
					Type: "*v1.Pod",
					Name: "kube_pod_container_info",
					ListMetrics: []ksmstore.DDMetric{
						{
							Labels: map[string]string{
								"container":    "restarts-increment-always-critical",
								"container_id": "docker://abfe1487e744da362ae503e20bf363544db13f1fa2dc491601a92cb3ce1ac3c3",
								"image":        "registry.k8s.io/busybox:latest",
								"image_id":     "docker-pullable://registry.k8s.io/busybox@sha256:d8d3bc2c183ed2f9f10e7258f84971202325ee6011ba137112e01e30f206de67",
								"namespace":    "kubernetes-monitors",
								"pod":          "restarts-increment-always-critical",
								"reason":       "Unknown",
								"uid":          "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
					},
				},
				{
					Type: "",
					Name: "kube_pod_container_status_reasons",
					ListMetrics: []ksmstore.DDMetric{
						{
							Labels: map[string]string{
								"container":    "restarts-increment-always-critical",
								"container_id": "docker://abfe1487e744da362ae503e20bf363544db13f1fa2dc491601a92cb3ce1ac3c3",
								"image":        "registry.k8s.io/busybox:latest",
								"image_id":     "docker-pullable://registry.k8s.io/busybox@sha256:d8d3bc2c183ed2f9f10e7258f84971202325ee6011ba137112e01e30f206de67",
								"namespace":    "kubernetes-monitors",
								"pod":          "restarts-increment-always-critical",
								"reason":       "Unknown",
								"uid":          "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 0,
						},
						{
							Labels: map[string]string{
								"container":    "this-should-not-exist-in-reasons-and-get-a-zero-state",
								"container_id": "docker://abfe1487e744da362ae503e20bf363544db13f1fa2dc491601a92cb3ce1ac3c3",
								"image":        "registry.k8s.io/busybox:latest",
								"image_id":     "docker-pullable://registry.k8s.io/busybox@sha256:d8d3bc2c183ed2f9f10e7258f84971202325ee6011ba137112e01e30f206de67",
								"namespace":    "kubernetes-monitors",
								"pod":          "restarts-increment-always-critical",
								"reason":       "Unknown",
								"uid":          "000000-000000-00000-000000-00000",
							},
							Val: 0,
						},
						{
							Labels: map[string]string{
								"container":    "restarts-increment-always-critical",
								"container_id": "docker://abfe1487e744da362ae503e20bf363544db13f1fa2dc491601a92cb3ce1ac3c3",
								"image":        "registry.k8s.io/busybox:latest",
								"image_id":     "docker-pullable://registry.k8s.io/busybox@sha256:d8d3bc2c183ed2f9f10e7258f84971202325ee6011ba137112e01e30f206de67",
								"namespace":    "kubernetes-monitors",
								"pod":          "restarts-increment-always-critical",
								"reason":       "Unknown",
								"uid":          "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 0,
						},
						{
							Labels: map[string]string{
								"container": "restarts-increment-always-critical",
								"namespace": "kubernetes-monitors",
								"pod":       "restarts-increment-always-critical",
								"reason":    "Pending",
								"uid":       "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
						{
							Labels: map[string]string{
								"container": "restarts-increment-always-critical",
								"namespace": "kubernetes-monitors",
								"pod":       "restarts-increment-always-critical",
								"reason":    "CrashLoopBackOff",
								"uid":       "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
						{
							Labels: map[string]string{
								"container": "restarts-increment-always-critical",
								"namespace": "kubernetes-monitors",
								"pod":       "restarts-increment-always-critical",
								"reason":    "Idle",
								"uid":       "c69e618f-dcca-4fe6-89de-c908ecbf662f",
							},
							Val: 1,
						},
					},
				},
			},
			accumulator: map[string]ksmstore.DDMetric{},
			isZeroValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregatedResults := aggregateStatusReasonMetrics(tt.metricFamily)
			assert.DeepEqual(t, &aggregatedResults, &tt.expected)
		})
	}
}
