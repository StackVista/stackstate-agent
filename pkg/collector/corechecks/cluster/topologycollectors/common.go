// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterTopologyCommon should be mixed in this interface for basic functionality on any real collector
type ClusterTopologyCommon interface {
	GetAPIClient() apiserver.APICollectorClient
	GetInstance() topology.Instance
	GetName() string
	CreateRelation(sourceExternalID, targetExternalID, typeName string) *topology.Relation
	CreateRelationData(sourceExternalID, targetExternalID, typeName string, data map[string]interface{}) *topology.Relation
	initTags(meta metav1.ObjectMeta) map[string]string
	buildClusterExternalID() string
	buildConfigMapExternalID(namespace, configMapName string) string
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
	buildPersistentVolumeExternalID(persistentVolumeID string) string
	buildEndpointExternalID(endpointID string) string
}

type clusterTopologyCommon struct {
	Instance           topology.Instance
	APICollectorClient apiserver.APICollectorClient
}

// NewClusterTopologyCommon creates a clusterTopologyCommon
func NewClusterTopologyCommon(instance topology.Instance, ac apiserver.APICollectorClient) ClusterTopologyCommon {
	return &clusterTopologyCommon{Instance: instance, APICollectorClient: ac}
}

// GetName
func (c *clusterTopologyCommon) GetName() string {
	return "Unknown Collector"
}

// GetInstance
func (c *clusterTopologyCommon) GetInstance() topology.Instance {
	return c.Instance
}

// GetAPIClient
func (c *clusterTopologyCommon) GetAPIClient() apiserver.APICollectorClient {
	return c.APICollectorClient
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

// buildClusterExternalID
func (c *clusterTopologyCommon) buildClusterExternalID() string {
	return fmt.Sprintf("urn:cluster:/%s:%s", c.Instance.Type, c.Instance.URL)
}

// buildNodeExternalID
// nodeName
func (c *clusterTopologyCommon) buildNodeExternalID(nodeName string) string {
	return fmt.Sprintf("urn:/%s:%s:node:%s", c.Instance.Type, c.Instance.URL, nodeName)
}

// buildPodExternalID
// podName
func (c *clusterTopologyCommon) buildPodExternalID(namespace, podName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:pod:%s", c.Instance.Type, c.Instance.URL, namespace, podName)
}

// buildContainerExternalID
// podName, containerName
func (c *clusterTopologyCommon) buildContainerExternalID(namespace, podName, containerName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:pod:%s:container:%s", c.Instance.Type, c.Instance.URL, namespace, podName, containerName)
}

// buildServiceExternalID
// serviceID
func (c *clusterTopologyCommon) buildServiceExternalID(namespace, serviceName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:service:%s", c.Instance.Type, c.Instance.URL, namespace, serviceName)
}

// buildDaemonSetExternalID
// daemonSetName
func (c *clusterTopologyCommon) buildDaemonSetExternalID(namespace, daemonSetName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:daemonset:%s", c.Instance.Type, c.Instance.URL, namespace, daemonSetName)
}

// buildDeploymentExternalID
// deploymentName
func (c *clusterTopologyCommon) buildDeploymentExternalID(namespace, deploymentName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:deployment:%s", c.Instance.Type, c.Instance.URL, namespace, deploymentName)
}

// buildReplicaSetExternalID
// replicaSetID
func (c *clusterTopologyCommon) buildReplicaSetExternalID(namespace, replicaSetName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:replicaset:%s", c.Instance.Type, c.Instance.URL, namespace, replicaSetName)
}

// buildStatefulSetExternalID
// statefulSetID
func (c *clusterTopologyCommon) buildStatefulSetExternalID(namespace, statefulSetName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:statefulset:%s", c.Instance.Type, c.Instance.URL, namespace, statefulSetName)
}

// buildConfigMapExternalID
// namespace
// configMapID
func (c *clusterTopologyCommon) buildConfigMapExternalID(namespace, configMapName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:configmap:%s", c.Instance.Type, c.Instance.URL, namespace, configMapName)
}

// buildCronJobExternalID
// cronJobID
func (c *clusterTopologyCommon) buildCronJobExternalID(namespace, cronJobName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:cronjob:%s", c.Instance.Type, c.Instance.URL, namespace, cronJobName)
}

// buildJobExternalID
// jobID
func (c *clusterTopologyCommon) buildJobExternalID(namespace, jobName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:job:%s", c.Instance.Type, c.Instance.URL, namespace, jobName)
}

// buildIngressExternalID
// ingressID
func (c *clusterTopologyCommon) buildIngressExternalID(namespace, ingressName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:ingress:%s", c.Instance.Type, c.Instance.URL, namespace, ingressName)
}

// buildVolumeExternalID
// volumeID
func (c *clusterTopologyCommon) buildVolumeExternalID(namespace, volumeName string) string {
	return fmt.Sprintf("urn:/%s:%s:namespace:%s:volume:%s", c.Instance.Type, c.Instance.URL, namespace, volumeName)
}

// buildPersistentVolumeExternalID
// persistentVolumeID
func (c *clusterTopologyCommon) buildPersistentVolumeExternalID(persistentVolumeID string) string {
	return fmt.Sprintf("urn:/%s:%s:persistent-volume:%s", c.Instance.Type, c.Instance.URL, persistentVolumeID)
}

// buildEndpointExternalID
// endpointID
func (c *clusterTopologyCommon) buildEndpointExternalID(endpointID string) string {
	return fmt.Sprintf("urn:endpoint:/%s:%s", c.Instance.URL, endpointID)
}

func (c *clusterTopologyCommon) initTags(meta metav1.ObjectMeta) map[string]string {
	tags := make(map[string]string, 0)
	if meta.Labels != nil {
		tags = meta.Labels
	}

	// set the cluster name and the namespace
	tags["cluster-name"] = c.Instance.URL
	if meta.Namespace != "" {
		tags["namespace"] = meta.Namespace
	}

	return tags
}
