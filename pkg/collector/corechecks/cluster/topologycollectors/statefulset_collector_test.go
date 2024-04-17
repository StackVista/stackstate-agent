// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver

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

var lastAppliedConfigurationStatefulset = `{"apiVersion":"apps/v1","kind":"StatefulSet","metadata":{"annotations":{"argocd.io/tracking-id":"tenant"},"labels":{"app.kubernetes.io/component":"api","app.kubernetes.io/instance":"app","app.kubernetes.io/managed-by":"Helm","helm.sh/chart":"1.0.0"},"name":"test-api","namespace":"tenant"},"spec":{"podManagementPolicy":"Parallel","replicas":1,"selector":{"matchLabels":{"app.kubernetes.io/component":"test","app.kubernetes.io/instance":"app","app.kubernetes.io/name":"test"}},"serviceName":"api-headless","template":{"spec":{"containers":[{"command":["bash","-ec"]}]}}}}`

func TestStatefulSetCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	replicas = int32(1)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			commonClusterCollector := NewTestCommonClusterCollector(MockStatefulSetAPICollectorClient{}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled)
			commonClusterCollector.SetUseRelationCache(false)
			cmc := NewStatefulSetCollector(commonClusterCollector)
			expectedCollectorName := "StatefulSet Collector"
			RunCollectorTest(t, cmc, expectedCollectorName)

			for _, tc := range []struct {
				testCase             string
				expectedNoSP         *topology.Component
				expectedSP           *topology.Component
				expectedSPPlusStatus *topology.Component
			}{
				{
					testCase: "Test StatefulSet 1",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/test-statefulset-1",
						Type:       topology.Type{Name: "statefulset"},
						Data: topology.Data{
							"name":              "test-statefulset-1",
							"kind":              "StatefulSet",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-statefulset",
								"namespace":      "test-namespace",
							},
							"uid":                 types.UID("test-statefulset-1"),
							"updateStrategy":      appsV1.RollingUpdateStatefulSetStrategyType,
							"desiredReplicas":     &replicas,
							"podManagementPolicy": appsV1.OrderedReadyPodManagement,
							"serviceName":         "statefulset-service-name",
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/test-statefulset-1",
						Type:       topology.Type{Name: "statefulset"},
						Data: topology.Data{
							"name": "test-statefulset-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-statefulset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "StatefulSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-statefulset-1",
								"namespace":         "test-namespace",
								"uid":               "test-statefulset-1"},
							"spec": map[string]interface{}{
								"podManagementPolicy": "OrderedReady",
								"replicas":            float64(1),
								"selector":            nil,
								"serviceName":         "statefulset-service-name",
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/test-statefulset-1",
						Type:       topology.Type{Name: "statefulset"},
						Data: topology.Data{
							"name": "test-statefulset-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-statefulset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "StatefulSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-statefulset-1",
								"namespace":         "test-namespace",
								"uid":               "test-statefulset-1",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationStatefulset},
							},
							"spec": map[string]interface{}{
								"podManagementPolicy": "OrderedReady",
								"replicas":            float64(1),
								"selector":            nil,
								"serviceName":         "statefulset-service-name",
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
							},
							"status": map[string]interface{}{
								"observedGeneration": float64(123),
								"replicas":           float64(1),
								"readyReplicas":      float64(1),
								"currentReplicas":    float64(1),
								"updatedReplicas":    float64(1),
								"currentRevision":    "abc-112233d",
								"updateRevision":     "abc-112233d",
							},
						},
					},
				},
				{
					testCase: "Test StatefulSet 2",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/test-statefulset-2",
						Type:       topology.Type{Name: "statefulset"},
						Data: topology.Data{
							"name":              "test-statefulset-2",
							"kind":              "StatefulSet",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-statefulset",
								"namespace":      "test-namespace",
							},
							"uid":                 types.UID("test-statefulset-2"),
							"updateStrategy":      appsV1.RollingUpdateStatefulSetStrategyType,
							"desiredReplicas":     &replicas,
							"podManagementPolicy": appsV1.OrderedReadyPodManagement,
							"serviceName":         "statefulset-service-name",
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/test-statefulset-2",
						Type:       topology.Type{Name: "statefulset"},
						Data: topology.Data{
							"name": "test-statefulset-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-statefulset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "StatefulSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-statefulset-2",
								"namespace":         "test-namespace",
								"uid":               "test-statefulset-2"},
							"spec": map[string]interface{}{
								"podManagementPolicy": "OrderedReady",
								"replicas":            float64(1),
								"selector":            nil,
								"serviceName":         "statefulset-service-name",
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/test-statefulset-2",
						Type:       topology.Type{Name: "statefulset"},
						Data: topology.Data{
							"name": "test-statefulset-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-statefulset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "StatefulSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-statefulset-2",
								"namespace":         "test-namespace",
								"uid":               "test-statefulset-2",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationStatefulset},
							},
							"spec": map[string]interface{}{
								"podManagementPolicy": "OrderedReady",
								"replicas":            float64(1),
								"selector":            nil,
								"serviceName":         "statefulset-service-name",
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
							},
							"status": map[string]interface{}{
								"observedGeneration": float64(123),
								"replicas":           float64(1),
								"readyReplicas":      float64(1),
								"currentReplicas":    float64(1),
								"updatedReplicas":    float64(1),
								"currentRevision":    "abc-112233d",
								"updateRevision":     "abc-112233d",
							},
						},
					},
				},
				{
					testCase: "Test StatefulSet 3 - Kind + Generate Name",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/test-statefulset-3",
						Type:       topology.Type{Name: "statefulset"},
						Data: topology.Data{
							"name":              "test-statefulset-3",
							"kind":              "StatefulSet",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-statefulset",
								"namespace":      "test-namespace",
							},
							"uid":                 types.UID("test-statefulset-3"),
							"generateName":        "some-specified-generation",
							"updateStrategy":      appsV1.RollingUpdateStatefulSetStrategyType,
							"desiredReplicas":     &replicas,
							"podManagementPolicy": appsV1.OrderedReadyPodManagement,
							"serviceName":         "statefulset-service-name",
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/test-statefulset-3",
						Type:       topology.Type{Name: "statefulset"},
						Data: topology.Data{
							"name": "test-statefulset-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-statefulset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "StatefulSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-statefulset-3",
								"namespace":         "test-namespace",
								"generateName":      "some-specified-generation",
								"uid":               "test-statefulset-3"},
							"spec": map[string]interface{}{
								"podManagementPolicy": "OrderedReady",
								"replicas":            float64(1),
								"selector":            nil,
								"serviceName":         "statefulset-service-name",
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/test-statefulset-3",
						Type:       topology.Type{Name: "statefulset"},
						Data: topology.Data{
							"name": "test-statefulset-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-statefulset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "StatefulSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-statefulset-3",
								"namespace":         "test-namespace",
								"generateName":      "some-specified-generation",
								"uid":               "test-statefulset-3",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationStatefulset},
							},
							"spec": map[string]interface{}{
								"podManagementPolicy": "OrderedReady",
								"replicas":            float64(1),
								"selector":            nil,
								"serviceName":         "statefulset-service-name",
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
							},
							"status": map[string]interface{}{
								"observedGeneration": float64(123),
								"replicas":           float64(1),
								"readyReplicas":      float64(1),
								"currentReplicas":    float64(1),
								"updatedReplicas":    float64(1),
								"currentRevision":    "abc-112233d",
								"updateRevision":     "abc-112233d",
							},
						},
					},
				},
			} {
				t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
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

					actualRelation := <-relationChannel
					expectedRelation := &topology.Relation{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" + component.ExternalID,
						Type:       topology.Type{Name: "encloses"},
						SourceID:   "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
						TargetID:   component.ExternalID,
						Data:       map[string]interface{}{},
					}
					assert.EqualValues(t, expectedRelation, actualRelation)
				})
			}
		}
	}
}

type MockStatefulSetAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockStatefulSetAPICollectorClient) GetStatefulSets() ([]appsV1.StatefulSet, error) {
	statefulSets := make([]appsV1.StatefulSet, 0)
	for i := 1; i <= 3; i++ {
		statefulSet := appsV1.StatefulSet{
			TypeMeta: v1.TypeMeta{
				Kind: "StatefulSet",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-statefulset-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:             types.UID(fmt.Sprintf("test-statefulset-%d", i)),
				GenerateName:    "",
				ResourceVersion: "123",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationStatefulset,
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
			Spec: appsV1.StatefulSetSpec{
				UpdateStrategy: appsV1.StatefulSetUpdateStrategy{
					Type: appsV1.RollingUpdateStatefulSetStrategyType,
				},
				Replicas:            &replicas,
				PodManagementPolicy: appsV1.OrderedReadyPodManagement,
				ServiceName:         "statefulset-service-name",
			},
			Status: appsV1.StatefulSetStatus{
				ObservedGeneration: int64(123),
				Replicas:           int32(1),
				ReadyReplicas:      int32(1),
				CurrentReplicas:    int32(1),
				UpdatedReplicas:    int32(1),
				CurrentRevision:    "abc-112233d",
				UpdateRevision:     "abc-112233d",
			},
		}

		if i == 3 {
			statefulSet.ObjectMeta.GenerateName = "some-specified-generation"
		}

		statefulSets = append(statefulSets, statefulSet)
	}

	return statefulSets, nil
}
