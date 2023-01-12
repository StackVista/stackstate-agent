//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"k8s.io/apimachinery/pkg/version"

	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/cluster/urn"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterTopologyCommon should be mixed in this interface for basic functionality on any real collector
type ClusterTopologyCommon interface {
	GetAPIClient() apiserver.APICollectorClient
	GetInstance() topology.Instance
	GetName() string
	GetURNBuilder() urn.Builder
	CreateRelation(sourceExternalID, targetExternalID, typeName string) *topology.Relation
	CreateRelationData(sourceExternalID, targetExternalID, typeName string, data map[string]interface{}) *topology.Relation
	IsSourcePropertiesFeatureEnabled() bool
	IsExposeKubernetesStatusEnabled() bool
	initTags(meta metav1.ObjectMeta) map[string]string
	buildClusterExternalID() string
	buildConfigMapExternalID(namespace, configMapName string) string
	buildSecretExternalID(namespace, secretName string) string
	buildNamespaceExternalID(namespaceName string) string
	buildContainerExternalID(namespace, podName, containerName string) string
	buildDaemonSetExternalID(namespace, daemonSetName string) string
	buildDeploymentExternalID(namespace, deploymentName string) string
	buildNodeExternalID(nodeName string) string
	buildPodExternalID(namespace, podName string) string
	buildReplicaSetExternalID(namespace, replicaSetName string) string
	buildServiceExternalID(namespace, serviceName string) string
	buildStatefulSetExternalID(namespace, statefulSetName string) string
	buildCronJobExternalID(namespace, cronJobName string) string
	buildJobExternalID(namespace, jobName string) string
	buildIngressExternalID(namespace, ingressName string) string
	buildVolumeExternalID(namespace, volumeName string) string
	buildPersistentVolumeExternalID(persistentVolumeName string) string
	buildPersistentVolumeClaimExternalID(persistentVolumeName string) string
	buildEndpointExternalID(endpointID string) string
	maximumMinorVersion(version int) bool
	minimumMinorVersion(version int) bool
	SubmitComponent(component *topology.Component)
	SubmitRelation(relation *topology.Relation)
	SetUseRelationCache(value bool)
	CorrelateRelations()
}

type clusterTopologyCommon struct {
	Instance                      topology.Instance
	APICollectorClient            apiserver.APICollectorClient
	urn                           urn.Builder
	sourcePropertiesEnabled       bool
	componentChan                 chan<- *topology.Component
	componentIDCache              sync.Map
	relationChan                  chan<- *topology.Relation
	possibleRelations             []*topology.Relation
	k8sVersion                    *version.Info
	useRelationCache              bool
	relationCacheWG               sync.WaitGroup
	exposeKubernetesStatusEnabled bool
}

// NewClusterTopologyCommon creates a clusterTopologyCommon
func NewClusterTopologyCommon(
	instance topology.Instance,
	ac apiserver.APICollectorClient,
	spEnabled bool,
	componentChan chan<- *topology.Component,
	relationChan chan<- *topology.Relation,
	k8sVersion *version.Info,
	kubernetesStatusEnabled bool,
) ClusterTopologyCommon {
	return &clusterTopologyCommon{
		Instance:                      instance,
		APICollectorClient:            ac,
		urn:                           urn.NewURNBuilder(urn.ClusterTypeFromString(instance.Type), instance.URL),
		sourcePropertiesEnabled:       spEnabled,
		componentChan:                 componentChan,
		componentIDCache:              sync.Map{},
		relationChan:                  relationChan,
		k8sVersion:                    k8sVersion,
		useRelationCache:              true,
		relationCacheWG:               sync.WaitGroup{},
		exposeKubernetesStatusEnabled: kubernetesStatusEnabled,
	}
}

// SubmitComponent sends a component to the Component channel and adds its External ID to the cache if the relation cache is being used
func (c *clusterTopologyCommon) SubmitComponent(component *topology.Component) {
	c.componentChan <- component
	if c.useRelationCache {
		c.componentIDCache.Store(component.ExternalID, true)
	}
}

// SubmitRelation sends a relation to the Relation channel or adds it to the possibleRelations cache if it's being used
func (c *clusterTopologyCommon) SubmitRelation(relation *topology.Relation) {
	if c.useRelationCache {
		_, sourceExists := c.componentIDCache.Load(relation.SourceID)
		_, targetExists := c.componentIDCache.Load(relation.TargetID)
		if sourceExists && targetExists {
			c.relationChan <- relation
		} else {
			c.relationCacheWG.Add(1)
			c.possibleRelations = append(c.possibleRelations, relation)
			c.relationCacheWG.Done()
		}
	} else {
		c.relationChan <- relation
	}
}

