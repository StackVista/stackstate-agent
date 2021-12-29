// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
	"testing"
	"time"
)

func TestConfigMapCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}

	cmc := NewConfigMapCollector(componentChannel, NewTestCommonClusterCollector(MockConfigMapAPICollectorClient{}), 500)
	expectedCollectorName := "ConfigMap Collector"
	RunCollectorTest(t, cmc, expectedCollectorName)

	for _, tc := range []struct {
		testCase string
		expected *topology.Component
	}{
		{
			testCase: "Test ConfigMap 1 - Complete",
			expected: &topology.Component{
				ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-1",
				Type:       topology.Type{Name: "configmap"},
				Data: topology.Data{
					"name":              "test-configmap-1",
					"creationTimestamp": creationTime,
					"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					"uid":               types.UID("test-configmap-1"),
					"data":              map[string]string{"key1": "value1", "key2": "longersecretvalue2", "key3": "[dropped 1000 chars, hashsum: c2e686823489ced2]"},
					"identifiers":       []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-1"},
				},
			},
		},
		{
			testCase: "Test ConfigMap 2 - Without Data",
			expected: &topology.Component{
				ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-2",
				Type:       topology.Type{Name: "configmap"},
				Data: topology.Data{
					"name":              "test-configmap-2",
					"creationTimestamp": creationTime,
					"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					"uid":               types.UID("test-configmap-2"),
					"identifiers":       []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-2"},
				},
			},
		},
		{
			testCase: "Test ConfigMap 3 - Minimal",
			expected: &topology.Component{
				ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-3",
				Type:       topology.Type{Name: "configmap"},
				Data: topology.Data{
					"name":              "test-configmap-3",
					"creationTimestamp": creationTime,
					"tags":              map[string]string{"cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					"uid":               types.UID("test-configmap-3"),
					"identifiers":       []string{"urn:kubernetes:/test-cluster-name:test-namespace:configmap/test-configmap-3"},
				},
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			component := <-componentChannel
			assert.EqualValues(t, tc.expected, component)
		})
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
				Kind: "",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-configmap-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				UID:               types.UID(fmt.Sprintf("test-configmap-%d", i)),
				GenerateName:      "",
			},
		}

		if i == 1 {
			configMap.Data = map[string]string{
				"key1": "value1",
				"key2": "longersecretvalue2",
				"key3": strings.Repeat("A", 1000),
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

func TestCutDataProportionally(t *testing.T) {
	result := cutData(map[string]string{
		"a": strings.Repeat("A", 11),
		"b": strings.Repeat("B", 22),
	}, 30)

	assert.EqualValues(t, map[string]string{
		"a": strings.Repeat("A", 10) + "[dropped 1 chars, hashsum: 559aead08264d579]",
		"b": strings.Repeat("B", 20) + "[dropped 2 chars, hashsum: fc686c314491e1f6]",
	}, result)
}

func TestCutDataBiggest(t *testing.T) {
	// if configmap has TLS certificate and some actual configuration,
	// it is very likely that cutting TLS certificate is enough,
	// in the same time config untouched makes sense
	// let's keep keys that have small strings untouched
	result := cutData(map[string]string{
		"root.crt":      strings.Repeat("A", 100000),
		"blacklist.txt": strings.Repeat("B", 150),
		"nginx.conf":    strings.Repeat("C", 150),
	}, 300)

	assert.EqualValues(t, map[string]string{
		"root.crt":      "[dropped 100000 chars, hashsum: e6631225e83d23bf]",
		"blacklist.txt": strings.Repeat("B", 150),
		"nginx.conf":    strings.Repeat("C", 150),
	}, result)
}

func TestCutDataCombined(t *testing.T) {
	// if configmap has TLS certificate and some actual configuration,
	// it is very likely that cutting TLS certificate is enough,
	// in the same time config untouched makes sense
	// let's keep keys that have small strings untouched
	result := cutData(map[string]string{
		"root.crt": strings.Repeat("A", 100000),
		"a":        strings.Repeat("A", 11),
		"b":        strings.Repeat("B", 22),
	}, 30)

	assert.EqualValues(t, map[string]string{
		"root.crt": "[dropped 100000 chars, hashsum: e6631225e83d23bf]",
		"a":        strings.Repeat("A", 10) + "[dropped 1 chars, hashsum: 559aead08264d579]",
		"b":        strings.Repeat("B", 20) + "[dropped 2 chars, hashsum: fc686c314491e1f6]",
	}, result)
}
