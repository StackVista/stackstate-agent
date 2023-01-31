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

// aggregateStatusReasonMetrics Generate additional metrics based on aggregating existing metrics to generate a new metric
func aggregateStatusReasonMetrics(metricFamilyList []ksmstore.DDMetricsFam) []ksmstore.DDMetricsFam {
	// Split this from the original metrics to merge it back in at the start
	var zeroStateMetrics []ksmstore.DDMetric
	var originalMetrics []ksmstore.DDMetric

	for _, metricFamily := range metricFamilyList {
		// Do not continue of there is no metrics to merge
		if len(metricFamily.ListMetrics) == 0 {
			continue
		}

		switch metricFamily.Name {
		case "kube_pod_container_info":
			var metricWithZeroValue []ksmstore.DDMetric
			for _, metric := range metricFamily.ListMetrics {
				metric.Labels["reason"] = "Unknown"
				// Overwrite the default metric count of 1 with 0 for a zero state
				metric.Val = 0
				metricWithZeroValue = append(metricWithZeroValue, metric)
			}
			zeroStateMetrics = append(zeroStateMetrics, metricWithZeroValue...)

		case "kube_pod_container_status_terminated_reason", "kube_pod_container_status_waiting_reason":
			// Remap all the reason metrics to have a default count of 1
			// This always guarantees a state if it is not a zero state
			for _, metric := range metricFamily.ListMetrics {
				metric.Val = 1
			}

			originalMetrics = append(originalMetrics, metricFamily.ListMetrics...)
		}
	}

	// We are sending both zero state metrics and the original metrics to StackState
	// What this means is that we will be sending a zero and non-zero state for the same metric but expect the agent to de-duplicate
	// the metrics or at least on StackState's aggregation side. If this does become a problem then we need to map and look for every single
	// possible reason type and build up separate metric groupings containing the zero state or the non-zero state, but this will add more weight on this aggregator
	return append(metricFamilyList, ksmstore.DDMetricsFam{
		Name:        "kube_pod_container_status_reasons",
		ListMetrics: append(zeroStateMetrics, originalMetrics...),
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
