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

	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var lastAppliedConfigurationNamespace = `{"apiVersion":"v1","kind":"Namespace","metadata":{"annotations":{"argocd.io/tracking-id":"tenant"},"labels":{"name":"test"},"name":"test"},"spec":{"finalizers":["kubernetes"]}}`

func TestNamespaceCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {

			nsc := NewNamespaceCollector(NewTestCommonClusterCollector(MockNamespaceAPICollectorClient{}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled))
			expectedCollectorName := "Namespace Collector"
			RunCollectorTest(t, nsc, expectedCollectorName)

			for _, tc := range []struct {
				testCase             string
				expectedSP           *topology.Component
				expectedNoSP         *topology.Component
				expectedSPPlusStatus *topology.Component
			}{
				{
					testCase: "Test Namespace 1 - Complete",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace-1",
						Type:       topology.Type{Name: "namespace"},
						Data: topology.Data{
							"name":              "test-namespace-1",
							"kind":              "Namespace",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-namespace",
							},
							"uid":         types.UID("test-namespace-1"),
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-1"},
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace-1",
						Type:       topology.Type{Name: "namespace"},
						Data: topology.Data{
							"name": "test-namespace-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-1"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Namespace",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-namespace-1",
								"uid":               "test-namespace-1",
							},
							"spec": map[string]interface{}{},
							"status": map[string]interface{}{
								"phase": "Active",
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace-1",
						Type:       topology.Type{Name: "namespace"},
						Data: topology.Data{
							"name": "test-namespace-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-1"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Namespace",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-namespace-1",
								"uid":               "test-namespace-1",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationNamespace},
							},
							"spec": map[string]interface{}{},
							"status": map[string]interface{}{
								"phase": "Active",
							},
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
							"kind":              "Namespace",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-namespace",
							},
							"uid":         types.UID("test-namespace-2"),
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-2"},
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace-2",
						Type:       topology.Type{Name: "namespace"},
						Data: topology.Data{
							"name": "test-namespace-2",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-2"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Namespace",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"name":              "test-namespace-2",
								"uid":               "test-namespace-2",
							},
							"spec": map[string]interface{}{},
							"status": map[string]interface{}{
								"phase": "Active",
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace-2",
						Type:       topology.Type{Name: "namespace"},
						Data: topology.Data{
							"name": "test-namespace-2",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-namespace",
							},
							"identifiers": []string{"urn:kubernetes:/test-cluster-name:namespace/test-namespace-2"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Namespace",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"name":              "test-namespace-2",
								"uid":               "test-namespace-2",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationNamespace},
							},
							"spec": map[string]interface{}{},
							"status": map[string]interface{}{
								"phase": "Active",
							},
						},
					},
				},
			} {
				t.Run(tc.testCase, func(t *testing.T) {
					component := <-componentChannel
					if sourcePropertiesEnabled {
						if kubernetesStatusEnabled {
							assert.EqualValues(t, tc.expectedSPPlusStatus, component)
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

type MockNamespaceAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockNamespaceAPICollectorClient) GetNamespaces() ([]coreV1.Namespace, error) {
	namespaces := make([]coreV1.Namespace, 0)
	for i := 1; i <= 2; i++ {

		namespace := coreV1.Namespace{
			TypeMeta: v1.TypeMeta{
				Kind: "Namespace",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-namespace-%d", i),
				CreationTimestamp: creationTime,
				UID:               types.UID(fmt.Sprintf("test-namespace-%d", i)),
				GenerateName:      "",
				ResourceVersion:   "123",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationNamespace,
				},
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
			Spec: coreV1.NamespaceSpec{},
			Status: coreV1.NamespaceStatus{
				Phase: "Active",
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