func (c *clusterTopologyCommon) CorrelateRelations() {
	c.relationCacheWG.Add(1)
	for _, relation := range c.possibleRelations {
		_, sourceExists := c.componentIDCache.Load(relation.SourceID)
		_, targetExists := c.componentIDCache.Load(relation.TargetID)
		if sourceExists && targetExists {
			c.relationChan <- relation
		} else {
			if !sourceExists {
				log.Debugf("Ignoring relation '%s' because source does not exist", relation.ExternalID)
			} else {
				log.Debugf("Ignoring relation '%s' because target does not exist", relation.ExternalID)
			}
		}
	}
	c.relationCacheWG.Done()
}

// SetUseRelationCache sets if the relation cache should be used or not
func (c *clusterTopologyCommon) SetUseRelationCache(value bool) {
	c.useRelationCache = value
}

// GetName returns the collector name
func (*clusterTopologyCommon) GetName() string {
	return "Unknown Collector"
}

// GetInstance returns the topology.Instance
func (c *clusterTopologyCommon) GetInstance() topology.Instance {
	return c.Instance
}

// GetAPIClient returns the Kubernetes API client
func (c *clusterTopologyCommon) GetAPIClient() apiserver.APICollectorClient {
	return c.APICollectorClient
}

// GetURNBuilder returns the URN builder
func (c *clusterTopologyCommon) GetURNBuilder() urn.Builder {
	return c.urn
}

// CreateRelationData creates a StackState relation called typeName for the given sourceExternalID and targetExternalID
func (c *clusterTopologyCommon) CreateRelationData(sourceExternalID, targetExternalID, typeName string, data map[string]interface{}) *topology.Relation {
	var _data map[string]interface{}

	if data != nil {
		_data = data
	} else {
		_data = map[string]interface{}{}
	}

	return &topology.Relation{
		ExternalID: fmt.Sprintf("%s->%s", sourceExternalID, targetExternalID),
		SourceID:   sourceExternalID,
		TargetID:   targetExternalID,
		Type:       topology.Type{Name: typeName},
		Data:       _data,
	}
}

// CreateRelation creates a StackState relation called typeName for the given sourceExternalID and targetExternalID
func (c *clusterTopologyCommon) CreateRelation(sourceExternalID, targetExternalID, typeName string) *topology.Relation {
	return &topology.Relation{
		ExternalID: fmt.Sprintf("%s->%s", sourceExternalID, targetExternalID),
		SourceID:   sourceExternalID,
		TargetID:   targetExternalID,
		Type:       topology.Type{Name: typeName},
		Data:       map[string]interface{}{},
	}
}

// IsSourcePropertiesFeatureEnabled return value of Source Properties feature flag
func (c *clusterTopologyCommon) IsSourcePropertiesFeatureEnabled() bool {
	return c.sourcePropertiesEnabled
}

// IsExposeKubernetesStatusEnabled return value of Expose Kubernetes Status feature flag
func (c *clusterTopologyCommon) IsExposeKubernetesStatusEnabled() bool {
	return c.exposeKubernetesStatusEnabled
}

// buildClusterExternalID
func (c *clusterTopologyCommon) buildClusterExternalID() string {
	return c.urn.BuildClusterExternalID()
}

// buildNodeExternalID creates the urn external identifier for a cluster node
func (c *clusterTopologyCommon) buildNodeExternalID(nodeName string) string {
	return c.urn.BuildNodeExternalID(nodeName)
}

// buildPodExternalID creates the urn external identifier for a cluster pod
func (c *clusterTopologyCommon) buildPodExternalID(namespace, podName string) string {
	return c.urn.BuildPodExternalID(namespace, podName)
}

// buildContainerExternalID creates the urn external identifier for a pod's container
func (c *clusterTopologyCommon) buildContainerExternalID(namespace, podName, containerName string) string {
	return c.urn.BuildContainerExternalID(namespace, podName, containerName)
}

// buildServiceExternalID creates the urn external identifier for a cluster service
func (c *clusterTopologyCommon) buildServiceExternalID(namespace, serviceName string) string {
	return c.urn.BuildServiceExternalID(namespace, serviceName)
}

