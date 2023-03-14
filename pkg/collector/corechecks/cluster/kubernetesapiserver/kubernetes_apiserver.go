// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package kubernetesapiserver

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"

	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	core "github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
	"github.com/StackVista/stackstate-agent/pkg/util"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/clustername"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// Covers the Control Plane service check and the in memory pod metadata.
const (
	KubeControlPaneCheck         = "kube_apiserver_controlplane.up"
	kubernetesAPIServerCheckName = "kubernetes_apiserver"
	defaultCacheExpire           = 2 * time.Minute
	defaultCachePurge            = 10 * time.Minute

	// Event collection
	eventTokenKey                    = "event"
	maxEventCardinality              = 300
	defaultTimeoutEventCollection    = 2000
	defaultEventResyncPeriodInSecond = 300

	// Pod collection
	podTokenKey                 = "pod"
	maxPodCardinality           = 300
	defaultTimeoutPodCollection = 2000
)

// KubeASConfig is the config of the API server.
type KubeASConfig struct {
	CollectOShiftQuotas bool `yaml:"collect_openshift_clusterquotas"`
	LeaderSkip          bool `yaml:"skip_leader_election"`
	UseComponentStatus  bool `yaml:"use_component_status"`

	// Events
	MaxEventCollection       int      `yaml:"max_events_per_run"`
	FilteredEventTypes       []string `yaml:"filtered_event_types"`
	EventCollectionTimeoutMs int      `yaml:"kubernetes_event_read_timeout_ms"`
	ResyncPeriodEvents       int      `yaml:"kubernetes_event_resync_period_s"`
	CollectEvent             bool     `yaml:"collect_events"`

	// Pods
	MaxPodCollection       int      `yaml:"max_pods_per_run"`
	FilteredPodTypes       []string `yaml:"filtered_pod_types"`
	PodCollectionTimeoutMs int      `yaml:"kubernetes_pod_read_timeout_ms"`
	ResyncPeriodPods       int      `yaml:"kubernetes_pod_resync_period_s"`
	CollectPods            bool     `yaml:"collect_pods"`
}

// EventC holds the information pertaining to which event we collected last and when we last re-synced.
type EventC struct {
	LastResVer string
	LastTime   time.Time
}

// KubeASCheck grabs metrics and events from the API server.
type KubeASCheck struct {
	core.CheckBase
	instance        *KubeASConfig
	eventCollection EventC
	podCollection   EventC
	ignoredEvents   string
	ignoredPods     string
	ac              *apiserver.APIClient
	oshiftAPILevel  apiserver.OpenShiftAPILevel
	providerIDCache *cache.Cache
}

func (c *KubeASConfig) parse(data []byte) error {
	// default values
	c.CollectEvent = config.Datadog.GetBool("collect_kubernetes_events")
	c.CollectPods = config.Datadog.GetBool("collect_kubernetes_pods")
	c.CollectOShiftQuotas = true
	c.ResyncPeriodEvents = defaultEventResyncPeriodInSecond
	c.UseComponentStatus = true

	return yaml.Unmarshal(data, c)
}

func NewKubeASCheck(base core.CheckBase, instance *KubeASConfig) *KubeASCheck {
	return &KubeASCheck{
		CheckBase:       base,
		instance:        instance,
		providerIDCache: cache.New(defaultCacheExpire, defaultCachePurge),
	}
}

// KubernetesASFactory is exported for integration testing.
func KubernetesASFactory() check.Check {
	return NewKubeASCheck(core.NewCheckBase(kubernetesAPIServerCheckName), &KubeASConfig{})
}

// Configure parses the check configuration and init the check.
func (k *KubeASCheck) Configure(config, initConfig integration.Data, source string) error {
	err := k.CommonConfigure(config, source)
	if err != nil {
		return err
	}

	// Check connectivity to the APIServer
	err = k.instance.parse(config)
	if err != nil {
		_ = log.Error("could not parse the config for the API server")
		return err
	}

	// Configure events collection
	k.ignoredEvents = convertFilter(k.instance.FilteredEventTypes)

	if k.instance.EventCollectionTimeoutMs == 0 {
		k.instance.EventCollectionTimeoutMs = defaultTimeoutEventCollection
	}

	if k.instance.MaxEventCollection == 0 {
		k.instance.MaxEventCollection = maxEventCardinality
	}

	// Configure pods collection
	k.ignoredPods = convertFilter(k.instance.FilteredPodTypes)

	if k.instance.PodCollectionTimeoutMs == 0 {
		k.instance.PodCollectionTimeoutMs = defaultTimeoutPodCollection
	}

	if k.instance.MaxPodCollection == 0 {
		k.instance.MaxPodCollection = maxPodCardinality
	}

	return nil
}

func convertFilter(conf []string) string {
	var formatedFilters []string
	for _, filter := range conf {
		f := strings.Split(filter, "=")
		if len(f) == 1 {
			formatedFilters = append(formatedFilters, fmt.Sprintf("reason!=%s", f[0]))
			continue
		}
		formatedFilters = append(formatedFilters, filter)
	}
	return strings.Join(formatedFilters, ",")
}

