// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
//go:build kubeapiserver
// +build kubeapiserver

package kubeapi

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/urn"
)

func TestEventTimestamps(t *testing.T) {
	event1 := createEvent(2, "default", "dca-789976f5d7-2ljx6", "Pod", "e6417a7f-f566-11e7-9749-0e4863e1cbf4", "default-scheduler", "machine-blue", "Unhealthy", "Liveness probe errored:", 709662600, 709662600, "Alerts")
	event2 := createEvent(2, "default", "dca-789976f5d7-2ljx6", "Pod", "e6417a7f-f566-11e7-9749-0e4863e1cbf4", "default-scheduler", "machine-blue", "Unhealthy", "Liveness probe errored:", 709662600, 709662800, "Alerts")
	event3 := createEvent(2, "default", "dca-789976f5d7-2ljx6", "Pod", "e6417a7f-f566-11e7-9749-0e4863e1cbf4", "default-scheduler", "machine-blue", "Unhealthy", "Liveness probe errored:", 709662800, 709662600, "Alerts")

	mapper := &kubernetesEventMapper{
		urn:                     urn.NewURNBuilder(urn.Kubernetes, "testCluster"),
		clusterName:             "testCluster",
		sourceType:              string(urn.Kubernetes),
		eventCategoriesOverride: nil,
	}

	mEvent1, _ := mapper.mapKubernetesEvent(event1)
	mEvent2, _ := mapper.mapKubernetesEvent(event2)
	mEvent3, _ := mapper.mapKubernetesEvent(event3)

	assert.Equal(t, int64(709662600), mEvent1.Ts)
	assert.Equal(t, int64(709662800), mEvent2.Ts)
	assert.Equal(t, int64(709662600), mEvent3.Ts)
}
