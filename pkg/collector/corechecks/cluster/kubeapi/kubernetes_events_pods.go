// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package kubeapi

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
)

// Covers the Control Plane service check and the in memory pod metadata.
const (
	customPodEventTokenKey                    = "pod-event"
	maxCustomPodEventCardinality              = 300
	defaultCustomPodEventResyncPeriodInSecond = 300
	defaultTimeoutCustomPodEventCollection    = 2000
)

// TODO: Leader election, should include?
func (c *EventsConfig) parsePods() {
	c.ResyncPeriodCustomPodEvents = defaultCustomPodEventResyncPeriodInSecond
}

// TODO: No ignored events as it is custom events, is this correct ?
// TODO: k.ignoredEvents = convertFilter(k.instance.FilteredEventTypes)
func (k *EventsCheck) setDefaultsForPods() {
	if k.instance.CustomPodEventCollectionTimeoutMs == 0 {
		k.instance.CustomPodEventCollectionTimeoutMs = defaultTimeoutCustomPodEventCollection
	}

	if k.instance.MaxCustomPodEventCollection == 0 {
		k.instance.MaxCustomPodEventCollection = maxCustomPodEventCardinality
	}
}

func (k *EventsCheck) podCollectionCheck() (newPods []*v1.Pod, err error) {
	// Retrieve the last resource version that was used within the customPodEvent collection cycle.
	// This allows us to continue from a certain point in time and not re-fetch all the pods over and over
	resourceVersion, lastTime, err := k.ac.GetTokenFromConfigmap(customPodEventTokenKey)
	if err != nil {
		return nil, err
	}

	// This is to avoid getting in a situation where we list all the events for multiple runs in a row.
	if resourceVersion == "" && k.customPodEventCollection.LastResVer != "" {
		_ = log.Errorf("Resource Version stored in the ConfigMap is incorrect. Will resume collecting from %s", k.customPodEventCollection.LastResVer)
		resourceVersion = k.customPodEventCollection.LastResVer
	}

	timeout := int64(k.instance.CustomPodEventCollectionTimeoutMs / 1000)
	limit := int64(k.instance.MaxCustomPodEventCollection)
	resync := int64(k.instance.ResyncPeriodCustomPodEvents)

	// TODO: Ignored events / Filter - Currently still passing in a blank string
	newPods, k.customPodEventCollection.LastResVer, k.customPodEventCollection.LastTime, err = k.ac.RunPodCollection(resourceVersion, lastTime, timeout, limit, resync, "")
	if err != nil {
		_ = k.Warnf("Could not collect pods from the api server: %s", err.Error()) //nolint:errcheck
		return nil, err
	}

	// Update the configMap to contain the new latest resources version so that we can continue from this version.
	configMapErr := k.ac.UpdateTokenInConfigmap(customPodEventTokenKey, k.customPodEventCollection.LastResVer, k.customPodEventCollection.LastTime)
	if configMapErr != nil {
		_ = k.Warnf("Could not store the LastCustomPodEventToken in the ConfigMap: %s", configMapErr.Error()) //nolint:errcheck
	}
	return newPods, nil
}

// processCustomPodEvents Lorem Ipsum
func (k *EventsCheck) processCustomPodEvents(sender aggregator.Sender, pods []*v1.Pod) error {
	log.Infof("---------- start processCustomPodEvents -----------")

	podsJSON, err := json.Marshal(pods)
	if err == nil {
		log.Infof("Found Custom Pod Events: %v", string(podsJSON))
	} else {
		log.Info("Unable to parse customPodEvents ...")
	}

	log.Infof("k.instance.EventCategories: %v", k.instance.EventCategories)

	mapper := k.mapperFactory(k.ac, k.clusterName, k.instance.EventCategories)

	log.Infof("mapper.clusterName: %v", mapper.clusterName)
	log.Infof("mapper.urn: %v", mapper.urn)
	log.Infof("mapper.sourceType: %v", mapper.sourceType)
	log.Infof("---------- end processCustomPodEvents -----------")

	//  mapper := k.mapperFactory(k.ac, k.clusterName, k.instance.EventCategories)
	//  for _, event := range events {
	//  	mappedEvent, err := mapper.mapKubernetesEvent(event)
	//  	if err != nil {
	//  		_ = k.Warnf("Error while mapping event, %s.", err.Error())
	//  		continue
	//  	}
	//  	log.Debugf("Sending event: %s", mappedEvent.String())
	//  	sender.Event(mappedEvent)
	//  }

	return nil
}
