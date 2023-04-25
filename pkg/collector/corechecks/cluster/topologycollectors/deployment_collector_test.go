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
	appsV1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var lastAppliedConfigurationDeployment = `{"apiVersion":"apps/v1","kind":"Deployment",
	"metadata":{"annotations":{},"name":"nginx-deployment","namespace":"default"},
	"spec":{"minReadySeconds":5,"selector":{"matchLabels":{"app":nginx}},"template":{"metadata":{"labels":{"app":"nginx"}},
	"spec":{"containers":[{"image":"nginx:1.14.2","name":"nginx",
	"ports":[{"containerPort":80}]}]}}}}`

func TestDeploymentCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	replicas = int32(1)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			commonClusterCollector := NewTestCommonClusterCollector(MockDeploymentAPICollectorClient{}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled)
			commonClusterCollector.SetUseRelationCache(false)
			cmc := NewDeploymentCollector(commonClusterCollector)
			expectedCollectorName := "Deployment Collector"
			RunCollectorTest(t, cmc, expectedCollectorName)

			for _, tc := range []struct {
				testCase             string
				expectedSP           *topology.Component
				expectedNoSP         *topology.Component
				expectedSPPlusStatus *topology.Component
			}{
				{
					testCase: "Test Deployment 1",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-1",
						Type:       topology.Type{Name: "deployment"},
						Data: topology.Data{
							"name":              "test-deployment-1",
							"kind":              "Deployment",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-deployment",
								"namespace":      "test-namespace",
							},
							"uid":                types.UID("test-deployment-1"),
							"deploymentStrategy": appsV1.RollingUpdateDeploymentStrategyType,
							"desiredReplicas":    &replicas,
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-1",
						Type:       topology.Type{Name: "deployment"},
						Data: topology.Data{
							"name": "test-deployment-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-deployment",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: topology.Data{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTime.UTC().Format(time.RFC3339),
								"labels": map[string]interface{}{
									"test": "label",
								},
								"name":      "test-deployment-1",
								"namespace": "test-namespace",
								"uid":       "test-deployment-1",
							},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"strategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-1",
						Type:       topology.Type{Name: "deployment"},
						Data: topology.Data{
							"name": "test-deployment-1",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-deployment",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: topology.Data{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTime.UTC().Format(time.RFC3339),
								"labels": map[string]interface{}{
									"test": "label",
								},
								"name":            "test-deployment-1",
								"namespace":       "test-namespace",
								"uid":             "test-deployment-1",
								"resourceVersion": "123",
								"annotations": map[string]interface{}{
									"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationDeployment,
								},
							},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"strategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
							"status": map[string]interface{}{
								"observedGeneration":  float64(321),
								"replicas":            float64(1),
								"updatedReplicas":     float64(1),
								"readyReplicas":       float64(1),
								"availableReplicas":   float64(1),
								"unavailableReplicas": float64(1),
								"conditions": []interface{}{map[string]interface{}{
									"type":               "Available",
									"status":             "True",
									"lastUpdateTime":     creationTime.UTC().Format(time.RFC3339),
									"lastTransitionTime": creationTime.UTC().Format(time.RFC3339),
									"reason":             "NewReplicaSetAvailable",
									"message":            "Deployment has minimum availability.",
								},
								},
							},
						},
					},
				},
				{
					testCase: "Test Deployment 2",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-2",
						Type:       topology.Type{Name: "deployment"},
						Data: topology.Data{
							"name":              "test-deployment-2",
							"kind":              "Deployment",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-deployment",
								"namespace":      "test-namespace",
							},
							"uid":                types.UID("test-deployment-2"),
							"deploymentStrategy": appsV1.RollingUpdateDeploymentStrategyType,
							"desiredReplicas":    &replicas,
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-2",
						Type:       topology.Type{Name: "deployment"},
						Data: topology.Data{
							"name": "test-deployment-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-deployment",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: topology.Data{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTime.UTC().Format(time.RFC3339),
								"labels": map[string]interface{}{
									"test": "label",
								},
								"name":      "test-deployment-2",
								"namespace": "test-namespace",
								"uid":       "test-deployment-2",
							},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"strategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-2",
						Type:       topology.Type{Name: "deployment"},
						Data: topology.Data{
							"name": "test-deployment-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-deployment",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: topology.Data{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTime.UTC().Format(time.RFC3339),
								"labels": map[string]interface{}{
									"test": "label",
								},
								"name":            "test-deployment-2",
								"namespace":       "test-namespace",
								"uid":             "test-deployment-2",
								"resourceVersion": "123",
								"annotations":     map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationDeployment},
							},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"strategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
							"status": map[string]interface{}{
								"observedGeneration":  float64(321),
								"replicas":            float64(1),
								"updatedReplicas":     float64(1),
								"readyReplicas":       float64(1),
								"availableReplicas":   float64(1),
								"unavailableReplicas": float64(1),
								"conditions": []interface{}{map[string]interface{}{
									"type":               "Available",
									"status":             "True",
									"lastUpdateTime":     creationTime.UTC().Format(time.RFC3339),
									"lastTransitionTime": creationTime.UTC().Format(time.RFC3339),
									"reason":             "NewReplicaSetAvailable",
									"message":            "Deployment has minimum availability.",
								},
								},
							},
						},
					},
				},
				{
					testCase: "Test Deployment 3 - Kind + Generate Name",
					expectedNoSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-3",
						Type:       topology.Type{Name: "deployment"},
						Data: topology.Data{
							"name":              "test-deployment-3",
							"kind":              "Deployment",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-deployment",
								"namespace":      "test-namespace",
							},
							"uid":                types.UID("test-deployment-3"),
							"generateName":       "some-specified-generation",
							"deploymentStrategy": appsV1.RollingUpdateDeploymentStrategyType,
							"desiredReplicas":    &replicas,
						},
					},
					expectedSP: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-3",
						Type:       topology.Type{Name: "deployment"},
						Data: topology.Data{
							"name": "test-deployment-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-deployment",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: topology.Data{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTime.UTC().Format(time.RFC3339),
								"labels": map[string]interface{}{
									"test": "label",
								},
								"annotations": map[string]interface{}{
									"another-annotation-1": "should-be-kept",
								},
								"name":         "test-deployment-3",
								"generateName": "some-specified-generation",
								"namespace":    "test-namespace",
								"uid":          "test-deployment-3",
							},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"strategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
						},
					},
					expectedSPPlusStatus: &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-3",
						Type:       topology.Type{Name: "deployment"},
						Data: topology.Data{
							"name": "test-deployment-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-deployment",
								"namespace":      "test-namespace",
							},
						},
						SourceProperties: topology.Data{
							"apiVersion": "apps/v1",
							"kind":       "Deployment",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTime.UTC().Format(time.RFC3339),
								"labels": map[string]interface{}{
									"test": "label",
								},
								"annotations": map[string]interface{}{
									"another-annotation-1":                             "should-be-kept",
									"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationDeployment,
								},
								"name":            "test-deployment-3",
								"generateName":    "some-specified-generation",
								"namespace":       "test-namespace",
								"uid":             "test-deployment-3",
								"resourceVersion": "123",
							},
							"spec": map[string]interface{}{
								"replicas": float64(1),
								"selector": nil,
								"strategy": map[string]interface{}{
									"type": "RollingUpdate",
								},
								"template": map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": nil,
									},
									"spec": map[string]interface{}{
										"containers": nil,
									},
								},
							},
							"status": map[string]interface{}{
								"observedGeneration":  float64(321),
								"replicas":            float64(1),
								"updatedReplicas":     float64(1),
								"readyReplicas":       float64(1),
								"availableReplicas":   float64(1),
								"unavailableReplicas": float64(1),
								"conditions": []interface{}{map[string]interface{}{
									"type":               "Available",
									"status":             "True",
									"lastUpdateTime":     creationTime.UTC().Format(time.RFC3339),
									"lastTransitionTime": creationTime.UTC().Format(time.RFC3339),
									"reason":             "NewReplicaSetAvailable",
									"message":            "Deployment has minimum availability.",
								},
								},
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

type MockDeploymentAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockDeploymentAPICollectorClient) GetDeployments() ([]appsV1.Deployment, error) {
	deployments := make([]appsV1.Deployment, 0)
	for i := 1; i <= 3; i++ {
		deployment := appsV1.Deployment{
			TypeMeta: v1.TypeMeta{
				Kind: "Deployment",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-deployment-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:             types.UID(fmt.Sprintf("test-deployment-%d", i)),
				GenerateName:    "",
				ResourceVersion: "123",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationDeployment,
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
			Spec: appsV1.DeploymentSpec{
				Strategy: appsV1.DeploymentStrategy{
					Type: appsV1.RollingUpdateDeploymentStrategyType,
				},
				Replicas: &replicas,
			},
			Status: appsV1.DeploymentStatus{
				ObservedGeneration:  int64(321),
				Replicas:            int32(1),
				UpdatedReplicas:     int32(1),
				ReadyReplicas:       int32(1),
				AvailableReplicas:   int32(1),
				UnavailableReplicas: int32(1),
				Conditions: []appsV1.DeploymentCondition{
					{
						Type:               "Available",
						Status:             "True",
						LastUpdateTime:     creationTime,
						LastTransitionTime: creationTime,
						Reason:             "NewReplicaSetAvailable",
						Message:            "Deployment has minimum availability.",
					},
				},
			},
		}

		if i == 3 {
			deployment.TypeMeta.Kind = "Deployment"
			deployment.ObjectMeta.GenerateName = "some-specified-generation"
			deployment.Annotations = map[string]string{
				"another-annotation-1":                             "should-be-kept",
				"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationDeployment,
			}
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}
