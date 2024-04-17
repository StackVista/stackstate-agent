// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver

package kubeapi

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/batcher"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	collectors "github.com/DataDog/datadog-agent/pkg/collector/corechecks/cluster/topologycollectors"
	agentConfig "github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/features"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/version"
)

var componentID int
var relationID int

var optionalRules = []string{
	"namespaces+get,list,watch",
	"configmaps+list,watch", // get is a required permission
	"persistentvolumeclaims+get,list,watch",
	"persistentvolumes+get,list,watch",
	"secrets+get,list,watch",
	"apps/daemonsets+get,list,watch",
	"apps/deployments+get,list,watch",
	"apps/replicasets+get,list,watch",
	"apps/statefulsets+get,list,watch",
	"extensions/ingresses+get,list,watch",
	"batch/cronjobs+get,list,watch",
	"batch/jobs+get,list,watch",
}

func TestDisablingAnyResourceWithoutDisablingCollectorCauseAnError(t *testing.T) {
	for _, rule := range optionalRules {
		mBatcher := batcher.NewMockBatcher()
		check := KubernetesAPITopologyFactory().(*TopologyCheck)
		check.ac = MockAPIClient([]Rule{parseRule(rule)})

		nothingIsDisabledConfig := `
cluster_name: mycluster
collect_topology: true
csi_pv_mapper_enabled: true
`
		err := check.Configure([]byte(nothingIsDisabledConfig), nil, "")
		check.SetFeatures(features.All())
		assert.NoError(t, err)

		err = check.Run()
		assert.NoError(t, err, "check itself should succeed despite failures of a particular collector")

		assert.NotEmpty(t, mBatcher.Errors, "Disabling %v should cause an error", rule)
	}
}

func TestDisablingAllPossibleCollectorsKeepErrorsOff(t *testing.T) {
	mBatcher := batcher.NewMockBatcher()
	check := KubernetesAPITopologyFactory().(*TopologyCheck)
	check.ac = MockAPIClient(parseRules(optionalRules))
	allResourcesAreDisabledConfig := `
cluster_name: mycluster
collect_topology: true
csi_pv_mapper_enabled: true
resources:
  persistentvolumes: false
  persistentvolumeclaims: false
  endpoints: false
  namespaces: false
  configmaps: false
  daemonsets: false
  deployments: false
  replicasets: false
  statefulsets: false
  ingresses: false
  jobs: false
  cronjobs: false
  secrets: false
`
	err := check.Configure([]byte(allResourcesAreDisabledConfig), nil, "")
	check.SetFeatures(features.All())
	assert.NoError(t, err)

	err = check.Run()
	assert.NoError(t, err, "check should succeed")

	assert.Empty(t, mBatcher.Errors, "No errors are expected because all resources are disabled in config")
}

func TestRunClusterCollectors(t *testing.T) {
	t.Run("with sourceProperties enabled", func(t *testing.T) {
		testRunClusterCollectors(t, true, true)
	})
	t.Run("with sourceProperties enabled", func(t *testing.T) {
		testRunClusterCollectors(t, true, false)
	})
	t.Run("with sourceProperties disabled", func(t *testing.T) {
		testRunClusterCollectors(t, false, false)
	})
	t.Run("with sourceProperties disabled", func(t *testing.T) {
		testRunClusterCollectors(t, false, true)
	})
}

func testConfigParsed(t *testing.T, input string, expected TopologyConfig) {
	check := KubernetesAPITopologyFactory().(*TopologyCheck)
	err := check.Configure([]byte(input), []byte(""), "whatever")
	assert.NoError(t, err)
	assert.EqualValues(t, &expected, check.instance)
}

