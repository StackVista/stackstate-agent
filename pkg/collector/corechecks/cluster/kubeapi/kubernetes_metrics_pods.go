// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package kubeapi

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
)

const (
	podMetricTokenKey                     = "pod-metrics"
	maxPodMetricsCardinality              = 300
	defaultPodMetricsResyncPeriodInSecond = 300
	defaultTimeoutPodMetricsCollection    = 2000
)

func (c *MetricsConfig) parsePodMetrics() {
	c.ResyncPeriodPodMetric = defaultPodMetricsResyncPeriodInSecond
}

func (k *MetricsCheck) setDefaultsPodEvents() {
	if k.instance.PodMetricCollectionTimeoutMs == 0 {
		k.instance.PodMetricCollectionTimeoutMs = defaultTimeoutPodMetricsCollection
	}

	if k.instance.MaxPodMetricsCollection == 0 {
		k.instance.MaxPodMetricsCollection = maxPodMetricsCardinality
	}
}

// podMetricsCollectionCheck
func (k *MetricsCheck) podMetricsCollectionCheck() (pods []*v1.Pod, err error) {
	// Retrieve the last resource version that was used within the pod metric collection cycle.
	// This allows us to continue from a certain point in time and not re-fetch all the pods over and over
	resourceVersion, lastTime, err := k.ac.GetTokenFromConfigmap(podMetricTokenKey)
	if err != nil {
		return nil, err
	}

	// This is to avoid getting in a situation where we list all the events for multiple runs in a row.
	if resourceVersion == "" && k.podMetricCollection.LastResVer != "" {
		_ = log.Errorf("Resource Version stored in the ConfigMap is incorrect. Will resume collecting from %s", k.podMetricCollection.LastResVer)
		resourceVersion = k.podMetricCollection.LastResVer
	}

	timeout := int64(k.instance.PodMetricCollectionTimeoutMs / 1000)
	limit := int64(k.instance.MaxPodMetricsCollection)
	resync := int64(k.instance.ResyncPeriodPodMetric)

	pods, k.podMetricCollection.LastResVer, k.podMetricCollection.LastTime, err = k.ac.RunPodCollection(resourceVersion, lastTime, timeout, limit, resync)
	if err != nil {
		_ = k.Warnf("Could not collect pods from the api server: %s", err.Error()) //nolint:errcheck
		return nil, err
	}

	// Update the configMap to contain the new latest resources version so that we can continue from this version.
	configMapErr := k.ac.UpdateTokenInConfigmap(podMetricTokenKey, k.podMetricCollection.LastResVer, k.podMetricCollection.LastTime)
	if configMapErr != nil {
		_ = k.Warnf("Could not store the pod metric token in the ConfigMap: %s", configMapErr.Error()) //nolint:errcheck
	}

	return pods, nil
}

func (k *MetricsCheck) processPods(sender aggregator.Sender, pods []*v1.Pod) {
	log.Info("Running kubernetes pod metric collector - processPods...")

	for _, pod := range pods {
		k.podToMetricMapper(pod, sender)
	}
}

func (k *MetricsCheck) podToMetricMapper(pod *v1.Pod, sender aggregator.Sender) {
	log.Info("Running kubernetes pod metric collector - podToMetricMapper...")
	tags := []string{fmt.Sprintf("example:%d", 10)}
	sender.Gauge("hello.world.testing", 1, "", tags)
}
