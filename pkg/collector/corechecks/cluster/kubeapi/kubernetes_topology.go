// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver

package kubeapi

import (
	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/version"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	core "github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	collectors "github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/topologycollectors"
	"github.com/DataDog/datadog-agent/pkg/util/features"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

const (
	kubernetesAPITopologyCheckName = "kubernetes_api_topology"
)

// TopologyCheck grabs events from the API server.
type TopologyCheck struct {
	CommonCheck
	instance  *TopologyConfig
	submitter TopologySubmitter
}

func warnDisabledResource(name string, additionalWarning string, isEnabled bool) {
	if !isEnabled {
		if additionalWarning != "" {
			additionalWarning = ": " + additionalWarning
		}
		log.Infof("Collection of %s is disabled%s. "+
			"To enable, set the `clusterAgent.collection.kubernetesResources.%s` setting "+
			"to true in your helm values.yaml file", name, additionalWarning, name)
	}
}

// Configure parses the check configuration and init the check.
func (t *TopologyCheck) Configure(senderManager sender.SenderManager, checkManager handler.CheckManager, integrationConfigDigest uint64, config, initConfig integration.Data, source string) error {
	err := t.ConfigureKubeAPICheck(senderManager, checkManager, integrationConfigDigest, config, initConfig, source)
	if err != nil {
		return err
	}

	err = t.instance.parse(config)
	if err != nil {
		_ = log.Error("could not parse the config for the API topology check")
		return err
	}

	warnDisabledResource("persistentvolumes", "", t.instance.Resources.Persistentvolumes)
	warnDisabledResource("persistentvolumeclaims", "it won't be possible to connect pods to a persistent volumes claimed by them", t.instance.Resources.Persistentvolumeclaims)
	warnDisabledResource("endpoints", "it won't be possible to connect services to underlying pods", t.instance.Resources.Endpoints)
	warnDisabledResource("namespaces", "", t.instance.Resources.Namespaces)
	warnDisabledResource("configmaps", "", t.instance.Resources.ConfigMaps)
	warnDisabledResource("daemonsets", "", t.instance.Resources.Daemonsets)
	warnDisabledResource("deployments", "", t.instance.Resources.Deployments)
	warnDisabledResource("replicasets", "", t.instance.Resources.Replicasets)
	warnDisabledResource("statefulsets", "", t.instance.Resources.Statefulsets)
	warnDisabledResource("ingresses", "", t.instance.Resources.Ingresses)
	warnDisabledResource("jobs", "", t.instance.Resources.Jobs)
	warnDisabledResource("cronjobs", "", t.instance.Resources.CronJobs)
	warnDisabledResource("secrets", "", t.instance.Resources.Secrets)

	log.Debugf("Running config %s", config)
	return nil
}

func (t *TopologyCheck) getKubernetesVersion() *version.Info {
	info, err := t.ac.GetVersion()
	if err != nil {
		_ = log.Warnf("Could not set Kubernetes version for topology collector: ", err)
	}
	log.Debugf("Kubernetes version: %+v", info)
	return info
}

/*
	cluster -> map cluster -> component

	node -> map node -> component
				     -> cluster relation
		   component <- container correlator
			relation <-

	pod -> map pod 	  		 -> component
							 -> node relation
		container correlator <- map func container -> component
												   -> relation

	service -> map service -> component
						   -> endpoints as identifiers
						   -> pod relation

	component -> publish component
	relation -> publish relation
*/
// Run executes the check.
func (t *TopologyCheck) Run() error {
	// Running the event collection.
	if !t.instance.CollectTopology {
		return nil
	}

	// initialize kube api check
	err := t.InitKubeAPICheck()
	if err != nil {
		return err
	}

	// set the check "instance id" for snapshots
	t.instance.CheckID = kubernetesAPITopologyCheckName
	t.instance.Instance = topology.Instance{Type: string(collectors.Kubernetes), URL: t.instance.ClusterName}

	// set up the batcher for this instance
	t.submitter = NewBatchTopologySubmitter(t.GetCheckHandler(), t.instance.Instance)

	// start the topology snapshot with the batch-er
	t.submitter.SubmitStartSnapshot()

	// create a wait group for all the collectors
	var waitGroup sync.WaitGroup

	// Make a channel for each of the relations to avoid passing data down into all the functions
	nodeIdentifierCorrelationChannel := make(chan *collectors.NodeIdentifierCorrelation)
	containerCorrelationChannel := make(chan *collectors.ContainerCorrelation)
	volumeCorrelationChannel := make(chan *collectors.VolumeCorrelation)
	podCorrelationChannel := make(chan *collectors.PodLabelCorrelation)
	endpointCorrelationChannel := make(chan *collectors.ServiceSelectorCorrelation)

	// make a channel that is responsible for publishing components and relations
	componentChannel := make(chan *topology.Component)
	relationChannel := make(chan *topology.Relation)
	errChannel := make(chan error)
	waitGroupChannel := make(chan bool)
	collectorsDoneChannel := make(chan bool)

	var instanceClusterType collectors.ClusterType
	switch openshiftPresence := t.ac.DetectOpenShiftAPILevel(); openshiftPresence {
	case apiserver.OpenShiftAPIGroup, apiserver.OpenShiftOAPI:
		instanceClusterType = collectors.OpenShift
	case apiserver.NotOpenShift:
		instanceClusterType = collectors.Kubernetes
	}
	clusterTopologyCommon := collectors.NewClusterTopologyCommon(t.instance.Instance, instanceClusterType, t.ac, t.instance.SourcePropertiesEnabled, componentChannel, relationChannel, t.getKubernetesVersion(), t.GetFeatures().FeatureEnabled(features.ExposeKubernetesStatus))
	commonClusterCollector := collectors.NewClusterTopologyCollector(clusterTopologyCommon)
	clusterCollectors := []collectors.ClusterTopologyCollector{
		// Register Cluster Component Collector
		collectors.NewClusterCollector(
			commonClusterCollector,
		),
		// Register Node Component Collector
		collectors.NewNodeCollector(
			nodeIdentifierCorrelationChannel,
			commonClusterCollector,
		),
		// Register Pod Component Collector
		collectors.NewPodCollector(
			containerCorrelationChannel,
			volumeCorrelationChannel,
			podCorrelationChannel,
			commonClusterCollector,
		),
		// Register Service Component Collector
		collectors.NewServiceCollector(
			endpointCorrelationChannel,
			commonClusterCollector,
		),
	}

	if t.instance.Resources.Persistentvolumes {
		clusterCollectors = append(clusterCollectors,
			collectors.NewPersistentVolumeCollector(
				commonClusterCollector,
				t.instance.CSIPVMapperEnabled,
			))
	}

	if t.instance.Resources.Namespaces {
		clusterCollectors = append(clusterCollectors,
			collectors.NewNamespaceCollector(
				commonClusterCollector,
			))
	}

	if t.instance.Resources.ConfigMaps {
		clusterCollectors = append(clusterCollectors,
			collectors.NewConfigMapCollector(
				commonClusterCollector,
				t.instance.ConfigMapMaxDataSize,
			))
	}

	if t.instance.Resources.Secrets {
		clusterCollectors = append(clusterCollectors,
			collectors.NewSecretCollector(
				commonClusterCollector,
			))
	}

	if t.instance.Resources.Daemonsets {
		clusterCollectors = append(clusterCollectors,
			collectors.NewDaemonSetCollector(
				commonClusterCollector,
			))
	}
	if t.instance.Resources.Deployments {
		clusterCollectors = append(clusterCollectors,
			collectors.NewDeploymentCollector(
				commonClusterCollector,
			))
	}
	if t.instance.Resources.Replicasets {
		clusterCollectors = append(clusterCollectors,
			collectors.NewReplicaSetCollector(
				commonClusterCollector,
			))
	}
	if t.instance.Resources.Statefulsets {
		clusterCollectors = append(clusterCollectors,
			collectors.NewStatefulSetCollector(
				commonClusterCollector,
			))
	}
	if t.instance.Resources.Ingresses {
		clusterCollectors = append(clusterCollectors,
			collectors.NewIngressCollector(
				commonClusterCollector,
			))
	}
	if t.instance.Resources.Jobs {
		clusterCollectors = append(clusterCollectors,
			collectors.NewJobCollector(
				commonClusterCollector,
			))
	}
	if t.instance.Resources.CronJobs {
		clusterCollectors = append(clusterCollectors,
			collectors.NewCronJobCollector(
				commonClusterCollector,
			))
	}

	commonClusterCorrelator := collectors.NewClusterTopologyCorrelator(clusterTopologyCommon)
	clusterCorrelators := []collectors.ClusterTopologyCorrelator{
		// Register Container -> Node Identifier Correlator
		collectors.NewContainerCorrelator(
			nodeIdentifierCorrelationChannel,
			containerCorrelationChannel,
			commonClusterCorrelator,
		),
		collectors.NewVolumeCorrelator(
			volumeCorrelationChannel,
			commonClusterCorrelator,
			t.instance.Resources.Persistentvolumeclaims,
		),
		collectors.NewService2PodCorrelator(
			podCorrelationChannel,
			endpointCorrelationChannel,
			commonClusterCorrelator,
		),
	}

	// starts all the cluster collectors and correlators
	t.RunClusterCollectors(clusterCollectors, clusterCorrelators, &waitGroup, errChannel, commonClusterCollector, collectorsDoneChannel)

	// receive all the components, will return once the wait group notifies
	t.WaitForTopology(componentChannel, relationChannel, errChannel, &waitGroup, waitGroupChannel)

	t.submitter.SubmitStopSnapshot()
	t.submitter.SubmitComplete()

	log.Infof("Topology Check for cluster: %s completed successfully", t.instance.ClusterName)
	// close all the created channels
	close(componentChannel)
	close(relationChannel)
	close(errChannel)
	close(waitGroupChannel)
	close(collectorsDoneChannel)

	return nil
}

// WaitForTopology sets up the receiver that handles the component and relation channel and publishes it to StackState, returns when all the collectors have finished or the timeout was reached.
func (t *TopologyCheck) WaitForTopology(componentChannel <-chan *topology.Component, relationChannel <-chan *topology.Relation,
	errorChannel <-chan error, waitGroup *sync.WaitGroup, waitGroupChannel chan bool) {
	log.Debugf("Waiting for Cluster Collectors to Finish")
	go func() {
	loop:
		for {
			select {
			case component := <-componentChannel:
				t.submitter.SubmitComponent(component)
			case relation := <-relationChannel:
				t.submitter.SubmitRelation(relation)
			case err := <-errorChannel:
				t.submitter.HandleError(err)
			case timedOut := <-waitGroupChannel:
				if timedOut {
					_ = log.Warn("WaitGroup for Cluster Collectors did not finish in time, stopping topology publish loop")
				} else {
					log.Debug("All collectors have been finished their work, continuing to publish data to StackState")
				}
				break loop // timed out
			default:
				// no message received, continue looping
			}
		}
	}()

	timeout := time.Duration(t.instance.CollectTimeout) * time.Minute
	log.Debugf("Waiting for Cluster Collectors to Finish")
	waitGroupChannel <- waitTimeout(waitGroup, timeout)
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	wgChan := make(chan struct{})
	go func() {
		defer close(wgChan)
		wg.Wait()
	}()
	select {
	case <-wgChan:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// RunClusterCollectors runs all the cluster collectors, notify the wait groups and submit errors to the error channel
func (t *TopologyCheck) RunClusterCollectors(
	clusterCollectors []collectors.ClusterTopologyCollector,
	clusterCorrelators []collectors.ClusterTopologyCorrelator,
	waitGroup *sync.WaitGroup,
	errorChannel chan<- error,
	commonCollector collectors.ClusterTopologyCommon,
	collectorsDoneChannel chan bool) {
	var collectorsWaitGroup sync.WaitGroup
	collectorsWaitGroup.Add(1 + len(clusterCorrelators)) // all collectors + all correlators (without RelationCorrelator)
	go func() {
		for _, collector := range clusterCollectors {
			// add this collector to the wait group
			runCollector(collector, errorChannel)
		}
		collectorsWaitGroup.Done()
	}()
	// Run all correlators in parallel to avoid blocking channels
	for _, correlator := range clusterCorrelators {
		go runCorrelator(correlator, errorChannel, &collectorsWaitGroup)
	}
	go func() {
		collectorsWaitGroup.Wait()
		collectorsDoneChannel <- true
	}()
	waitGroup.Add(1)
	go func() {
		<-collectorsDoneChannel
		commonCollector.CorrelateRelations()
		waitGroup.Done()
	}()
}

// runCollector
func runCollector(collector collectors.ClusterTopologyCollector, errorChannel chan<- error) {
	log.Debugf("Starting cluster topology collector: %s\n", collector.GetName())
	err := collector.CollectorFunction()
	if err != nil {
		errorChannel <- err
	}
	// mark this collector as complete
	log.Debugf("Finished cluster topology collector: %s\n", collector.GetName())
}

// runCorrelator
func runCorrelator(correlator collectors.ClusterTopologyCorrelator, errorChannel chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Debugf("Starting cluster topology correlator: %s\n", correlator.GetName())
	err := correlator.CorrelateFunction()
	if err != nil {
		errorChannel <- err
	}
	log.Debugf("Finished cluster topology correlator: %s\n", correlator.GetName())
}

// KubernetesAPITopologyFactory is exported for integration testing.
func KubernetesAPITopologyFactory() check.Check {
	return &TopologyCheck{
		CommonCheck: CommonCheck{
			CheckBase: core.NewCheckBase(kubernetesAPITopologyCheckName),
		},
		instance: &TopologyConfig{},
	}
}

func init() {
	core.RegisterCheck(kubernetesAPITopologyCheckName, KubernetesAPITopologyFactory)
}
