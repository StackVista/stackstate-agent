// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package kubeapi

import (
	"encoding/json"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
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
	// TODO: Leader election, should include?
	c.ResyncPeriodPodEvent = defaultPodEventsResyncPeriodInSecond
}

func (k *EventsCheck) setDefaultsPodEvents() {
	// TODO: No ignored events as it is custom events, is this correct ?
	// TODO: k.ignoredEvents = convertFilter(k.instance.FilteredEventTypes)

	if k.instance.PodEventCollectionTimeoutMs == 0 {
		k.instance.PodEventCollectionTimeoutMs = defaultTimeoutPodEventsCollection
	}

	if k.instance.MaxPodEventsCollection == 0 {
		k.instance.MaxPodEventsCollection = maxPodEventsCardinality
	}
}

func (k *EventsCheck) podEventsCollectionCheck() (newPods []*v1.Pod, err error) {
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

	// TODO: Ignored events / Filter - Currently still passing in a blank string
	newPods, k.customPodEventCollection.LastResVer, k.customPodEventCollection.LastTime, err = k.ac.RunPodCollection(resourceVersion, lastTime, timeout, limit, resync, "")
	if err != nil {
		_ = k.Warnf("Could not collect pods from the api server: %s", err.Error()) //nolint:errcheck
		return nil, err
	}

	// Update the configMap to contain the new latest resources version so that we can continue from this version.
	configMapErr := k.ac.UpdateTokenInConfigmap(podEventTokenKey, k.customPodEventCollection.LastResVer, k.customPodEventCollection.LastTime)
	if configMapErr != nil {
		_ = k.Warnf("Could not store the LastCustomPodEventToken in the ConfigMap: %s", configMapErr.Error()) //nolint:errcheck
	}
	return newPods, nil
}

func (k *EventsCheck) processPodEvents(sender aggregator.Sender, pods []*v1.Pod) {
	log.Infof("---------- start processCustomPodEvents -----------")

	mapper := k.mapperFactory(k.ac, k.clusterName, k.instance.EventCategories)
	for _, pod := range pods {
		k.podEventMapper(pod, mapper, sender)
	}

	log.Infof("---------- end processCustomPodEvents -----------")
}

func (k *EventsCheck) podEventMapper(pod *v1.Pod, mapper *kubernetesEventMapper, sender aggregator.Sender) {
	log.Infof("---------- start podEventMapper -----------")

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.LastTerminationState.Terminated != nil &&
			containerStatus.LastTerminationState.Terminated.Reason == "OOMKilled" {
			k.mapEventForOutOfMemoryPod(pod, containerStatus, mapper, sender)
		}
	}

	log.Infof("---------- end podEventMapper -----------")
}

func (k *EventsCheck) mapEventForOutOfMemoryPod(pod *v1.Pod, containerStatus v1.ContainerStatus, mapper *kubernetesEventMapper, sender aggregator.Sender) {
	a, err := json.Marshal(pod)
	if err == nil {
		log.Infof("Test Print A: %v", string(a))
	} else {
		log.Info("Unable to parse Test Print A ...")
	}

	eventKind := "Pod"
	eventType := "warning"
	eventMessage := "Random Testing Message"
	eventCount := int32(1)
	eventReason := containerStatus.LastTerminationState.Terminated.Reason
	podUID := pod.UID
	podName := pod.Name
	podNameSpace := pod.Namespace
	containerName := containerStatus.Name
	startedAtTime := containerStatus.LastTerminationState.Terminated.StartedAt.Unix()
	endedAtTime := containerStatus.LastTerminationState.Terminated.FinishedAt.Unix()
	objKind := eventKind
	objName := podName
	// TODO: ???
	component := "kubelet"
	hostname := pod.Spec.NodeName

	event := &v1.Event{
		InvolvedObject: v1.ObjectReference{
			Name: podName,
			Kind: eventKind,
			// TODO: Remove types.UID
			UID:       podUID,
			Namespace: podNameSpace,
			// TODO: ????
			FieldPath: fmt.Sprintf("{%s %s %s %s %s}", objKind, podNameSpace, objName, podUID, containerName),
		},
		Count: eventCount,
		Type:  eventType,
		Source: v1.EventSource{
			Component: component,
			Host:      hostname,
		},
		Reason: eventReason,
		FirstTimestamp: obj.Time{
			Time: time.Unix(startedAtTime, 0),
		},
		LastTimestamp: obj.Time{
			Time: time.Unix(endedAtTime, 0),
		},
		Message: eventMessage,
	}

	b, err := json.Marshal(event)
	if err == nil {
		log.Infof("Test Print B: %v", string(b))
	} else {
		log.Info("Unable to parse Test Print B ...")
	}

	mEvent, err := mapper.mapKubernetesEvent(event)
	if err != nil {
		_ = k.Warnf("Error while mapping the pod event to a STS event, %s.", err.Error())
	}

	c, err := json.Marshal(mEvent)
	if err == nil {
		log.Infof("Test Print C: %v", string(c))
	} else {
		log.Info("Unable to parse Test Print C ...")
	}

	log.Debugf("Sending event: %s", mEvent.String())

	sender.Event(mEvent)
}
