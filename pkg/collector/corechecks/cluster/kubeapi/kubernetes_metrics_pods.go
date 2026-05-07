// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

//go:build kubeapiserver

package kubeapi

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
	"time"
)

const (
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
	timeout := int64(k.instance.PodMetricCollectionTimeoutMs / 1000)
	limit := int64(k.instance.MaxPodMetricsCollection)
	resync := int64(k.instance.ResyncPeriodPodMetric)

	// We always start with a fresh resourceVersion, this allows the pod collection to collect all pods on the first run and only report changes within the check run cycle
	// This allows us to always report a zero state, but also capture and changes that happens while the agent is running
	pods, k.podMetricCollection.LastResVer, k.podMetricCollection.LastTime, err = k.ac.RunPodCollection(time.Now(), timeout, limit, resync, "0", true)
	if err != nil {
		_ = k.Warnf("Could not collect pods from the api server: %s", err.Error()) //nolint:errcheck
		return nil, err
	}

	return pods, nil
}

func (k *MetricsCheck) processPods(sender sender.Sender, pods []*v1.Pod) {
	log.Info("Running kubernetes pod metric collector - processPods...")

	for _, pod := range pods {
		k.podToMetricMappingForOutOfMemory(pod, sender)
	}
}

func (k *MetricsCheck) podToMetricMappingForOutOfMemory(pod *v1.Pod, sender sender.Sender) {
	// Go through the pods statuses and attempt to find a OOM state
	for _, containerStatus := range pod.Status.ContainerStatuses {
		value := float64(0)
		tags := []string{
			fmt.Sprintf("kube_cluster_name:%v", k.clusterName),
			fmt.Sprintf("kube_namespace:%v", pod.Namespace),
			fmt.Sprintf("pod:%v", pod.Name),
			fmt.Sprintf("pod_name:%v", pod.Name),
			fmt.Sprintf("container_name:%v", containerStatus.Name),
		}

		// Determine that there should be a terminate state and that is OOMKilled
		// The container state mapped should be the same as the container we are looking for
		if containerStatus.State.Running == nil &&
			containerStatus.LastTerminationState.Terminated != nil &&
			containerStatus.LastTerminationState.Terminated.Reason == "OOMKilled" {
			// Set the value to 1 as we found a OOM event and break out of the loop as we do not need multiple OOM events
			value = 1
		}

		sender.Gauge("kubernetes.state.container.status.report.count.oom", value, "", tags)
	}
}
