//go:build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestRelationCorrelation(t *testing.T) {
	testClusterName := "test-cluster-name"
	clusterName := "test-cluster-name"
	namespace := "default"
	pod1Name := "pod-1"
	pod2Name := "pod-2"
	configMap1Name := "config-map-1"
	configMap2Name := "config-map-2"
	secret1Name := "secret-1"
	secret2Name := "secret-2"
	node1Name := "node-1"
	instanceID := "instance-id"
	node1Provider := "aws://eu-west-1/" + instanceID
	someTimestamp := metav1.NewTime(time.Now())

	mockConfig := config.Mock(t)
	mockConfig.SetWithoutSource("cluster_name", testClusterName)

	pod1 := podWithConfigMapEnv(namespace, pod1Name, configMap1Name, configMap2Name, someTimestamp)
	pod2 := podWithSecretEnv(namespace, pod2Name, secret1Name, secret2Name, someTimestamp)
	configMap1 := configMap(namespace, configMap1Name, someTimestamp)
	configMap2 := configMap(namespace, configMap2Name, someTimestamp)
	secret1 := secret(namespace, secret1Name, someTimestamp)
	secret2 := secret(namespace, secret2Name, someTimestamp)
	node1 := node(node1Name, node1Provider, someTimestamp)

	components, relations := executeRelationCorrelation(t,
		[]coreV1.Pod{pod1, pod2},
		[]coreV1.ConfigMap{configMap1, configMap2},
		[]coreV1.Secret{secret1, secret2},
		[]coreV1.Node{node1},
	)

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
		nodeComponent(node1Name, instanceID, clusterName, someTimestamp),
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
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
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
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
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

func podComponent(namespace string, name string, timestamp metav1.Time) *topology.Component {
	return &topology.Component{
		ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, name),
		Type: topology.Type{
			Name: "pod",
		},
		Data: topology.Data{
			"name": name,
			"kind": "Pod",
			"tags": map[string]string{
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-pod",
				"namespace":      namespace,
			},
			"identifiers":       []string{},
			"creationTimestamp": timestamp,
			"uid":               types.UID(""),
			"restartPolicy":     coreV1.RestartPolicy(""),
			"status": coreV1.PodStatus{
				Phase:                      "",
				Conditions:                 nil,
				Message:                    "",
				Reason:                     "",
				NominatedNodeName:          "",
				HostIP:                     "",
				PodIP:                      "",
				PodIPs:                     nil,
				StartTime:                  nil,
				InitContainerStatuses:      nil,
				ContainerStatuses:          nil,
				QOSClass:                   "",
				EphemeralContainerStatuses: nil,
			},
		},
	}
}

func configMap(namespace string, name string, timestamp metav1.Time) coreV1.ConfigMap {
	return coreV1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind: "ConfigMap",
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
			"kind": "ConfigMap",
			"tags": map[string]string{
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-configmap",
				"namespace":      namespace,
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
			Kind: "Secret",
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

func node(name string, providerID string, timestamp metav1.Time) coreV1.Node {
	return coreV1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind: "Node",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: timestamp,
			UID:               types.UID(name),
			GenerateName:      "",
			ResourceVersion:   "123",
		},
		Status: coreV1.NodeStatus{
			Phase: coreV1.NodeRunning,
			NodeInfo: coreV1.NodeSystemInfo{
				MachineID:     "machineID",
				KernelVersion: "5.10.23",
			},
			DaemonEndpoints: coreV1.NodeDaemonEndpoints{
				KubeletEndpoint: coreV1.DaemonEndpoint{
					Port: 33,
				},
			},
		},
		Spec: coreV1.NodeSpec{
			ProviderID: providerID,
		},
	}
}
func nodeComponent(name, instanceID, clusterName string, timestamp metav1.Time) *topology.Component {
	hostname := fmt.Sprintf("%s-%s", name, clusterName)

	return &topology.Component{
		ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:node/%s", name),
		Type: topology.Type{
			Name: "node",
		},
		Data: topology.Data{
			"name":       name,
			"kind":       "Node",
			"instanceId": hostname,
			"sts_host":   hostname,
			"tags": map[string]string{
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-node",
			},
			"identifiers": []string{
				fmt.Sprintf("urn:host:/%s", hostname),
			},
			"creationTimestamp": timestamp,
			"uid":               types.UID(name),
			"status": NodeStatus{
				Phase: coreV1.NodeRunning,
				NodeInfo: coreV1.NodeSystemInfo{
					MachineID:               "machineID",
					SystemUUID:              "",
					BootID:                  "",
					KernelVersion:           "5.10.23",
					OSImage:                 "",
					ContainerRuntimeVersion: "",
					KubeletVersion:          "",
					KubeProxyVersion:        "",
					OperatingSystem:         "",
					Architecture:            "",
				},
				KubeletEndpoint: coreV1.DaemonEndpoint{
					Port: 33,
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
			"kind": "Secret",
			"tags": map[string]string{
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-secret",
				"namespace":      namespace,
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
	nodes []coreV1.Node,
) ([]*topology.Component, []*topology.Relation) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	clusterAPIClient := MockRelationCorrelatorAPIClient{
		pods: pods, configMaps: configMaps, secrets: secrets, nodes: nodes,
	}

	podCorrChannel := make(chan *PodLabelCorrelation)
	containerCorrChannel := make(chan *ContainerCorrelation)
	nodeIdentifierCorrChan := make(chan *NodeIdentifierCorrelation)
	volumeCorrChannel := make(chan *VolumeCorrelation)
	collectorsDoneChan := make(chan bool)
	correlatorsDoneChan := make(chan bool)
	relationCorrelationDoneChan := make(chan bool)

	// Pod correlation is just a no-op sink to assure progress
	go func() {
		for range podCorrChannel {
		}
	}()

	commonClusterCollector := NewTestCommonClusterCollector(clusterAPIClient, componentChannel, relationChannel, false, false)
	podCollector := NewPodCollector(
		containerCorrChannel, volumeCorrChannel,
		podCorrChannel,
		commonClusterCollector,
	)
	configMapCollector := NewConfigMapCollector(commonClusterCollector, TestMaxDataSize)
	secretCollector := NewSecretCollector(commonClusterCollector)
	nodeCollector := NewNodeCollector(nodeIdentifierCorrChan, commonClusterCollector)

	containerCorrelator := NewContainerCorrelator(nodeIdentifierCorrChan, containerCorrChannel, NewClusterTopologyCorrelator(commonClusterCollector))

	collectorsFinished := false

	go func() {
		containerCorrelator.CorrelateFunction()
		correlatorsDoneChan <- true
	}()

	go func() {
		var err error
		err = podCollector.CollectorFunction()
		assert.NoError(t, err)
		err = configMapCollector.CollectorFunction()
		assert.NoError(t, err)
		err = secretCollector.CollectorFunction()
		assert.NoError(t, err)
		err = nodeCollector.CollectorFunction()
		assert.NoError(t, err)

		collectorsFinished = true
		collectorsDoneChan <- true
	}()

	go func() {
		<-collectorsDoneChan
		<-correlatorsDoneChan
		commonClusterCollector.CorrelateRelations()
		relationCorrelationDoneChan <- true
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
		case <-relationCorrelationDoneChan:
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
	nodes      []coreV1.Node
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

func (m MockRelationCorrelatorAPIClient) GetNodes() ([]coreV1.Node, error) {
	return m.nodes, nil
}