func TestConfigurationParsing(t *testing.T) {
	defaultConfig := TopologyConfig{
		// for empty config something is coming from global configuration
		ClusterName:             agentConfig.Datadog.GetString("cluster_name"),
		CollectTopology:         agentConfig.Datadog.GetBool("collect_kubernetes_topology"),
		CollectTimeout:          agentConfig.Datadog.GetInt("collect_kubernetes_timeout"),
		SourcePropertiesEnabled: agentConfig.Datadog.GetBool("kubernetes_source_properties_enabled"),
		ConfigMapMaxDataSize:    DefaultConfigMapDataSizeLimit,
		CSIPVMapperEnabled:      agentConfig.Datadog.GetBool("kubernetes_csi_pv_mapper_enabled"),
		Resources: ResourcesConfig{
			Persistentvolumes:      true,
			Persistentvolumeclaims: true,
			Endpoints:              true,
			Namespaces:             true,
			ConfigMaps:             true,
			Daemonsets:             true,
			Deployments:            true,
			Replicasets:            true,
			Statefulsets:           true,
			Ingresses:              true,
			Jobs:                   true,
			CronJobs:               true,
			Secrets:                true,
		},
	}
	testConfigParsed(t, "", defaultConfig)

	allResourcesAreDisabledConfig := `
cluster_name: mycluster
source_properties_enabled: false
resources:
  persistentvolumes: false
  persistentvolumeclaims: false
  endpoints: false
  namespaces: false
  configmaps: false
  daemonsets: false
  deployments: false
  replicasets: false
  statefulsets: false
  ingresses: false
  jobs: false
  cronjobs: false
  secrets: false
`
	expectedSimple := defaultConfig
	expectedSimple.ClusterName = "mycluster"
	expectedSimple.SourcePropertiesEnabled = false
	expectedSimple.Resources = ResourcesConfig{}
	testConfigParsed(t, allResourcesAreDisabledConfig, expectedSimple)
}

func testRunClusterCollectors(t *testing.T, sourceProperties bool, exposeKubernetesStatus bool) {
	// set the initial id values
	componentID = 1
	relationID = 1

	kubernetesTopologyCheck := KubernetesAPITopologyFactory().(*TopologyCheck)
	instance := topology.Instance{Type: "kubernetes", URL: "test-cluster-name"}
	clusterType := collectors.Kubernetes
	// set up the batcher for this instance
	kubernetesTopologyCheck.instance.CollectTimeout = 5
	kubernetesTopologyCheck.submitter = NewTestTopologySubmitter(t, "kubernetes_api_topology", instance)

	var waitGroup sync.WaitGroup
	componentChannel := make(chan *topology.Component)
	relationChannel := make(chan *topology.Relation)
	errChannel := make(chan error)
	waitGroupChannel := make(chan bool)
	collectorsDoneChannel := make(chan bool)

	clusterTopologyCommon := collectors.NewClusterTopologyCommon(instance, clusterType, nil, sourceProperties, componentChannel, relationChannel, &version.Info{Major: "1", Minor: "21"}, exposeKubernetesStatus)
	commonClusterCollector := collectors.NewClusterTopologyCollector(clusterTopologyCommon)

	clusterCollectors := []collectors.ClusterTopologyCollector{
		NewTestCollector(componentChannel, relationChannel, commonClusterCollector),
		NewErrorTestCollector(componentChannel, relationChannel, commonClusterCollector),
	}
	clusterCorrelators := make([]collectors.ClusterTopologyCorrelator, 0)

	// starts all the cluster collectors
	kubernetesTopologyCheck.RunClusterCollectors(clusterCollectors, clusterCorrelators, &waitGroup, errChannel, commonClusterCollector, collectorsDoneChannel)

	// receive all the components, will return once the wait group notifies
	kubernetesTopologyCheck.WaitForTopology(componentChannel, relationChannel, errChannel, &waitGroup, waitGroupChannel)

	close(componentChannel)
	close(relationChannel)
	close(errChannel)
	close(waitGroupChannel)
	close(collectorsDoneChannel)
}

