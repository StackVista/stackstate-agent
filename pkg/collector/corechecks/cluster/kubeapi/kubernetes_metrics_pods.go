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
		k.podToMetricMappingForOutOfMemory(pod, sender)
	}
}

func (k *MetricsCheck) podToMetricMappingForOutOfMemory(pod *v1.Pod, sender aggregator.Sender) {
	// Go through the list of containers as we want to map a OOMEvent for each of these containers
	for _, container := range pod.Spec.Containers {
		value := float64(0)
		tags := []string{
			fmt.Sprintf("kube_cluster_name:%v", k.clusterName),
			fmt.Sprintf("kube_namespace:%v", pod.Namespace),
			fmt.Sprintf("pod:%v", pod.Name),
			fmt.Sprintf("pod_name:%v", pod.Name),
			fmt.Sprintf("container_name:%v", container.Name),
		}

		// Conditions:
		// 	 If the pod is not in a running or successful state
		// 	 If the pods previous state was a OOM then it means we still see it in a OOM state until the pod is back up and running
		if pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodSucceeded {
			// Go through the pods statuses and attempt to find a OOM state
			for _, containerStatus := range pod.Status.ContainerStatuses {
				// Determine that there should be a terminate state and that is OOMKilled
				// The container state mapped should be the same as the container we are looking for
				if containerStatus.LastTerminationState.Terminated != nil &&
					containerStatus.LastTerminationState.Terminated.Reason == "OOMKilled" &&
					containerStatus.Name == container.Name {
					// Set the value to 1 as we found a OOM event and break out of the loop as we do not need multiple OOM events
					value = 1
					break
				}
			}
		}

		log.Info(fmt.Sprintf("Sending metric kubernetes.state.container.status.report.count.oom (%v) ...", value))
		sender.Gauge("kubernetes.state.container.status.report.count.oom", value, "", tags)
	}
}