// Run executes the check.
func (k *KubeASCheck) Run() error {
	log.Infof("Running the cluster agent check.")

	sender, err := aggregator.GetSender(k.ID())
	if err != nil {
		return err
	}
	defer sender.Commit()

	if config.Datadog.GetBool("cluster_agent.enabled") {
		log.Debug("Cluster agent is enabled. Not running Kubernetes API Server check or collecting Kubernetes Events.")
		return nil
	}
	// If the check is configured as a cluster check, the cluster check worker needs to skip the leader election section.
	// The Cluster Agent will passed in the `skip_leader_election` bool.
	if !k.instance.LeaderSkip {
		// Only run if Leader Election is enabled.
		if !config.Datadog.GetBool("leader_election") {
			return log.Error("Leader Election not enabled. Not running Kubernetes API Server check or collecting Kubernetes Events.")
		}
		leader, errLeader := cluster.RunLeaderElection()
		if errLeader != nil {
			if errLeader == apiserver.ErrNotLeader {
				// Only the leader can instantiate the apiserver client.
				log.Debugf("Not leader (leader is %q). Skipping the Kubernetes API Server check", leader)
				return nil
			}

			_ = k.Warn("Leader Election error. Not running the Kubernetes API Server check.")
			return err
		}

		log.Tracef("Current leader: %q, running the Kubernetes API Server check", leader)
	}
	// API Server client initialisation on first run
	if k.ac == nil {
		// Using GetAPIClient (no wait) as check we'll naturally retry with each check run
		k.ac, err = apiserver.GetAPIClient()
		if err != nil {
			k.Warnf("Could not connect to apiserver: %s", err) //nolint:errcheck
			return err
		}

		// We detect OpenShift presence for quota collection
		if k.instance.CollectOShiftQuotas {
			k.oshiftAPILevel = k.ac.DetectOpenShiftAPILevel()
		}
	}

	// Running the Control Plane status check.
	if k.instance.UseComponentStatus {
		err = k.componentStatusCheck(sender)
		if err != nil {
			k.Warnf("Could not collect control plane status from ComponentStatus: %s", err.Error()) //nolint:errcheck
		}
	} else {
		err = k.controlPlaneHealthCheck(context.TODO(), sender)
		if err != nil {
			k.Warnf("Could not collect control plane status from health checks: %s", err.Error()) //nolint:errcheck
		}
	}

	// Running OpenShift ClusterResourceQuota collection if available
	if k.instance.CollectOShiftQuotas && k.oshiftAPILevel != apiserver.NotOpenShift {
		quotas, err := k.retrieveOShiftClusterQuotas()
		if err != nil {
			k.Warnf("Could not collect OpenShift cluster quotas: %s", err.Error()) //nolint:errcheck
		} else {
			k.reportClusterQuotas(quotas, sender)
		}
	}

	// Running the event collection.
	if k.instance.CollectEvent {
		// Get the events from the API server
		events, err := k.eventCollectionCheck()
		if err != nil {
			return err
		}

		// Process the events to have a Datadog format.
		err = k.processEvents(sender, events)
		if err != nil {
			k.Warnf("Could not submit new event %s", err.Error()) //nolint:errcheck
		}
	} else {
		// TODO: Change from warn to debug
		_ = log.Warnf("Not collecting Kubernetes Events")
	}

	// Running the pod collection.
	if k.instance.CollectPods {
		// Get the pods from the API server
		pods, err := k.podCollectionCheck()
		if err != nil {
			return err
		}

		// TODO: Remove logging
		fmt.Printf("collected total pods: %v", len(pods))
		fmt.Printf("collected pods: %v", pods)

	} else {
		// TODO: Change from warn to debug
		_ = log.Warnf("Not collecting Kubernetes Pods")
	}

	return nil
}

func (k *KubeASCheck) eventCollectionCheck() (newEvents []*v1.Event, err error) {
	resVer, lastTime, err := k.ac.GetTokenFromConfigmap(eventTokenKey)
	if err != nil {
		return nil, err
	}

	// This is to avoid getting in a situation where we list all the events for multiple runs in a row.
	if resVer == "" && k.eventCollection.LastResVer != "" {
		log.Errorf("Resource Version stored in the ConfigMap is incorrect. Will resume collecting from %s", k.eventCollection.LastResVer)
		resVer = k.eventCollection.LastResVer
	}

	timeout := int64(k.instance.EventCollectionTimeoutMs / 1000)
	limit := int64(k.instance.MaxEventCollection)
	resync := int64(k.instance.ResyncPeriodEvents)
	newEvents, k.eventCollection.LastResVer, k.eventCollection.LastTime, err = k.ac.RunEventCollection(resVer, lastTime, timeout, limit, resync, k.ignoredEvents)

	if err != nil {
		k.Warnf("Could not collect events from the api server: %s", err.Error()) //nolint:errcheck
		return nil, err
	}

	configMapErr := k.ac.UpdateTokenInConfigmap(eventTokenKey, k.eventCollection.LastResVer, k.eventCollection.LastTime)
	if configMapErr != nil {
		k.Warnf("Could not store the LastEventToken in the ConfigMap: %s", configMapErr.Error()) //nolint:errcheck
	}
	return newEvents, nil
}

