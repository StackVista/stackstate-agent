//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

func TestRelationCorrelator(t *testing.T) {
	clusterName := "test-cluster-name"
	namespace := "default"
	pod1Name := "pod-1"
	pod2Name := "pod-2"
	configMap1Name := "config-map-1"
	configMap2Name := "config-map-2"
	secret1Name := "secret-1"
	secret2Name := "secret-2"
	someTimestamp := metav1.NewTime(time.Now())

	pod1 := podWithConfigMapEnv(namespace, pod1Name, configMap1Name, configMap2Name, someTimestamp)
	pod2 := podWithSecretEnv(namespace, pod2Name, secret1Name, secret2Name, someTimestamp)
	configMap1 := configMap(namespace, configMap1Name, someTimestamp)
	configMap2 := configMap(namespace, configMap2Name, someTimestamp)
	secret1 := secret(namespace, secret1Name, someTimestamp)
	secret2 := secret(namespace, secret2Name, someTimestamp)

	components, relations := executeRelationCorrelation(t,
		[]coreV1.Pod{pod1, pod2},
		[]coreV1.ConfigMap{configMap1, configMap2},
		[]coreV1.Secret{secret1, secret2})

	expectedPod1Id := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod1Name)
	expectedPod2Id := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod2Name)
	expectedConfigMap1Id := fmt.Sprintf("urn:kubernetes:/%s:%s:configmap/%s", clusterName, namespace, configMap1Name)
	expectedConfigMap2Id := fmt.Sprintf("urn:kubernetes:/%s:%s:configmap/%s", clusterName, namespace, configMap2Name)
	expectedSecret1Id := fmt.Sprintf("urn:kubernetes:/%s:%s:secret/%s", clusterName, namespace, secret1Name)
	expectedSecret2Id := fmt.Sprintf("urn:kubernetes:/%s:%s:secret/%s", clusterName, namespace, secret2Name)

	expectedComponents := []*topology.Component{
		podComponent(namespace, pod1Name, someTimestamp),
		podComponent(namespace, pod2Name, someTimestamp),
		configMapComponent(namespace, configMap1Name, someTimestamp),
		configMapComponent(namespace, configMap2Name, someTimestamp),
		secretComponent(namespace, secret1Name, someTimestamp),
		secretComponent(namespace, secret2Name, someTimestamp),
	}
	expectedRelations := []*topology.Relation{
		simpleRelation(expectedPod1Id, expectedConfigMap2Id, "uses"),
		simpleRelation(expectedPod1Id, expectedConfigMap1Id, "uses_value"),
		simpleRelation(expectedPod2Id, expectedSecret2Id, "uses"),
		simpleRelation(expectedPod2Id, expectedSecret1Id, "uses_value"),
		// it should not create relations:
		// - pod1 -> non-existing-configMap
		// - pod2 -> non-existing-secret
	}

	assert.EqualValues(t, expectedComponents, components)
	assert.EqualValues(t, expectedRelations, relations)
	for _, expected := range expectedRelations {
		found := false
		for _, actual := range relations {
			if expected.ExternalID == actual.ExternalID {
				assert.EqualValues(t, expected, actual)
				found = true
				break
			}
		}
		assert.True(t, found, "Could not find relation %s", expected.ExternalID)
	}
	return
}

func podWithConfigMapEnv(namespace string, name string, configMapName string, configMapEnvSourceName string, timestamp metav1.Time) coreV1.Pod {
	trueV := true
	return coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		Spec: coreV1.PodSpec{
			HostNetwork: true,
			Containers: []coreV1.Container{
				{
					Name:  "container-1",
					Image: "docker/image/repo/container:latest",
					Env: []coreV1.EnvVar{
						{
							Name: "env-var",
							ValueFrom: &coreV1.EnvVarSource{
								ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
									LocalObjectReference: coreV1.LocalObjectReference{Name: configMapName},
								},
							},
						},
						{
							Name: "env-var2",
							ValueFrom: &coreV1.EnvVarSource{
								ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
									LocalObjectReference: coreV1.LocalObjectReference{Name: "non-existing-configMap"},
									Optional:             &trueV,
								},
							},
						},
					},
					EnvFrom: []coreV1.EnvFromSource{
						{
							ConfigMapRef: &coreV1.ConfigMapEnvSource{
								LocalObjectReference: coreV1.LocalObjectReference{Name: configMapEnvSourceName},
							},
						},
					},
				},
			},
		},
	}
}

func podWithSecretEnv(namespace string, name string, secretEnvName string, secretEnvSourceName string, timestamp metav1.Time) coreV1.Pod {
	trueV := true
	return coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		Spec: coreV1.PodSpec{
			HostNetwork: true,
			Containers: []coreV1.Container{
				{
					Name:  "container-1",
					Image: "docker/image/repo/container:latest",
					Env: []coreV1.EnvVar{
						{
							Name: "env-var",
							ValueFrom: &coreV1.EnvVarSource{
								SecretKeyRef: &coreV1.SecretKeySelector{
									LocalObjectReference: coreV1.LocalObjectReference{Name: secretEnvName},
								},
							},
						},
					},
					EnvFrom: []coreV1.EnvFromSource{
						{
							SecretRef: &coreV1.SecretEnvSource{
								LocalObjectReference: coreV1.LocalObjectReference{Name: secretEnvSourceName},
							},
						},
						{
							SecretRef: &coreV1.SecretEnvSource{
								LocalObjectReference: coreV1.LocalObjectReference{Name: "non-existing-secret"},
								Optional:             &trueV,
							},
						},
					},
				},
			},
		},
	}
}