// buildDaemonSetExternalID creates the urn external identifier for a cluster daemon set
func (c *clusterTopologyCommon) buildDaemonSetExternalID(namespace, daemonSetName string) string {
	return c.urn.BuildDaemonSetExternalID(namespace, daemonSetName)
}

// buildDeploymentExternalID creates the urn external identifier for a cluster deployment
func (c *clusterTopologyCommon) buildDeploymentExternalID(namespace, deploymentName string) string {
	return c.urn.BuildDeploymentExternalID(namespace, deploymentName)
}

// buildReplicaSetExternalID creates the urn external identifier for a cluster replica set
func (c *clusterTopologyCommon) buildReplicaSetExternalID(namespace, replicaSetName string) string {
	return c.urn.BuildReplicaSetExternalID(namespace, replicaSetName)
}

// buildStatefulSetExternalID creates the urn external identifier for a cluster stateful set
func (c *clusterTopologyCommon) buildStatefulSetExternalID(namespace, statefulSetName string) string {
	return c.urn.BuildStatefulSetExternalID(namespace, statefulSetName)
}

// buildConfigMapExternalID creates the urn external identifier for a cluster config map
func (c *clusterTopologyCommon) buildConfigMapExternalID(namespace, configMapName string) string {
	return c.urn.BuildConfigMapExternalID(namespace, configMapName)
}

// buildSecretExternalID creates the urn external identifier for a cluster secret
func (c *clusterTopologyCommon) buildSecretExternalID(namespace, secretName string) string {
	return c.urn.BuildSecretExternalID(namespace, secretName)
}

// buildNamespaceExternalID creates the urn external identifier for a cluster namespace
func (c *clusterTopologyCommon) buildNamespaceExternalID(namespaceName string) string {
	return c.urn.BuildNamespaceExternalID(namespaceName)
}

// buildCronJobExternalID creates the urn external identifier for a cluster cron job
func (c *clusterTopologyCommon) buildCronJobExternalID(namespace, cronJobName string) string {
	return c.urn.BuildCronJobExternalID(namespace, cronJobName)
}

// buildJobExternalID creates the urn external identifier for a cluster job
func (c *clusterTopologyCommon) buildJobExternalID(namespace, jobName string) string {
	return c.urn.BuildJobExternalID(namespace, jobName)
}

// buildIngressExternalID creates the urn external identifier for a cluster ingress
func (c *clusterTopologyCommon) buildIngressExternalID(namespace, ingressName string) string {
	return c.urn.BuildIngressExternalID(namespace, ingressName)
}

// buildVolumeExternalID creates the urn external identifier for a cluster volume
func (c *clusterTopologyCommon) buildVolumeExternalID(namespace, volumeName string) string {
	return c.urn.BuildVolumeExternalID(namespace, volumeName)
}

// buildPersistentVolumeExternalID creates the urn external identifier for a cluster persistent volume
func (c *clusterTopologyCommon) buildPersistentVolumeExternalID(persistentVolumeName string) string {
	return c.urn.BuildPersistentVolumeExternalID(persistentVolumeName)
}

// buildPersistentVolumeClaimExternalID creates the urn external identifier for a cluster persistent volume
func (c *clusterTopologyCommon) buildPersistentVolumeClaimExternalID(persistentVolumeClaimName string) string {
	return c.urn.BuildPersistentVolumeClaimExternalID(persistentVolumeClaimName)
}

// buildEndpointExternalID
// endpointID
func (c *clusterTopologyCommon) buildEndpointExternalID(endpointID string) string {
	return c.urn.BuildEndpointExternalID(endpointID)
}

func (c *clusterTopologyCommon) initTags(meta metav1.ObjectMeta) map[string]string {
	tags := make(map[string]string, 0)
	if meta.Labels != nil {
		for k, v := range meta.Labels {
			tags[k] = v
		}
	}

	// set the cluster name and the namespace
	tags["cluster-name"] = c.Instance.URL
	if meta.Namespace != "" {
		tags["namespace"] = meta.Namespace
	}

	return tags
}

var protoJSONMarshaler = jsonpb.Marshaler{
	EnumsAsInts:  false,
	EmitDefaults: false,
}

