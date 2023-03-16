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

func TestProcessPodEvents(t *testing.T) {
	rawPodEvents, _ := ioutil.ReadFile("./testdata/process_pod_events_data.json")

	var podEvents []*v1.Pod
	err := json.Unmarshal(rawPodEvents, &podEvents)
	if err != nil {
		t.Fatal("Unable to case process_pod_events_data.json into []*v1.Event")
	}

	evCheck := KubernetesAPIEventsFactory().(*EventsCheck)
	evCheck.ac = MockAPIClient(nil)

	mockSender := mocksender.NewMockSender(evCheck.ID())
	mockSender.On("Event", mock.AnythingOfType("metrics.Event"))

	evCheck.processPods(mockSender, podEvents)
	mockSender.AssertNumberOfCalls(t, "Event", 1)
}
