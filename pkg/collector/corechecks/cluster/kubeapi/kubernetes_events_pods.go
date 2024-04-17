// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

//go:build kubeapiserver

package kubeapi

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/metrics"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
	obj "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	podEventTokenKey                     = "pod-event"
	maxPodEventsCardinality              = 300
	defaultPodEventsResyncPeriodInSecond = 300
	defaultTimeoutPodEventsCollection    = 2000
)

func (c *EventsConfig) parsePodEvents() {
	c.ResyncPeriodPodEvent = defaultPodEventsResyncPeriodInSecond
}

func (k *EventsCheck) setDefaultsPodEvents() {
	if k.instance.PodEventCollectionTimeoutMs == 0 {
		k.instance.PodEventCollectionTimeoutMs = defaultTimeoutPodEventsCollection
	}

	if k.instance.MaxPodEventsCollection == 0 {
		k.instance.MaxPodEventsCollection = maxPodEventsCardinality
	}
}

// podEventsCollectionCheck The collection check triggered as part of the kubernetes_event this will create custom events based on specific pod conditions
func (k *EventsCheck) podEventsCollectionCheck() (pods []*v1.Pod, err error) {
	// Retrieve the last resource version that was used within the customPodEvent collection cycle.
	// This allows us to continue from a certain point in time and not re-fetch all the pods over and over
	resourceVersion, lastTime, err := k.ac.GetTokenFromConfigmap(podEventTokenKey)
	if err != nil {
		return nil, err
	}

	// This is to avoid getting in a situation where we list all the events for multiple runs in a row.
	if resourceVersion == "" && k.customPodEventCollection.LastResVer != "" {
		_ = log.Errorf("Resource Version stored in the ConfigMap is incorrect. Will resume collecting from %s", k.customPodEventCollection.LastResVer)
		resourceVersion = k.customPodEventCollection.LastResVer
	}

	timeout := int64(k.instance.PodEventCollectionTimeoutMs / 1000)
	limit := int64(k.instance.MaxPodEventsCollection)
	resync := int64(k.instance.ResyncPeriodPodEvent)

	pods, k.customPodEventCollection.LastResVer, k.customPodEventCollection.LastTime, err = k.ac.RunPodCollection(lastTime, timeout, limit, resync, resourceVersion, false)
	if err != nil {
		_ = k.Warnf("Could not collect pods from the api server: %s", err.Error()) //nolint:errcheck
		return nil, err
	}

	// Update the configMap to contain the new latest resources version so that we can continue from this version.
	configMapErr := k.ac.UpdateTokenInConfigmap(podEventTokenKey, k.customPodEventCollection.LastResVer, k.customPodEventCollection.LastTime)
	if configMapErr != nil {
		_ = k.Warnf("Could not store the LastCustomPodEventToken in the ConfigMap: %s", configMapErr.Error()) //nolint:errcheck
	}

	return pods, nil
}

func (k *EventsCheck) processPods(sender aggregator.Sender, pods []*v1.Pod) {
	mapper := k.mapperFactory(k.ac, k.clusterName, k.instance.EventCategories)
	for _, pod := range pods {
		events := k.podToEventMapper(pod, mapper, sender)

		for _, event := range events {
			log.Debug("Sending metric pod event: %s", event.String())
			sender.Event(event)
		}
	}
}

func (k *EventsCheck) podToEventMapper(pod *v1.Pod, mapper *kubernetesEventMapper, sender aggregator.Sender) []metrics.Event {
	var events []metrics.Event

	// Test on active Status. This will be the current state the pod is in not the previous state
	for _, containerStatus := range pod.Status.ContainerStatuses {
		// Terminated is an optional as the state can be anything, for example waiting will make terminated nil
		if containerStatus.State.Terminated != nil {
			switch containerStatus.State.Terminated.Reason {
			case "OOMKilled":
				event, err := k.mapPodToMetricEventForOutOfMemory(pod, containerStatus.Name, containerStatus.State.Terminated, mapper)
				if err != nil {
					_ = log.Errorf("Could not map pod to metric event: %s", err.Error())
					break
				}
				events = append(events, event)
			}
		}
	}

	return events
}

// mapPodToMetricEventForOutOfMemory Attempt to map a pod to a metric event which can be forwarded to the aggregator
func (k *EventsCheck) mapPodToMetricEventForOutOfMemory(pod *v1.Pod, containerName string, terminatedState *v1.ContainerStateTerminated, mapper *kubernetesEventMapper) (metrics.Event, error) {
	event := &v1.Event{
		InvolvedObject: v1.ObjectReference{
			Name:      pod.Name,
			Kind:      "Pod",
			UID:       pod.UID,
			Namespace: pod.Namespace,
			FieldPath: fmt.Sprintf("{%s %s %s %s %s}", "Pod", pod.Namespace, pod.Name, pod.UID, containerName),
		},
		Count: int32(1),
		Type:  "warning",
		Source: v1.EventSource{
			Host: pod.Spec.NodeName,
		},
		Reason: terminatedState.Reason,
		FirstTimestamp: obj.Time{
			Time: time.Unix(terminatedState.StartedAt.Unix(), 0),
		},
		LastTimestamp: obj.Time{
			Time: time.Unix(terminatedState.FinishedAt.Unix(), 0),
		},
		Message: fmt.Sprintf("Container '%s' was killed due to an out of memory (OOM) condition", containerName),
	}

	metricEvent, err := mapper.mapKubernetesEvent(event)
	if err != nil {
		_ = k.Warnf("Error while mapping a kubernetes event to a metric event, %s.", err.Error())
	}

	return metricEvent, nil
}
