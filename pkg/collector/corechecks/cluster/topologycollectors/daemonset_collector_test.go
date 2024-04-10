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

func TestDaemonSetCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			commonClusterCollector := NewTestCommonClusterCollector(MockDaemonSetAPICollectorClient{}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled)
			commonClusterCollector.SetUseRelationCache(false)
			cmc := NewDaemonSetCollector(commonClusterCollector)
			expectedCollectorName := "DaemonSet Collector"
			RunCollectorTest(t, cmc, expectedCollectorName)

			for _, tc := range []struct {
				testCase           string
				expectedNoSP       *topology.Component
				expectedSP         *topology.Component
				expectedKubeStatus *topology.Component
			}{
				{
					testCase: "Test DaemonSet 1",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/test-daemonset-1",
						Type:       topology.Type{Name: "daemonset"},
						Data: topology.Data{
							"name":              "test-daemonset-1",
							"kind":              "DaemonSet",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-daemonset",
								"namespace":      "test-namespace",
							},
							"uid":            types.UID("test-daemonset-1"),
							"updateStrategy": appsV1.RollingUpdateDaemonSetStrategyType,
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/test-daemonset-1",
						Type:       topology.Type{Name: "daemonset"},
						Data: topology.Data{
							"name": "test-daemonset-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-daemonset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "DaemonSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-daemonset-1",
								"namespace":         "test-namespace",
								"uid":               "test-daemonset-1",
							},
							"spec": map[string]interface{}{
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"selector": nil,
							},
						},
					},
					expectedKubeStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/test-daemonset-1",
						Type:       topology.Type{Name: "daemonset"},
						Data: topology.Data{
							"name": "test-daemonset-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-daemonset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "DaemonSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-daemonset-1",
								"namespace":         "test-namespace",
								"uid":               "test-daemonset-1",
								"resourceVersion":   "123",
							},
							"spec": map[string]interface{}{
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"selector": nil,
							},
							"status": map[string]interface{}{
								"currentNumberScheduled": float64(1),
								"desiredNumberScheduled": float64(1),
								"numberAvailable":        float64(1),
								"numberMisscheduled":     float64(1),
								"numberReady":            float64(1),
							},
						},
					},
				},
				{
					testCase: "Test DaemonSet 2",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/test-daemonset-2",
						Type:       topology.Type{Name: "daemonset"},
						Data: topology.Data{
							"name":              "test-daemonset-2",
							"kind":              "DaemonSet",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-daemonset",
								"namespace":      "test-namespace",
							},
							"uid":            types.UID("test-daemonset-2"),
							"updateStrategy": appsV1.RollingUpdateDaemonSetStrategyType,
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/test-daemonset-2",
						Type:       topology.Type{Name: "daemonset"},
						Data: topology.Data{
							"name": "test-daemonset-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-daemonset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "DaemonSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-daemonset-2",
								"namespace":         "test-namespace",
								"uid":               "test-daemonset-2",
							},
							"spec": map[string]interface{}{
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"selector": nil,
							},
						},
					},
					expectedKubeStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/test-daemonset-2",
						Type:       topology.Type{Name: "daemonset"},
						Data: topology.Data{
							"name": "test-daemonset-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-daemonset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "DaemonSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-daemonset-2",
								"namespace":         "test-namespace",
								"uid":               "test-daemonset-2",
								"resourceVersion":   "123",
							},
							"spec": map[string]interface{}{
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"selector": nil,
							},
							"status": map[string]interface{}{
								"currentNumberScheduled": float64(1),
								"desiredNumberScheduled": float64(1),
								"numberAvailable":        float64(1),
								"numberMisscheduled":     float64(1),
								"numberReady":            float64(1),
							},
						},
					},
				},
				{
					testCase: "Test DaemonSet 3 - Kind + Generate Name",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/test-daemonset-3",
						Type:       topology.Type{Name: "daemonset"},
						Data: topology.Data{
							"name":              "test-daemonset-3",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-daemonset",
								"namespace":      "test-namespace",
							},
							"uid":            types.UID("test-daemonset-3"),
							"updateStrategy": appsV1.RollingUpdateDaemonSetStrategyType,
							"kind":           "DaemonSet",
							"generateName":   "some-specified-generation",
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/test-daemonset-3",
						Type:       topology.Type{Name: "daemonset"},
						Data: topology.Data{
							"name": "test-daemonset-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-daemonset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "DaemonSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"generateName":      "some-specified-generation",
								"name":              "test-daemonset-3",
								"namespace":         "test-namespace",
								"uid":               "test-daemonset-3",
							},
							"spec": map[string]interface{}{
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"selector": nil,
							},
						},
					},
					expectedKubeStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/test-daemonset-3",
						Type:       topology.Type{Name: "daemonset"},
						Data: topology.Data{
							"name": "test-daemonset-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-daemonset",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "apps/v1",
							"kind":       "DaemonSet",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"generateName":      "some-specified-generation",
								"name":              "test-daemonset-3",
								"namespace":         "test-namespace",
								"uid":               "test-daemonset-3",
								"resourceVersion":   "123",
							},
							"spec": map[string]interface{}{
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": interface{}(nil),
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
								"updateStrategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"selector": nil,
							},
							"status": map[string]interface{}{
								"currentNumberScheduled": float64(1),
								"desiredNumberScheduled": float64(1),
								"numberAvailable":        float64(1),
								"numberMisscheduled":     float64(1),
								"numberReady":            float64(1),
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

type MockDaemonSetAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockDaemonSetAPICollectorClient) GetDaemonSets() ([]appsV1.DaemonSet, error) {
	daemonSets := make([]appsV1.DaemonSet, 0)
	for i := 1; i <= 3; i++ {
		daemonSet := appsV1.DaemonSet{
			TypeMeta: v1.TypeMeta{
				Kind: "DaemonSet",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-daemonset-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:             types.UID(fmt.Sprintf("test-daemonset-%d", i)),
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
			Spec: appsV1.DaemonSetSpec{
				UpdateStrategy: appsV1.DaemonSetUpdateStrategy{
					Type: appsV1.RollingUpdateDaemonSetStrategyType,
				},
			},
			Status: appsV1.DaemonSetStatus{
				CurrentNumberScheduled: int32(1),
				NumberMisscheduled:     int32(1),
				DesiredNumberScheduled: int32(1),
				NumberReady:            int32(1),
				NumberAvailable:        int32(1),
			},
		}

		if i == 3 {
			daemonSet.TypeMeta.Kind = "DaemonSet"
			daemonSet.ObjectMeta.GenerateName = "some-specified-generation"
		}

		daemonSets = append(daemonSets, daemonSet)
	}

	return daemonSets, nil
}