func configMap(namespace string, name string, timestamp metav1.Time) coreV1.ConfigMap {
	return coreV1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: timestamp,
			Namespace:         namespace,
			UID:               types.UID(name),
			GenerateName:      "",
			ResourceVersion:   "123",
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Manager:    "ignored",
					Operation:  "Updated",
					APIVersion: "whatever",
					Time:       &timestamp,
					FieldsType: "whatever",
				},
			},
		},
	}
}

func configMapComponent(namespace string, name string, timestamp metav1.Time) *topology.Component {
	return &topology.Component{
		ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:configmap/%s", namespace, name),
		Type: topology.Type{
			Name: "configmap",
		},
		Data: topology.Data{
			"name": name,
			"tags": map[string]string{
				"cluster-name": "test-cluster-name",
				"namespace":    namespace,
			},
			"identifiers": []string{
				fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:configmap/%s", namespace, name),
			},
			"creationTimestamp": timestamp,
			"uid":               types.UID(name),
		},
	}
}

func secret(namespace string, name string, timestamp metav1.Time) coreV1.Secret {
	return coreV1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind: "",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: timestamp,
			Namespace:         namespace,
			UID:               types.UID(name),
			GenerateName:      "",
			ResourceVersion:   "123",
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Manager:    "ignored",
					Operation:  "Updated",
					APIVersion: "whatever",
					Time:       &timestamp,
					FieldsType: "whatever",
				},
			},
		},
	}
}

func secretComponent(namespace string, name string, timestamp metav1.Time) *topology.Component {
	return &topology.Component{
		ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:secret/%s", namespace, name),
		Type: topology.Type{
			Name: "secret",
		},
		Data: topology.Data{
			"name": name,
			"tags": map[string]string{
				"cluster-name": "test-cluster-name",
				"namespace":    namespace,
			},
			"creationTimestamp": timestamp,
			"data":              "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			"identifiers": []string{
				fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:secret/%s", namespace, name),
			},
			"uid": types.UID(name),
		},
	}
}

func executeRelationCorrelation(
	t *testing.T,
	pods []coreV1.Pod,
	configMaps []coreV1.ConfigMap,
	secrets []coreV1.Secret,
) ([]*topology.Component, []*topology.Relation) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	clusterAPIClient := MockRelationCorrelatorAPIClient{
		pods: pods, configMaps: configMaps, secrets: secrets,
	}

	podCorrChannel := make(chan *PodEndpointCorrelation)
	containerCorrChannel := make(chan *ContainerCorrelation)
	volumeCorrChannel := make(chan *VolumeCorrelation)
	collectorsDoneChan := make(chan bool)
	correlatorsDoneChannel := make(chan bool)

	commonClusterCollector := NewTestCommonClusterCollector(clusterAPIClient, componentChannel, relationChannel, false)
	podCollector := NewPodCollector(
		containerCorrChannel, volumeCorrChannel,
		podCorrChannel,
		commonClusterCollector,
	)
	configMapCollector := NewConfigMapCollector(commonClusterCollector, TestMaxDataSize)
	secretCollector := NewSecretCollector(commonClusterCollector)

	relationCorrelator := NewRelationCorrelator(relationChannel,
		collectorsDoneChan, NewClusterTopologyCorrelator(commonClusterCollector))

	collectorsFinished := false

	go func() {
		var err error
		err = podCollector.CollectorFunction()
		assert.NoError(t, err)
		err = configMapCollector.CollectorFunction()
		assert.NoError(t, err)
		err = secretCollector.CollectorFunction()
		assert.NoError(t, err)

		collectorsFinished = true
		collectorsDoneChan <- true
	}()

	go func() {
		var err error
		err = relationCorrelator.CorrelateFunction()
		assert.NoError(t, err)
		correlatorsDoneChannel <- true
	}()

	components := make([]*topology.Component, 0)
	relations := make([]*topology.Relation, 0)

L:
	for {
		select {
		case c := <-componentChannel:
			components = append(components, c)
		case r := <-relationChannel:
			relations = append(relations, r)
		case <-correlatorsDoneChannel:
			if collectorsFinished {
				break L
			}
		}
	}

	return components, relations
}

type MockRelationCorrelatorAPIClient struct {
	pods       []coreV1.Pod
	configMaps []coreV1.ConfigMap
	secrets    []coreV1.Secret
	apiserver.APICollectorClient
}

func (m MockRelationCorrelatorAPIClient) GetPods() ([]coreV1.Pod, error) {
	return m.pods, nil
}

func (m MockRelationCorrelatorAPIClient) GetConfigMaps() ([]coreV1.ConfigMap, error) {
	return m.configMaps, nil
}

func (m MockRelationCorrelatorAPIClient) GetSecrets() ([]coreV1.Secret, error) {
	return m.secrets, nil
}
