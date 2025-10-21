//go:build kubeapiserver

package kubeapi

import (
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"gopkg.in/yaml.v2"
)

const (
	DefaultConfigMapDataSizeLimit = 100 * 1024
)

// TopologyConfig is the config of the API server.
type TopologyConfig struct {
	ClusterName             string          `yaml:"cluster_name"`
	CollectTopology         bool            `yaml:"collect_topology"`
	CollectTimeout          int             `yaml:"collect_timeout"`
	SourcePropertiesEnabled bool            `yaml:"source_properties_enabled"`
	ConfigMapMaxDataSize    int             `yaml:"configmap_max_datasize"`
	CSIPVMapperEnabled      bool            `yaml:"csi_pv_mapper_enabled"`
	Resources               ResourcesConfig `yaml:"resources"`
	CheckID                 checkid.ID
	Instance                topology.Instance
}

type ResourcesConfig struct {
	Persistentvolumes      bool `yaml:"persistentvolumes"`
	Persistentvolumeclaims bool `yaml:"persistentvolumeclaims"`
	Endpoints              bool `yaml:"endpoints"`
	Namespaces             bool `yaml:"namespaces"`
	ConfigMaps             bool `yaml:"configmaps"`
	Daemonsets             bool `yaml:"daemonsets"`
	Deployments            bool `yaml:"deployments"`
	Replicasets            bool `yaml:"replicasets"`
	Statefulsets           bool `yaml:"statefulsets"`
	Ingresses              bool `yaml:"ingresses"`
	Jobs                   bool `yaml:"jobs"`
	CronJobs               bool `yaml:"cronjobs"`
	Secrets                bool `yaml:"secrets"`
}

var defaultResourcesConfig = ResourcesConfig{
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
}

func (c *TopologyConfig) parse(data []byte) error {
	// default values
	c.Resources = defaultResourcesConfig
	c.ClusterName = config.Datadog.GetString("cluster_name")
	c.CollectTopology = config.Datadog.GetBool("collect_kubernetes_topology")
	c.CollectTimeout = config.Datadog.GetInt("collect_kubernetes_timeout")
	c.SourcePropertiesEnabled = config.Datadog.GetBool("kubernetes_source_properties_enabled")
	c.ConfigMapMaxDataSize = config.Datadog.GetInt("configmap_max_datasize")
	c.CSIPVMapperEnabled = config.Datadog.GetBool("kubernetes_csi_pv_mapper_enabled")
	if c.ConfigMapMaxDataSize == 0 {
		c.ConfigMapMaxDataSize = DefaultConfigMapDataSizeLimit
	}

	return yaml.Unmarshal(data, c)
}

// TopologySubmitter provides functionality to submit topology data
type TopologySubmitter interface {
	SubmitStartSnapshot()
	SubmitStopSnapshot()
	SubmitComplete()
	SubmitComponent(component *topology.Component)
	SubmitRelation(relation *topology.Relation)
	HandleError(err error)

	GetErrors() []error
}

// NewBatchTopologySubmitter creates a new instance of BatchTopologySubmitter
func NewBatchTopologySubmitter(api handler.CheckAPI, instance topology.Instance) TopologySubmitter {
	return &BatchTopologySubmitter{
		checkApi: api,
		Instance: instance,
		Errors:   make([]error, 0, 0),
	}
}

// BatchTopologySubmitter provides functionality to submit topology data with the Batcher.
type BatchTopologySubmitter struct {
	checkApi handler.CheckAPI
	Instance topology.Instance
	Errors   []error
}

// SubmitStartSnapshot submits the start for this Check ID and instance
func (b *BatchTopologySubmitter) SubmitStartSnapshot() {
	b.checkApi.SubmitStartSnapshot(b.Instance)
}

// SubmitStopSnapshot submits the stop for this Check ID and instance
func (b *BatchTopologySubmitter) SubmitStopSnapshot() {
	b.checkApi.SubmitStopSnapshot(b.Instance)
}

// SubmitComplete submits the completion for this Check ID
func (b *BatchTopologySubmitter) SubmitComplete() {
	b.checkApi.SubmitComplete()
}

// SubmitComponent takes a component and submits it with the Batcher
func (b *BatchTopologySubmitter) SubmitComponent(component *topology.Component) {
	log.Debugf("Publishing StackState %s component for %s: %v", component.Type.Name, component.ExternalID, component.JSONString())
	b.checkApi.SubmitComponent(b.Instance, *component)
}

// SubmitRelation takes a relation and submits it with the Batcher
func (b *BatchTopologySubmitter) SubmitRelation(relation *topology.Relation) {
	log.Debugf("Publishing StackState %s relation %s->%s", relation.Type.Name, relation.SourceID, relation.TargetID)
	b.checkApi.SubmitRelation(b.Instance, *relation)
}

// HandleError handles any errors during topology gathering
func (b *BatchTopologySubmitter) HandleError(err error) {
	_ = log.Errorf("Error occurred in during topology collection: %s", err.Error())
	b.Errors = append(b.Errors, err)
}

// GetErrors produces the errors reported so far
func (b *BatchTopologySubmitter) GetErrors() []error {
	errs := b.Errors
	b.Errors = make([]error, 0, 0)
	return errs
}