// NewTestTopologySubmitter creates a new instance of TestTopologySubmitter
func NewTestTopologySubmitter(t *testing.T, checkID check.ID, instance topology.Instance) TopologySubmitter {
	return &TestTopologySubmitter{
		t:        t,
		CheckID:  checkID,
		Instance: instance,
	}
}

// TestTopologySubmitter provides functionality to submit topology data with the Batcher.
type TestTopologySubmitter struct {
	t        *testing.T
	CheckID  check.ID
	Instance topology.Instance
}

func (b *TestTopologySubmitter) SubmitStartSnapshot() {}
func (b *TestTopologySubmitter) SubmitStopSnapshot()  {}
func (b *TestTopologySubmitter) SubmitComplete()      {}

// SubmitRelation takes a component and submits it with the Batcher
func (b *TestTopologySubmitter) SubmitComponent(component *topology.Component) {
	// match the component with the count number that represents the ExternalID
	assert.Equal(b.t, strconv.Itoa(componentID), component.ExternalID)
	componentID = componentID + 1
}

// SubmitRelation takes a relation and submits it with the Batcher
func (b *TestTopologySubmitter) SubmitRelation(relation *topology.Relation) {
	// match the relation with the count number -> +1 that represents the ExternalID
	assert.Equal(b.t, fmt.Sprintf("%s->%s", strconv.Itoa(relationID), strconv.Itoa(relationID+1)), relation.ExternalID)
	relationID = relationID + 2
}

// HandleError handles any errors during topology gathering
func (b *TestTopologySubmitter) HandleError(err error) {
	// match the error message
	assert.Equal(b.t, "ErrorTestCollector", err.Error())
}

// TestCollector implements the ClusterTopologyCollector interface.
type TestCollector struct {
	ComponentChan chan<- *topology.Component
	RelationChan  chan<- *topology.Relation
	collectors.ClusterTopologyCollector
}

// NewTestCollector
func NewTestCollector(componentChannel chan<- *topology.Component, relationChannel chan<- *topology.Relation, clusterTopologyCollector collectors.ClusterTopologyCollector) collectors.ClusterTopologyCollector {
	return &TestCollector{
		ComponentChan:            componentChannel,
		RelationChan:             relationChannel,
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the TestCollector
func (*TestCollector) GetName() string {
	return "Test Collector"
}

// Collects and Publishes dummy Components and Relations
func (tc *TestCollector) CollectorFunction() error {
	tc.ComponentChan <- &topology.Component{ExternalID: "1", Type: topology.Type{Name: "component-type"}}
	tc.ComponentChan <- &topology.Component{ExternalID: "2", Type: topology.Type{Name: "component-type"}}
	tc.ComponentChan <- &topology.Component{ExternalID: "3", Type: topology.Type{Name: "component-type"}}
	tc.ComponentChan <- &topology.Component{ExternalID: "4", Type: topology.Type{Name: "component-type"}}

	tc.RelationChan <- &topology.Relation{ExternalID: "1->2"}
	tc.RelationChan <- &topology.Relation{ExternalID: "3->4"}

	return nil
}

// ErrorTestCollector implements the ClusterTopologyCollector interface.
type ErrorTestCollector struct {
	ComponentChan chan<- *topology.Component
	RelationChan  chan<- *topology.Relation
	collectors.ClusterTopologyCollector
}

// NewErrorTestCollector
func NewErrorTestCollector(componentChannel chan<- *topology.Component, relationChannel chan<- *topology.Relation, clusterTopologyCollector collectors.ClusterTopologyCollector) collectors.ClusterTopologyCollector {
	return &ErrorTestCollector{
		ComponentChan:            componentChannel,
		RelationChan:             relationChannel,
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the ErrorTestCollector
func (*ErrorTestCollector) GetName() string {
	return "Error Test Collector"
}

// Returns a error
func (etc *ErrorTestCollector) CollectorFunction() error {
	return errors.New("ErrorTestCollector")
}
