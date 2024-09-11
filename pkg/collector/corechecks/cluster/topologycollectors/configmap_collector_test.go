// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	TestMaxDataSize   = 120
	TestKey3Length    = 500
	lastAppliedConfig = `{"apiVersion":"v1","data":{"logback.groovy":"..."},"kind":"ConfigMap","metadata":{"annotations":{"argocd.argoproj.io/tracking-id":"saas-tenants:/ConfigMap:tenant/log"},"labels":{"app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"app","app.kubernetes.io/version":"1.0.0","helm.sh/chart":"1.0.0"},"name":"log","namespace":"tenant"}}`
)

func TestConfigMapCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)
	expectedCollectorName := "ConfigMap Collector"

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			cmc := NewConfigMapCollector(
				NewTestCommonClusterCollector(MockConfigMapAPICollectorClient{}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled),
				TestMaxDataSize,
			)
			RunCollectorTest(t, cmc, expectedCollectorName)

			for _, tc := range []struct {
				testCase           string
				expectedNoSP       *topology.Component
				expectedSP         *topology.Component
				expectedKubeStatus *topology.Component
			}{
				{
					testCase: "Test ConfigMap 1 - Complete",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-1",
						Type:       topology.Type{Name: "configmap"},
						Data: topology.Data{
							"name":              "test-configmap-1",
							"kind":              "ConfigMap",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-configmap",
								"namespace":      "test-namespace",
							},
							"uid": types.UID("test-configmap-1"),
							"data": map[string]string{
								"key1": "value1",
								"key2": "longersecretvalue2",
								"key3": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA[dropped 460 chars, hashsum: 828798a87da42aa9]",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-1"},
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-1",
						Type:       topology.Type{Name: "configmap"},
						Data: topology.Data{
							"name": "test-configmap-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-configmap",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-1"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-configmap-1",
								"namespace":         "test-namespace",
								"uid":               "test-configmap-1",
							},
							"data": map[string]interface{}{
								"key1": "value1",
								"key2": "longersecretvalue2",
								"key3": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA[dropped 460 chars, hashsum: 828798a87da42aa9]",
							},
						},
					},
					expectedKubeStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-1",
						Type:       topology.Type{Name: "configmap"},
						Data: topology.Data{
							"name": "test-configmap-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-configmap",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-1"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-configmap-1",
								"namespace":         "test-namespace",
								"uid":               "test-configmap-1",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfig},
							},
							"data": map[string]interface{}{
								"key1": "value1",
								"key2": "longersecretvalue2",
								"key3": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA[dropped 460 chars, hashsum: 828798a87da42aa9]",
							},
						},
					},
				},
				{
					testCase: "Test ConfigMap 2 - Without Data",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-2",
						Type:       topology.Type{Name: "configmap"},
						Data: topology.Data{
							"name":              "test-configmap-2",
							"kind":              "ConfigMap",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-configmap",
								"namespace":      "test-namespace",
							},
							"uid":         types.UID("test-configmap-2"),
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-2"},
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-2",
						Type:       topology.Type{Name: "configmap"},
						Data: topology.Data{
							"name": "test-configmap-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-configmap",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-2"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-configmap-2",
								"namespace":         "test-namespace",
								"uid":               "test-configmap-2",
							},
						},
					},
					expectedKubeStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-2",
						Type:       topology.Type{Name: "configmap"},
						Data: topology.Data{
							"name": "test-configmap-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-configmap",
								"namespace":      "test-namespace"},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-2"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-configmap-2",
								"namespace":         "test-namespace",
								"uid":               "test-configmap-2",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfig},
							},
						},
					},
				},
				{
					testCase: "Test ConfigMap 3 - Minimal",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-3",
						Type:       topology.Type{Name: "configmap"},
						Data: topology.Data{
							"name":              "test-configmap-3",
							"kind":              "ConfigMap",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-configmap",
								"namespace":      "test-namespace",
							},
							"uid":         types.UID("test-configmap-3"),
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-3"},
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-3",
						Type:       topology.Type{Name: "configmap"},
						Data: topology.Data{
							"name": "test-configmap-3",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-configmap",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-3"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"name":              "test-configmap-3",
								"namespace":         "test-namespace",
								"uid":               "test-configmap-3",
							},
						},
					},
					expectedKubeStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-3",
						Type:       topology.Type{Name: "configmap"},
						Data: topology.Data{
							"name": "test-configmap-3",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-configmap",
								"namespace":      "test-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-3"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"name":              "test-configmap-3",
								"namespace":         "test-namespace",
								"uid":               "test-configmap-3",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfig},
							},
						},
					},
				},
			} {
				t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
					component := <-componentChannel
					if sourcePropertiesEnabled {
						if kubernetesStatusEnabled {
							assert.EqualValues(t, tc.expectedKubeStatus, component)
						} else {
							assert.EqualValues(t, tc.expectedSP, component)
						}
					} else {
						assert.EqualValues(t, tc.expectedNoSP, component)
					}
				})
			}
		}
	}
}

type MockConfigMapAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockConfigMapAPICollectorClient) GetConfigMaps() ([]coreV1.ConfigMap, error) {
	configMaps := make([]coreV1.ConfigMap, 0)
	for i := 1; i <= 3; i++ {

		configMap := coreV1.ConfigMap{
			TypeMeta: v1.TypeMeta{
				Kind: "ConfigMap",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-configmap-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				UID:               types.UID(fmt.Sprintf("test-configmap-%d", i)),
				GenerateName:      "",
				ResourceVersion:   "123",
				ManagedFields: []v1.ManagedFieldsEntry{
					{
						Manager:    "ignored",
						Operation:  "Updated",
						APIVersion: "whatever",
						Time:       &v1.Time{Time: time.Now()},
						FieldsType: "whatever",
					},
				},
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfig,
				},
			},
		}

		if i == 1 {
			configMap.Data = map[string]string{
				"key1": "value1",
				"key2": "longersecretvalue2",
				"key3": strings.Repeat("A", TestKey3Length),
			}
		}

		if i != 3 {
			configMap.Labels = map[string]string{
				"test": "label",
			}
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

func TestCutDataProportionally1(t *testing.T) {
	result := cutData(map[string]string{
		"a": strings.Repeat("A", 30),
	}, 20)

	assert.EqualValues(t, map[string]string{
		"a": "AAAAAAAAAAAAAAAAAAAA[dropped 10 chars, hashsum: 1d65bf29403e4fb1]",
	}, result)
}

func TestCutDataProportionally2(t *testing.T) {
	result := cutData(map[string]string{
		"a": strings.Repeat("A", 11),
		"b": strings.Repeat("B", 22),
	}, 30)

	assert.EqualValues(t, map[string]string{
		"a": "AAAAAAAAAAA",
		"b": "BBBBBBBBBBBBBBB[dropped 7 chars, hashsum: f4205e933dd99030]",
	}, result)
}
