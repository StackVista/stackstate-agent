// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver

package topologycollectors

import (
	"testing"

	"github.com/DataDog/datadog-agent/pkg/topology"
	"k8s.io/apimachinery/pkg/version"

	"github.com/stretchr/testify/assert"
)

func TestCollectorInterface(t *testing.T) {
	t.Run("with sourceProperties enabled and status disabled", func(t *testing.T) {
		testCollectorInterface(t, true, false)
	})
	t.Run("with sourceProperties enabled and status enabled", func(t *testing.T) {
		testCollectorInterface(t, true, true)
	})
	t.Run("with sourceProperties disabled and status disabled", func(t *testing.T) {
		testCollectorInterface(t, false, false)
	})
	t.Run("with sourceProperties disabled and status enabled", func(t *testing.T) {
		testCollectorInterface(t, false, true)
	})

}

func testCollectorInterface(t *testing.T, sourcePropertiesEnabled bool, exposeKubernetesStatusEnabled bool) {
	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)
	instance := topology.Instance{Type: "kubernetes", URL: "Test-Cluster-Name"}
	clusterType := Kubernetes
	clusterTopologyCommon := NewClusterTopologyCommon(instance, clusterType, nil, sourcePropertiesEnabled, componentChannel, relationChannel, &version.Info{Major: "1", Minor: "21"}, exposeKubernetesStatusEnabled)
	testCollector := NewTestCollector(NewClusterTopologyCollector(clusterTopologyCommon))

	actualClusterExternalID := testCollector.buildClusterExternalID()
	assert.Equal(t, "urn:cluster:/kubernetes:Test-Cluster-Name", actualClusterExternalID)

	testNameSpace := "test-namespace"

	podName := "test-pod-name"
	containerName := "test-container-name"
	actualContainerExternalID := testCollector.buildContainerExternalID(testNameSpace, podName, containerName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:pod/test-pod-name:container/test-container-name", actualContainerExternalID)

	daemonSetName := "test-daemonset"
	actualDaemonSetExternalID := testCollector.buildDaemonSetExternalID(testNameSpace, daemonSetName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:daemonset/test-daemonset", actualDaemonSetExternalID)

	deploymentName := "test-deployment"
	actualDeploymentExternalID := testCollector.buildDeploymentExternalID(testNameSpace, deploymentName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:deployment/test-deployment", actualDeploymentExternalID)

	nodeName := "test-node"
	actualNodeExternalID := testCollector.buildNodeExternalID(nodeName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:node/test-node", actualNodeExternalID)

	actualPodExternalID := testCollector.buildPodExternalID(testNameSpace, podName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:pod/test-pod-name", actualPodExternalID)

	replicaSetName := "test-replicaset"
	actualReplicaSetExternalID := testCollector.buildReplicaSetExternalID(testNameSpace, replicaSetName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:replicaset/test-replicaset", actualReplicaSetExternalID)

	serviceName := "test-service"
	actualServiceExternalID := testCollector.buildServiceExternalID(testNameSpace, serviceName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:service/test-service", actualServiceExternalID)

	statefulSetName := "test-statefulset"
	actualStatefulSetExternalID := testCollector.buildStatefulSetExternalID(testNameSpace, statefulSetName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:statefulset/test-statefulset", actualStatefulSetExternalID)

	configMapNamespace := "test-namespace"
	configMapName := "test-configmap"
	actualConfigMapExternalID := testCollector.buildConfigMapExternalID(configMapNamespace, configMapName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:configmap/test-configmap", actualConfigMapExternalID)

	cronJobName := "test-cronjob"
	actualCronJobExternalID := testCollector.buildCronJobExternalID(testNameSpace, cronJobName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:cronjob/test-cronjob", actualCronJobExternalID)

	jobName := "test-job"
	actualJobExternalID := testCollector.buildJobExternalID(testNameSpace, jobName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:job/test-job", actualJobExternalID)

	ingressName := "test-ingress"
	actualIngressExternalID := testCollector.buildIngressExternalID(testNameSpace, ingressName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:ingress/test-ingress", actualIngressExternalID)

	volumeName := "test-volume"
	actualVolumeExternalID := testCollector.buildVolumeExternalID(testNameSpace, volumeName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:test-namespace:volume/test-volume", actualVolumeExternalID)

	persistentVolumeName := "test-persistent-volume"
	actualPersistentVolumeExternalID := testCollector.buildPersistentVolumeExternalID(persistentVolumeName)
	assert.Equal(t, "urn:kubernetes:/Test-Cluster-Name:persistent-volume/test-persistent-volume", actualPersistentVolumeExternalID)

	endpointName := "test-url"
	actualEndpointExternalID := testCollector.buildEndpointExternalID(endpointName)
	assert.Equal(t, "urn:endpoint:/Test-Cluster-Name:test-url", actualEndpointExternalID)

	expectedCollectorName := "Test Collector"
	actualCollectorName := testCollector.GetName()
	assert.Equal(t, expectedCollectorName, actualCollectorName)

	expectedErrorMessage := "CollectorFunction NotImplemented"
	actualResult := testCollector.CollectorFunction()
	if actualResult != nil && actualResult.Error() != expectedErrorMessage {
		t.Errorf("Error actual = %v, and Expected = %v.", actualResult, expectedErrorMessage)
	}

	actualCollectorInstanceURL := testCollector.GetInstance().URL
	assert.Equal(t, instance.URL, actualCollectorInstanceURL)
	actualCollectorInstanceType := testCollector.GetInstance().Type
	assert.Equal(t, instance.Type, actualCollectorInstanceType)
}

// TestCollector implements the ClusterTopologyCollector interface.
type TestCollector struct {
	ClusterTopologyCollector
}

// NewTestCollector
func NewTestCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &TestCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the TestCollector
func (*TestCollector) GetName() string {
	return "Test Collector"
}
