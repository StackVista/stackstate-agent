// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package kubeapi

import (
	"context"
	"errors"
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/clustername"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/api/core/v1"

	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	core "github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	"github.com/DataDog/datadog-agent/pkg/metrics"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// Covers the Control Plane service check and the in memory pod metadata.
const (
	KubeControlPaneCheck          = "kube_apiserver_controlplane.up"
	kubernetesAPIMetricsCheckName = "kubernetes_api_metrics"
)

// MetricsConfig
type MetricsConfig struct {
	CollectMetrics      bool `yaml:"collect_metrics"`
	CollectOShiftQuotas bool `yaml:"collect_openshift_clusterquotas"`

	// Pod Metrics
	ResyncPeriodPodMetric        int `yaml:"kubernetes_pod_metric_resync_period_s"`
	PodMetricCollectionTimeoutMs int `yaml:"kubernetes_pod_metric_read_timeout_ms"`
	MaxPodMetricsCollection      int `yaml:"max_custom_pod_metrics_per_run"`
}

// MetricC holds the information pertaining to which metric we collected last and when we last re-synced.
type MetricC struct {
	LastResVer string
	LastTime   time.Time
}

// MetricsCheck grabs metrics from the API server.
type MetricsCheck struct {
	CommonCheck
	podMetricCollection MetricC
	instance            *MetricsConfig
	oshiftAPILevel      apiserver.OpenShiftAPILevel
	clusterName         string
}

func (c *MetricsConfig) parse(data []byte) error {
	// default values
	c.CollectMetrics = config.Datadog.GetBool("collect_kubernetes_metrics")
	c.CollectOShiftQuotas = true
	c.parsePodMetrics()

	return yaml.Unmarshal(data, c)
}

// NewKubernetesAPIMetricsCheck creates a instance of the kubernetes MetricsCheck given the base and instance
func NewKubernetesAPIMetricsCheck(base core.CheckBase, instance *MetricsConfig) *MetricsCheck {
	return &MetricsCheck{
		CommonCheck: CommonCheck{
			CheckBase: base,
		},
		instance: instance,
	}
}

// KubernetesAPIMetricsFactory is exported for integration testing.
func KubernetesAPIMetricsFactory() check.Check {
	return NewKubernetesAPIMetricsCheck(core.NewCheckBase(kubernetesAPIMetricsCheckName), &MetricsConfig{})
}

// getClusterName retrieves the name of the cluster, if found
func (k *MetricsCheck) getClusterName() {
	hostname, _ := util.GetHostname(context.TODO())
	if clusterName := clustername.GetClusterName(context.TODO(), hostname); clusterName != "" {
		k.clusterName = clusterName
	}
}

// Configure parses the check configuration and init the check.
func (k *MetricsCheck) Configure(config, initConfig integration.Data, source string) error {
	err := k.CommonConfigure(config, source)
	if err != nil {
		return err
	}

	// Check connectivity to the APIServer
	err = k.instance.parse(config)
	if err != nil {
		_ = log.Error("could not parse the config for the API metrics check")
		return err
	}

	k.setDefaultsPodEvents()
	k.getClusterName()

	log.Debugf("Running config %s", config)
	return nil
}

// Run executes the check.
func (k *MetricsCheck) Run() error {
	// Running the metric collection.
	if !k.instance.CollectMetrics {
		return nil
	}

	// initialize kube api check
	err := k.InitKubeAPICheck()
	if err == apiserver.ErrNotLeader {
		log.Debug("Agent is not leader, will not run the check")
		return nil
	} else if err != nil {
		return err
	}

	sender, err := aggregator.GetSender(k.ID())
	if err != nil {
		return err
	}
	defer sender.Commit()

	// Running the Control Plane status check.
	componentsStatus, err := k.ac.ComponentStatuses()
	if err != nil {
		_ = k.Warnf("Could not retrieve the status from the control plane's components %s", err.Error())
	} else {
		err = k.parseComponentStatus(sender, componentsStatus)
		if err != nil {
			_ = k.Warnf("Could not collect API Server component status: %s", err.Error())
		}
	}

	log.Info("Running kubernetes pod metric collector ...")
	// Get pods from the API server to produce custom metrics
	pods, podMetricErr := k.podMetricsCollectionCheck()
	if podMetricErr == nil {
		k.processPods(sender, pods)
	}

	// Running OpenShift ClusterResourceQuota collection if available
	if k.instance.CollectOShiftQuotas && k.oshiftAPILevel != apiserver.NotOpenShift {
		quotas, err := k.retrieveOShiftClusterQuotas()
		if err != nil {
			// [STS] log this as a debug message instead. TODO: make k.instance.CollectOShiftQuotas con
			log.Debugf("Could not collect OpenShift cluster quotas: %s", err.Error())
		} else {
			k.reportClusterQuotas(quotas, sender)
		}
	}

	return nil
}

func (k *MetricsCheck) parseComponentStatus(sender aggregator.Sender, componentsStatus *v1.ComponentStatusList) error {
	for _, component := range componentsStatus.Items {

		if component.ObjectMeta.Name == "" {
			return errors.New("metadata structure has changed. Not collecting API Server's Components status")
		}
		if component.Conditions == nil || component.Name == "" {
			log.Debug("API Server component's structure is not expected")
			continue
		}
		tagComp := []string{fmt.Sprintf("component:%s", component.Name)}
		for _, condition := range component.Conditions {
			statusCheck := metrics.ServiceCheckUnknown
			message := ""

			// We only expect the Healthy condition. May change in the future. https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#typical-status-properties
			if condition.Type != "Healthy" {
				log.Debugf("Condition %q not supported", condition.Type)
				continue
			}
			// We only expect True, False and Unknown (default).
			switch condition.Status {
			case "True":
				statusCheck = metrics.ServiceCheckOK
				message = condition.Message
			case "False":
				statusCheck = metrics.ServiceCheckCritical
				message = condition.Error
			}
			sender.ServiceCheck(KubeControlPaneCheck, statusCheck, k.KubeAPIServerHostname, tagComp, message)
		}
	}
	return nil
}

func init() {
	core.RegisterCheck(kubernetesAPIMetricsCheckName, KubernetesAPIMetricsFactory)
}
