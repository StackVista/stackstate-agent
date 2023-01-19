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

// aggregatedStatusMetrics Generate additional metrics based on aggregating existing metrics
func aggregatedStatusReasonMetrics(metricFamilyList []ksmstore.DDMetricsFam) []ksmstore.DDMetricsFam {
	mapMetricFamilies := func(metricFamily ksmstore.DDMetricsFam, accumulator map[string]ksmstore.DDMetric, overwriteValue bool) {
		for _, metric := range metricFamily.ListMetrics {
			// Find the pod id for each metric to match them to the status reason pod
			uid, ok := metric.Labels["uid"]
			if !ok {
				continue
			}

			// If the pod exists then we attempt to combine the labels we found if not then we assign a new map entry
			if pod, ok := accumulator[uid]; !ok {
				// Combine the original labels and the new labels
				combinedLabels := make(map[string]string)
				for key, value := range pod.Labels {
					combinedLabels[key] = value
				}

				if overwriteValue {
					// Use the new value provided by the pod metric
					accumulator[uid] = ksmstore.DDMetric{
						Labels: combinedLabels,
						Val:    metric.Val,
					}
				} else {
					// Use the original value set by the previous dict entry
					accumulator[uid] = ksmstore.DDMetric{
						Labels: combinedLabels,
						Val:    pod.Val,
					}
				}
			} else {
				// Found non-existing pods, Adding it to the dictionary
				// We want all the monitor states to start on a zero value and overwrite that with a state
				accumulator[uid] = ksmstore.DDMetric{
					Labels: metric.Labels,
					Val:    metric.Val,
				}
			}
		}
	}

	// List of pods based on an uid
	podList := make(map[string]ksmstore.DDMetric)

	// Find kube_pod_container_info metrics based on the container_ids
	for _, metricFamily := range metricFamilyList {
		switch metricFamily.Name {
		case "kube_pod_container_status_terminated_reason":
			mapMetricFamilies(metricFamily, podList, true)

		case "kube_pod_container_status_waiting_reason":
			mapMetricFamilies(metricFamily, podList, true)

		case "kube_pod_container_info":
			mapMetricFamilies(metricFamily, podList, false)
		}
	}

	var listMetrics []ksmstore.DDMetric
	for _, value := range podList {
		listMetrics = append(listMetrics, value)
	}

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
