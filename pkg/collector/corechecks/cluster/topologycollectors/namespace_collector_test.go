// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"testing"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestNamespaceCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	componentIDChannel := make(chan string)
	defer close(componentIDChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	for _, sourcePropertiesEnabled := range []bool{false, true} {

		nsc := NewNamespaceCollector(NewTestCommonClusterCollector(MockNamespaceAPICollectorClient{}, componentChannel, componentIDChannel, sourcePropertiesEnabled))
		expectedCollectorName := "Namespace Collector"
		RunCollectorTest(t, nsc, expectedCollectorName)

		for _, tc := range []struct {
			testCase     string
			expectedSP   *topology.Component
			expectedNoSP *topology.Component
		}{
			{
				testCase: "Test Namespace 1 - Complete",
				expectedNoSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace-1",
					Type:       topology.Type{Name: "namespace"},
					Data: topology.Data{
						"name":              "test-namespace-1",
						"creationTimestamp": creationTime,
						"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name"},
						"uid":               types.UID("test-namespace-1"),
						"identifiers":       []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-1"},
					},
				},
				expectedSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace-1",
					Type:       topology.Type{Name: "namespace"},
					Data: topology.Data{
						"name":        "test-namespace-1",
						"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name"},
						"identifiers": []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-1"},
					},
					SourceProperties: map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
							"labels":            map[string]interface{}{"test": "label"},
							"name":              "test-namespace-1",
							"uid":               "test-namespace-1",
						},
						"spec": map[string]interface{}{},
					},
				},
			},
			{
				testCase: "Test Namespace 2 - Minimal",
				expectedNoSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace-2",
					Type:       topology.Type{Name: "namespace"},
					Data: topology.Data{
						"name":              "test-namespace-2",
						"creationTimestamp": creationTime,
						"tags":              map[string]string{"cluster-name": "test-cluster-name"},
						"uid":               types.UID("test-namespace-2"),
						"identifiers":       []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-2"},
					},
				},
				expectedSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace-2",
					Type:       topology.Type{Name: "namespace"},
					Data: topology.Data{
						"name":        "test-namespace-2",
						"tags":        map[string]string{"cluster-name": "test-cluster-name"},
						"identifiers": []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-2"},
					},
					SourceProperties: map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
							"name":              "test-namespace-2",
							"uid":               "test-namespace-2",
						},
						"spec": map[string]interface{}{},
					},
				},
			},
		} {
			t.Run(tc.testCase, func(t *testing.T) {
				component := <-componentChannel
				<-componentIDChannel
				if sourcePropertiesEnabled {
					assert.EqualValues(t, tc.expectedSP, component)
				} else {
					assert.EqualValues(t, tc.expectedNoSP, component)
				}
			})
		}
	}
}

type MockNamespaceAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockNamespaceAPICollectorClient) GetNamespaces() ([]coreV1.Namespace, error) {
	namespaces := make([]coreV1.Namespace, 0)
	for i := 1; i <= 2; i++ {

		namespace := coreV1.Namespace{
			TypeMeta: v1.TypeMeta{
				Kind: "",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-namespace-%d", i),
				CreationTimestamp: creationTime,
				UID:               types.UID(fmt.Sprintf("test-namespace-%d", i)),
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
			},
		}

		if i != 2 {
			namespace.Labels = map[string]string{
				"test": "label",
			}
		}

		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}
