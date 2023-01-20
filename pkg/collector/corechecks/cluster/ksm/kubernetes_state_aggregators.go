// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2021-present Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package ksm

import (
	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	ksmstore "github.com/StackVista/stackstate-agent/pkg/kubestatemetrics/store"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

type metricAggregator interface {
	accumulate(ksmstore.DDMetric)
	flush(aggregator.Sender, *KSMCheck, *labelJoiner)
}

// maxNumberOfAllowedLabels contains the maximum number of labels that can be used to aggregate metrics.
// The only reason why there is a maximum is because the `accumulator` map is indexed on the label values
// and GO accepts arrays as valid map key type, but not slices.
// This hard coded limit is fine because the metrics to aggregate and the label list to use are hardcoded
// in the code and cannot be arbitrarily set by the end-user.
const maxNumberOfAllowedLabels = 4

type counterAggregator struct {
	ddMetricName  string
	ksmMetricName string
	allowedLabels []string

	accumulator map[[maxNumberOfAllowedLabels]string]float64
}

type sumValuesAggregator struct {
	counterAggregator
}

type countObjectsAggregator struct {
	counterAggregator
}

func newSumValuesAggregator(ddMetricName, ksmMetricName string, allowedLabels []string) metricAggregator {
	if len(allowedLabels) > maxNumberOfAllowedLabels {
		// `maxNumberOfAllowedLabels` is hardcoded to the maximum number of labels passed to this function from the metricsAggregators definition below.
		// The only possibility to arrive here is to add a new aggregator in `metricAggregator` below and to forget to update `maxNumberOfAllowedLabels` accordingly.
		log.Error("BUG in KSM metric aggregator")
		return nil
	}

	return &sumValuesAggregator{
		counterAggregator{
			ddMetricName:  ddMetricName,
			ksmMetricName: ksmMetricName,
			allowedLabels: allowedLabels,
			accumulator:   make(map[[maxNumberOfAllowedLabels]string]float64),
		},
	}
}

// aggregatedStatusMetrics Generate additional metrics based on aggregating existing metrics to generate a new metric
func aggregatedStatusReasonMetrics(metricFamilyList []ksmstore.DDMetricsFam) []ksmstore.DDMetricsFam {
	// Attempt to merge a metric family to a exsting dictionary
	metricFamilyMerge := func(metricFamily ksmstore.DDMetricsFam, accumulator map[string]ksmstore.DDMetric, isZeroValue bool) {
		for _, metric := range metricFamily.ListMetrics {
			// Verify that UID exists as this will be used when merging status results
			// Also verify that there is Labels available otherwise there is no reason to merge
			uid, ok := metric.Labels["uid"]
			if !ok || len(metric.Labels) <= 0 {
				continue
			}

			// If the dictionary already contains the entry then we want to merge the existing data
			if pod, ok := accumulator[uid]; !ok && len(pod.Labels) > 0 {
				// Combine the original labels and the new labels
				labels := make(map[string]string)
				for key, value := range pod.Labels {
					labels[key] = value
				}

				// Use the original value if is zero value
				if isZeroValue {
					accumulator[uid] = ksmstore.DDMetric{
						Labels: labels,
						Val:    pod.Val,
					}
				} else {
					accumulator[uid] = ksmstore.DDMetric{
						Labels: labels,
						Val:    metric.Val,
					}
				}
			} else {
				// Adding a default reason if there is no reason, This allows us to have a zero state
				// There should be no assumption on what the default state is, and we should keep it unknown
				labels := map[string]string{
					"reason": "Unknown",
				}

				// Combine the original labels with the default
				for key, value := range metric.Labels {
					labels[key] = value
				}

				// Found non-existing pods, Adding it to the dictionary
				if isZeroValue {
					accumulator[uid] = ksmstore.DDMetric{
						Labels: labels,
						Val:    0,
					}
				} else {
					accumulator[uid] = ksmstore.DDMetric{
						Labels: labels,
						Val:    metric.Val,
					}
				}
			}
		}
	}

	// List of pods based on an uid
	podList := make(map[string]ksmstore.DDMetric)

	for _, metricFamily := range metricFamilyList {
		// Only attempt a mapping if there is actual metrics in the metric family
		if len(metricFamily.ListMetrics) == 0 {
			continue
		}

		// Attempt to merge the metric with a pre-existing dictionary of metrics
		switch metricFamily.Name {
		// This is the core metric used to create a zero state - By default the reason will be Unknown
		case "kube_pod_container_info":
			metricFamilyMerge(metricFamily, podList, true)
		// This metric will overwrite a zero state if there is a reason
		case "kube_pod_container_status_terminated_reason":
			metricFamilyMerge(metricFamily, podList, false)
		// This metric will overwrite a zero state if there is a reason
		case "kube_pod_container_status_waiting_reason":
			metricFamilyMerge(metricFamily, podList, false)
		}
	}

	// If there is no results to create the new metric with then return the original metric family
	if len(podList) == 0 {
		return metricFamilyList
	}

	// Convert dictionary to a list to create a family metric
	var listMetrics []ksmstore.DDMetric
	for _, value := range podList {
		listMetrics = append(listMetrics, value)
	}

	// Combine the original metric families and the new metric list
	return append(metricFamilyList, ksmstore.DDMetricsFam{
		Name:        "kube_pod_container_status",
		ListMetrics: listMetrics,
	})
}

func newCountObjectsAggregator(ddMetricName, ksmMetricName string, allowedLabels []string) metricAggregator {
	if len(allowedLabels) > maxNumberOfAllowedLabels {
		// `maxNumberOfAllowedLabels` is hardcoded to the maximum number of labels passed to this function from the metricsAggregators definition below.
		// The only possibility to arrive here is to add a new aggregator in `metricAggregator` below and to forget to update `maxNumberOfAllowedLabels` accordingly.
		log.Error("BUG in KSM metric aggregator")
		return nil
	}

	return &countObjectsAggregator{
		counterAggregator{
			ddMetricName:  ddMetricName,
			ksmMetricName: ksmMetricName,
			allowedLabels: allowedLabels,
			accumulator:   make(map[[maxNumberOfAllowedLabels]string]float64),
		},
	}
}

func (a *sumValuesAggregator) accumulate(metric ksmstore.DDMetric) {
	var labelValues [maxNumberOfAllowedLabels]string

	for i, allowedLabel := range a.allowedLabels {
		if allowedLabel == "" {
			break
		}

		labelValues[i] = metric.Labels[allowedLabel]
	}

	a.accumulator[labelValues] += metric.Val
}

func (a *countObjectsAggregator) accumulate(metric ksmstore.DDMetric) {
	var labelValues [maxNumberOfAllowedLabels]string

	for i, allowedLabel := range a.allowedLabels {
		if allowedLabel == "" {
			break
		}

		labelValues[i] = metric.Labels[allowedLabel]
	}

	a.accumulator[labelValues]++
}

func (a *counterAggregator) flush(sender aggregator.Sender, k *KSMCheck, labelJoiner *labelJoiner) {
	for labelValues, count := range a.accumulator {

		labels := make(map[string]string)
		for i, allowedLabel := range a.allowedLabels {
			if allowedLabel == "" {
				break
			}

			labels[allowedLabel] = labelValues[i]
		}

		hostname, tags := k.hostnameAndTags(labels, labelJoiner, labelsMapperOverride(a.ksmMetricName))

		sender.Gauge(ksmMetricPrefix+a.ddMetricName, count, hostname, tags)
	}

	a.accumulator = make(map[[maxNumberOfAllowedLabels]string]float64)
}

var metricAggregators = map[string]metricAggregator{
	"kube_persistentvolume_status_phase": newSumValuesAggregator(
		"persistentvolumes.by_phase",
		"kube_persistentvolume_status_phase",
		[]string{"storageclass", "phase"},
	),
	"kube_service_spec_type": newCountObjectsAggregator(
		"service.count",
		"kube_service_spec_type",
		[]string{"namespace", "type"},
	),
	"kube_namespace_status_phase": newSumValuesAggregator(
		"namespace.count",
		"kube_namespace_status_phase",
		[]string{"phase"},
	),
	"kube_replicaset_owner": newCountObjectsAggregator(
		"replicaset.count",
		"kube_replicaset_owner",
		[]string{"namespace", "owner_name", "owner_kind"},
	),
	"kube_job_owner": newCountObjectsAggregator(
		"job.count",
		"kube_job_owner",
		[]string{"namespace", "owner_name", "owner_kind"},
	),
	"kube_deployment_labels": newCountObjectsAggregator(
		"deployment.count",
		"kube_deployment_labels",
		[]string{"namespace"},
	),
	"kube_daemonset_labels": newCountObjectsAggregator(
		"daemonset.count",
		"kube_daemonset_labels",
		[]string{"namespace"},
	),
	"kube_statefulset_labels": newCountObjectsAggregator(
		"statefulset.count",
		"kube_statefulset_labels",
		[]string{"namespace"},
	),
	"kube_cronjob_labels": newCountObjectsAggregator(
		"cronjob.count",
		"kube_cronjob_labels",
		[]string{"namespace"},
	),
	"kube_endpoint_labels": newCountObjectsAggregator(
		"endpoint.count",
		"kube_endpoint_labels",
		[]string{"namespace"},
	),
	"kube_horizontalpodautoscaler_labels": newCountObjectsAggregator(
		"hpa.count",
		"kube_horizontalpodautoscaler_labels",
		[]string{"namespace"},
	),
	"kube_verticalpodautoscaler_labels": newCountObjectsAggregator(
		"vpa.count",
		"kube_verticalpodautoscaler_labels",
		[]string{"namespace"},
	),
	"kube_node_info": newCountObjectsAggregator(
		"node.count",
		"kube_node_info",
		[]string{"kubelet_version", "container_runtime_version", "kernel_version", "os_image"},
	),
	"kube_pod_info": newCountObjectsAggregator(
		"pod.count",
		"kube_pod_info",
		[]string{"node", "namespace", "created_by_kind", "created_by_name"},
	),
}
