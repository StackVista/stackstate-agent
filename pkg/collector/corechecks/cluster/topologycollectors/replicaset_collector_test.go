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
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestReplicaSetCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)
	replicas = 1

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			commonClusterCollector := NewTestCommonClusterCollector(MockReplicaSetAPICollectorClient{}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled)
			commonClusterCollector.SetUseRelationCache(false)
			ic := NewReplicaSetCollector(commonClusterCollector)
			expectedCollectorName := "ReplicaSet Collector"
			RunCollectorTest(t, ic, expectedCollectorName)

			for _, tc := range []struct {
				testCase                      string
				expectedComponentSP           *topology.Component
				expectedComponentNoSP         *topology.Component
				expectedComponentSPPlusStatus *topology.Component
				expectedRelations             []*topology.Relation
			}{
				{
					testCase: "Test ReplicaSet 1 - Minimal",
					expectedComponentNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-1",
						Type:       topology.Type{Name: "replicaset"},
						Data: topology.Data{
							"name":              "test-replicaset-1",
							"kind":              "ReplicaSet",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-replicaset",
								"namespace":      "test-namespace",
							},
							"uid":             types.UID("test-replicaset-1"),
							"desiredReplicas": &replicas,
						},
					},
					expectedComponentSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-1",
						Type:       topology.Type{Name: "replicaset"},
						Data: topology.Data{
							"name": "test-replicaset-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-replicaset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "ReplicaSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-replicaset-1",
								"namespace":         "test-namespace",
								"uid":               "test-replicaset-1"},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								}},
						},
					},
					expectedComponentSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-1",
						Type:       topology.Type{Name: "replicaset"},
						Data: topology.Data{
							"name": "test-replicaset-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-replicaset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "ReplicaSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-replicaset-1",
								"namespace":         "test-namespace",
								"uid":               "test-replicaset-1",
								"resourceVersion":   "123",
							},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
							"status": map[string]interface{}{
								"replicas":           float64(1),
								"availableReplicas":  float64(1),
								"readyReplicas":      float64(1),
								"observedGeneration": float64(123),
							},
						},
					},
					expectedRelations: []*topology.Relation{
						{
							ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-1",
							Type:       topology.Type{Name: "encloses"},
							SourceID:   "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
							TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-1",
							Data:       map[string]interface{}{},
						},
					},
				},
				{
					testCase: "Test ReplicaSet 2 - Kind + Generate Name",
					expectedComponentNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-2",
						Type:       topology.Type{Name: "replicaset"},
						Data: topology.Data{
							"name":              "test-replicaset-2",
							"kind":              "ReplicaSet",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-replicaset",
								"namespace":      "test-namespace",
							},
							"uid":             types.UID("test-replicaset-2"),
							"desiredReplicas": &replicas,
							"generateName":    "some-specified-generation",
						},
					},
					expectedComponentSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-2",
						Type:       topology.Type{Name: "replicaset"},
						Data: topology.Data{
							"name": "test-replicaset-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-replicaset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "ReplicaSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-replicaset-2",
								"namespace":         "test-namespace",
								"generateName":      "some-specified-generation",
								"uid":               "test-replicaset-2"},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								}},
						},
					},
					expectedComponentSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-2",
						Type:       topology.Type{Name: "replicaset"},
						Data: topology.Data{
							"name": "test-replicaset-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-replicaset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "ReplicaSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-replicaset-2",
								"namespace":         "test-namespace",
								"generateName":      "some-specified-generation",
								"uid":               "test-replicaset-2",
								"resourceVersion":   "123",
							},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
							"status": map[string]interface{}{
								"replicas":           float64(1),
								"availableReplicas":  float64(1),
								"readyReplicas":      float64(1),
								"observedGeneration": float64(123),
							},
						},
					},
					expectedRelations: []*topology.Relation{
						{
							ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-2",
							Type:       topology.Type{Name: "encloses"},
							SourceID:   "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
							TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-2",
							Data:       map[string]interface{}{},
						},
					},
				},
				{
					testCase: "Test ReplicaSet 3 - Complete",
					expectedComponentNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-3",
						Type:       topology.Type{Name: "replicaset"},
						Data: topology.Data{
							"name":              "test-replicaset-3",
							"kind":              "ReplicaSet",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-replicaset",
								"namespace":      "test-namespace",
							},
							"uid":             types.UID("test-replicaset-3"),
							"desiredReplicas": &replicas,
							"generateName":    "some-specified-generation",
						},
					},
					expectedComponentSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-3",
						Type:       topology.Type{Name: "replicaset"},
						Data: topology.Data{
							"name": "test-replicaset-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-replicaset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "ReplicaSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-replicaset-3",
								"namespace":         "test-namespace",
								"generateName":      "some-specified-generation",
								"ownerReferences":   []interface{}{map[string]interface{}{"apiVersion": "", "kind": "Deployment", "name": "test-deployment-3", "uid": ""}},
								"uid":               "test-replicaset-3"},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								}},
						},
					},
					expectedComponentSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-3",
						Type:       topology.Type{Name: "replicaset"},
						Data: topology.Data{
							"name": "test-replicaset-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-replicaset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "ReplicaSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-replicaset-3",
								"namespace":         "test-namespace",
								"generateName":      "some-specified-generation",
								"ownerReferences":   []interface{}{map[string]interface{}{"apiVersion": "", "kind": "Deployment", "name": "test-deployment-3", "uid": ""}},
								"uid":               "test-replicaset-3",
								"resourceVersion":   "123",
							},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								}},
							"status": map[string]interface{}{
								"replicas":           float64(1),
								"availableReplicas":  float64(1),
								"readyReplicas":      float64(1),
								"observedGeneration": float64(123),
							},
						},
					},
					expectedRelations: []*topology.Relation{
						{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-3->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-3",
							Type:     topology.Type{Name: "controls"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-3",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-3",
							Data:     map[string]interface{}{},
						},
					},
				},
			} {
				t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
					service := <-componentChannel
					if sourcePropertiesEnabled {
						if kubernetesStatusEnabled {
							assert.EqualValues(t, tc.expectedComponentSPPlusStatus, service)
						} else {
							assert.EqualValues(t, tc.expectedComponentSP, service)
						}
					} else {
						assert.EqualValues(t, tc.expectedComponentNoSP, service)
					}
					for _, expectedRelation := range tc.expectedRelations {
						serviceRelation := <-relationChannel
						assert.EqualValues(t, expectedRelation, serviceRelation)
					}
				})
			}
		}
	}
}

type MockReplicaSetAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockReplicaSetAPICollectorClient) GetReplicaSets() ([]appsV1.ReplicaSet, error) {
	replicaSets := make([]appsV1.ReplicaSet, 0)
	for i := 1; i <= 3; i++ {
		replicaSet := appsV1.ReplicaSet{
			TypeMeta: v1.TypeMeta{
				Kind: "ReplicaSet",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-replicaset-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:             types.UID(fmt.Sprintf("test-replicaset-%d", i)),
				GenerateName:    "",
				ResourceVersion: "123",
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
			Spec: appsV1.ReplicaSetSpec{
				Replicas: &replicas,
			},
			Status: appsV1.ReplicaSetStatus{
				Replicas:           int32(1),
				AvailableReplicas:  int32(1),
				ReadyReplicas:      int32(1),
				ObservedGeneration: int64(123),
			},
		}

		if i > 1 {
			replicaSet.ObjectMeta.GenerateName = "some-specified-generation"
		}

		if i == 3 {
			replicaSet.OwnerReferences = []v1.OwnerReference{
				{Kind: "Deployment", Name: "test-deployment-3"},
			}
		}

		replicaSets = append(replicaSets, replicaSet)
	}

	return replicaSets, nil
}