func (k *KubeASCheck) podCollectionCheck() (newPods []*v1.Pod, err error) {
	resourceVersion, lastTime, err := k.ac.GetTokenFromConfigmap(podTokenKey)
	if err != nil {
		_ = log.Error("Unable to get the pod token from the ConfigMap")
		return nil, err
	}

	// Avoid getting in a situation where we list all the pods for multiple runs in a row.
	if resourceVersion == "" && k.podCollection.LastResVer != "" {
		_ = log.Errorf("Resource Version stored in the ConfigMap is incorrect. Will resume collecting from %s", k.podCollection.LastResVer)
		resourceVersion = k.podCollection.LastResVer
	}

	timeout := int64(k.instance.PodCollectionTimeoutMs / 1000)
	// TODO: Verify if pod collection limit is needed (Unlike events)
	limit := int64(k.instance.MaxPodCollection)
	resync := int64(k.instance.ResyncPeriodPods)

	// Start collecting the pods
	newPods, k.podCollection.LastResVer, k.podCollection.LastTime, err = k.ac.RunPodCollection(resourceVersion, lastTime, timeout, limit, resync, k.ignoredPods)

	if err != nil {
		_ = k.Warnf("Could not collect pods from the api server: %s", err.Error()) //nolint:errcheck
		return nil, err
	}

	configMapErr := k.ac.UpdateTokenInConfigmap(podTokenKey, k.podCollection.LastResVer, k.podCollection.LastTime)
	if configMapErr != nil {
		_ = k.Warnf("Could not store the LastPodToken in the ConfigMap: %s", configMapErr.Error()) //nolint:errcheck
	}
	return newPods, nil
}

func (k *KubeASCheck) parseComponentStatus(sender aggregator.Sender, componentsStatus *v1.ComponentStatusList) error {
	for _, component := range componentsStatus.Items {
		if component.ObjectMeta.Name == "" {
			return errors.New("metadata structure has changed. Not collecting API Server's Components status")
		}
		if component.Conditions == nil || component.Name == "" {
			log.Debug("API Server component's structure is not expected")
			continue
		}

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
				if message == "" {
					message = condition.Message
				}
			}

			tags := []string{fmt.Sprintf("component:%s", component.Name)}
			sender.ServiceCheck(KubeControlPaneCheck, statusCheck, "", tags, message)
		}
	}
	return nil
}

// processEvents:
// - iterates over the Kubernetes Events
// - extracts some attributes and builds a structure ready to be submitted as a Datadog event (bundle)
// - formats the bundle and submit the Datadog event
func (k *KubeASCheck) processEvents(sender aggregator.Sender, events []*v1.Event) error {
	eventsByObject := make(map[string]*kubernetesEventBundle)

	for _, event := range events {
		id := bundleID(event)
		bundle, found := eventsByObject[id]
		if found == false {
			bundle = newKubernetesEventBundler(event)
			eventsByObject[id] = bundle
		}
		err := bundle.addEvent(event)
		if err != nil {
			k.Warnf("Error while bundling events, %s.", err.Error()) //nolint:errcheck
		}
	}
	hostname, _ := util.GetHostname(context.TODO())
	clusterName := clustername.GetClusterName(context.TODO(), hostname)
	for _, bundle := range eventsByObject {
		datadogEv, err := bundle.formatEvents(clusterName, k.providerIDCache)
		if err != nil {
			k.Warnf("Error while formatting bundled events, %s. Not submitting", err.Error()) //nolint:errcheck
			continue
		}
		sender.Event(datadogEv)
	}
	return nil
}

func (k *KubeASCheck) componentStatusCheck(sender aggregator.Sender) error {
	componentsStatus, err := k.ac.ComponentStatuses()
	if err != nil {
		return err
	}

	return k.parseComponentStatus(sender, componentsStatus)
}

func (k *KubeASCheck) controlPlaneHealthCheck(ctx context.Context, sender aggregator.Sender) error {
	ready, err := k.ac.IsAPIServerReady(ctx)

	var (
		msg    string
		status metrics.ServiceCheckStatus
	)

	if ready {
		msg = "ok"
		status = metrics.ServiceCheckOK
	} else {
		status = metrics.ServiceCheckCritical
		if err != nil {
			msg = err.Error()
		} else {
			msg = "unknown error"
		}
	}

	sender.ServiceCheck(KubeControlPaneCheck, status, "", nil, msg)

	return nil
}

// bundleID generates a unique ID to separate k8s events
// based on their InvolvedObject UIDs and event Types
func bundleID(e *v1.Event) string {
	return fmt.Sprintf("%s/%s", e.InvolvedObject.UID, e.Type)
}

func init() {
	core.RegisterCheck(kubernetesAPIServerCheckName, KubernetesASFactory)
}
