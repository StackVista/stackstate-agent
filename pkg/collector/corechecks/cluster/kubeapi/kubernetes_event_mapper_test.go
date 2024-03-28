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

var firstTimestamp = int64(709662600)
var laterTimeSTamp = int64(709662800)

var eventContainer1 = "stackstate-agent"
var eventContainer2 = "cluster-agent"

var event1 = createEvent(2, "default", "dca-789976f5d7-2ljx6", "Pod", "e6417a7f-f566-11e7-9749-0e4863e1cbf4", "default-scheduler", "machine-blue", "Unhealthy", "Liveness probe errored:", firstTimestamp, firstTimestamp, "Alerts", eventContainer1)
var event2 = createEvent(2, "default", "dca-789976f5d7-2ljx6", "Pod", "e6417a7f-f566-11e7-9749-0e4863e1cbf4", "default-scheduler", "machine-blue", "Unhealthy", "Liveness probe errored:", firstTimestamp, laterTimeSTamp, "Alerts", eventContainer2)
var event3 = createEvent(2, "default", "dca-789976f5d7-2ljx6", "Pod", "e6417a7f-f566-11e7-9749-0e4863e1cbf4", "default-scheduler", "machine-blue", "Unhealthy", "Liveness probe errored:", laterTimeSTamp, firstTimestamp, "Alerts", "")

func TestEventTimestamps(t *testing.T) {

	mapper := &kubernetesEventMapper{
		urn:                     urn.NewURNBuilder(urn.Kubernetes, "testCluster"),
		clusterName:             "testCluster",
		sourceType:              string(urn.Kubernetes),
		eventCategoriesOverride: nil,
	}

	mEvent1, _ := mapper.mapKubernetesEvent(event1)
	mEvent2, _ := mapper.mapKubernetesEvent(event2)
	mEvent3, _ := mapper.mapKubernetesEvent(event3)

	assert.Equal(t, firstTimestamp, mEvent1.Ts)
	assert.Equal(t, laterTimeSTamp, mEvent2.Ts)
	assert.Equal(t, firstTimestamp, mEvent3.Ts)
}

func TestContainerNameFromEvent(t *testing.T) {

	containerName1 := getContainerNameFromEvent(event1)
	containerName2 := getContainerNameFromEvent(event2)
	containerName3 := getContainerNameFromEvent(event3)

	assert.Equal(t, eventContainer1, containerName1)
	assert.Equal(t, eventContainer2, containerName2)
	assert.Equal(t, "", containerName3)
}

func TestEventTags(t *testing.T) {

	mapper := &kubernetesEventMapper{
		urn:                     urn.NewURNBuilder(urn.Kubernetes, "testCluster"),
		clusterName:             "testCluster",
		sourceType:              string(urn.Kubernetes),
		eventCategoriesOverride: nil,
	}

	eventTags1 := mapper.getTags(event1)
	eventTags2 := mapper.getTags(event2)
	eventTags3 := mapper.getTags(event3)

	assert.Contains(t, eventTags1, "kube_container_name:"+eventContainer1)
	assert.Contains(t, eventTags2, "kube_container_name:"+eventContainer2)
	assert.NotContains(t, eventTags3, "kube_container_name:")
}

func TestEventElementIdentifiers(t *testing.T) {

	mapper := &kubernetesEventMapper{
		urn:                     urn.NewURNBuilder(urn.Kubernetes, "testCluster"),
		clusterName:             "testCluster",
		sourceType:              string(urn.Kubernetes),
		eventCategoriesOverride: nil,
	}

	elementIdentifier1 := mapper.externalIdentifierForInvolvedObject(event1)
	elementIdentifier2 := mapper.externalIdentifierForInvolvedObject(event2)
	elementIdentifier3 := mapper.externalIdentifierForInvolvedObject(event3)

	assert.Equal(t, 2, len(elementIdentifier1))
	assert.Contains(t, elementIdentifier1, "urn:kubernetes:/testCluster:default:pod/dca-789976f5d7-2ljx6:container/"+eventContainer1)
	assert.Contains(t, elementIdentifier1, "urn:kubernetes:/testCluster:default:pod/dca-789976f5d7-2ljx6")

	assert.Equal(t, 2, len(elementIdentifier2))
	assert.Contains(t, elementIdentifier2, "urn:kubernetes:/testCluster:default:pod/dca-789976f5d7-2ljx6:container/"+eventContainer2)
	assert.Contains(t, elementIdentifier2, "urn:kubernetes:/testCluster:default:pod/dca-789976f5d7-2ljx6")

	assert.Equal(t, 1, len(elementIdentifier3))
	assert.Contains(t, elementIdentifier3, "urn:kubernetes:/testCluster:default:pod/dca-789976f5d7-2ljx6")
}
