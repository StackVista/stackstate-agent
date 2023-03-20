// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver
// +build kubeapiserver

package kubeapi

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/aggregator/mocksender"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

func TestProcessPodMetrics(t *testing.T) {
	rawPodEvents, _ := ioutil.ReadFile("./testdata/process_pod_metrics_data.json")

	var podEvents []*v1.Pod
	err := json.Unmarshal(rawPodEvents, &podEvents)
	if err != nil {
		t.Fatal("Unable to case process_pod_metrics_data.json into []*v1.Event")
	}

	check := KubernetesAPIMetricsFactory().(*MetricsCheck)
	check.ac = MockAPIClient(nil)

	mockSender := mocksender.NewMockSender(check.ID())
	mockSender.On("Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	check.processPods(mockSender, podEvents)

	mockSender.AssertNumberOfCalls(t, "Gauge", 24)

	mockSender.AssertCalled(t, "Gauge", "kubernetes.state.container.status.report.count.oom", float64(0), "", []string{
		"kube_cluster_name:",
		"kube_namespace:kubernetes-monitors",
		"pod:rrf-broken-app-6d8f67cf4d-7cn6l",
		"pod_name:rrf-broken-app-6d8f67cf4d-7cn6l",
		"container_name:broken-app",
	})

	mockSender.AssertCalled(t, "Gauge", "kubernetes.state.container.status.report.count.oom", float64(1), "", []string{
		"kube_cluster_name:",
		"kube_namespace:kubernetes-monitors",
		"pod:out-of-memory-always-critical",
		"pod_name:out-of-memory-always-critical",
		"container_name:out-of-memory-always-critical",
	})
}