func marshallK8sObjectToData(msg proto.Message) (map[string]interface{}, error) {
	var buf bytes.Buffer
	if err := protoJSONMarshaler.Marshal(&buf, msg); err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func removeRedundantFields(result map[string]interface{}, keepStatus bool) {
	if !keepStatus {
		visitNestedMap(result, "status", true, func(status map[string]interface{}) {
			for k := range status {
				if !(k == "phase" || k == "nodeInfo" || k == "daemonEndpoints" || k == "message") {
					delete(status, k)
				}
			}
		})
	}
	visitNestedMap(result, "metadata", false, func(metadata map[string]interface{}) {
		// managedFields contains information about who is able to modify certain parts of an object
		// this information is irrelevant to runtime, hence is being dropped here to have smaller status
		// https://kubernetes.io/docs/reference/using-api/server-side-apply/#field-management
		delete(metadata, "managedFields")
		delete(metadata, "resourceVersion")
		visitNestedMap(metadata, "annotations", true, func(annotations map[string]interface{}) {
			delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
		})
	})
}

func removeManagedFields(result map[string]interface{}) {
	visitNestedMap(result, "metadata", false, func(metadata map[string]interface{}) {
		// managedFields contains information about who is able to modify certain parts of an object
		// this information is irrelevant to runtime, hence is being dropped here to have smaller status
		// https://kubernetes.io/docs/reference/using-api/server-side-apply/#field-management
		delete(metadata, "managedFields")
	})
}

func visitNestedMap(parentMap map[string]interface{}, key string, removeEmpty bool, callback func(map[string]interface{})) {
	if nested, ok := parentMap[key]; ok {
		switch nestedMap := nested.(type) {
		case map[string]interface{}:
			callback(nestedMap)
			if removeEmpty && len(nestedMap) == 0 {
				delete(parentMap, key)
			}
		default:
		}
	}
}

type MarshalableKubernetesObject interface {
	metav1.Object
	proto.Message
}

func makeSourceProperties(object MarshalableKubernetesObject) map[string]interface{} {
	sourceProperties, err := marshallK8sObjectToData(object)
	if err != nil {
		_ = log.Warnf("Can't serialize sourceProperties for %s: %v", object.GetSelfLink(), err)
		sourceProperties = map[string]interface{}{
			"serialization_error": fmt.Sprintf("error occurred during serialization of this object: %v", err),
		}
	}
	removeRedundantFields(sourceProperties, false)
	return sourceProperties
}

func makeSourcePropertiesKS(object MarshalableKubernetesObject) map[string]interface{} {
	sourceProperties, err := marshallK8sObjectToData(object)
	if err != nil {
		_ = log.Warnf("Can't serialize sourceProperties for %s: %v", object.GetSelfLink(), err)
		sourceProperties = map[string]interface{}{
			"serialization_error": fmt.Sprintf("error occurred during serialization of this object: %v", err),
		}
	}
	removeRedundantFields(sourceProperties, true)
	return sourceProperties
}

func makeSourcePropertiesFullDetails(object MarshalableKubernetesObject) map[string]interface{} {
	sourceProperties, err := marshallK8sObjectToData(object)
	if err != nil {
		_ = log.Warnf("Can't serialize sourceProperties for %s: %v", object.GetSelfLink(), err)
		sourceProperties = map[string]interface{}{
			"serialization_error": fmt.Sprintf("error occurred during serialization of this object: %v", err),
		}
	}
	removeManagedFields(sourceProperties)
	return sourceProperties
}

func (c *clusterTopologyCommon) minimumMinorVersion(minimumVersion int) bool {
	return c.checkVersion(func(version int) bool {
		return version >= minimumVersion
	})
}

func (c *clusterTopologyCommon) maximumMinorVersion(maximumVersion int) bool {
	return c.checkVersion(func(version int) bool {
		return version <= maximumVersion
	})
}

func (c *clusterTopologyCommon) checkVersion(compare func(version int) bool) bool {
	if c.k8sVersion != nil && c.k8sVersion.Major != "" {
		if c.k8sVersion.Major == "1" {
			minor, err := strconv.Atoi(c.k8sVersion.Minor[:2])
			if err != nil {
				log.Warnf("cannot parse server minor version %q: %w", c.k8sVersion.Minor[:2], err)
				return true
			}
			return compare(minor)
		}
		log.Warnf("Kubernetes versions check failed (Major version is not '1')")
		return false
	}
	log.Warnf("Kubernetes version is undefined")
	return true
}
