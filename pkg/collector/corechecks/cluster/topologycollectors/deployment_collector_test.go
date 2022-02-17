// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
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

func TestDeploymentCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	replicas = int32(1)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		cmc := NewDeploymentCollector(componentChannel, relationChannel, NewTestCommonClusterCollector(MockDeploymentAPICollectorClient{}, sourcePropertiesEnabled))
		expectedCollectorName := "Deployment Collector"
		RunCollectorTest(t, cmc, expectedCollectorName)

		for _, tc := range []struct {
			testCase     string
			expectedSP   *topology.Component
			expectedNoSP *topology.Component
		}{
			{
				testCase: "Test Deployment 1",
				expectedNoSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/test-deployment-1",
					Type:       topology.Type{Name: "deployment"},
					Data: topology.Data{
						"name":               "test-deployment-1",
						"creationTimestamp":  creationTime,
						"tags":               map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
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
						"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					},
					SourceProperties: topology.Data{
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
							"strategy": map[string]interface{}{
								"type": "RollingUpdate",
							},
							"template": map[string]interface{}{
								"metadata": map[string]interface{}{
									"creationTimestamp": nil,
								},
								"spec": map[string]interface{}{},
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
						"name":               "test-deployment-2",
						"creationTimestamp":  creationTime,
						"tags":               map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
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
						"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					},
					SourceProperties: topology.Data{
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
							"strategy": map[string]interface{}{
								"type": "RollingUpdate",
							},
							"template": map[string]interface{}{
								"metadata": map[string]interface{}{
									"creationTimestamp": nil,
								},
								"spec": map[string]interface{}{},
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
						"name":               "test-deployment-3",
						"creationTimestamp":  creationTime,
						"tags":               map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
						"uid":                types.UID("test-deployment-3"),
						"kind":               "some-specified-kind",
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
						"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					},
					SourceProperties: topology.Data{
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
							"strategy": map[string]interface{}{
								"type": "RollingUpdate",
							},
							"template": map[string]interface{}{
								"metadata": map[string]interface{}{
									"creationTimestamp": nil,
								},
								"spec": map[string]interface{}{},
							},
						},
					},
				},
			},
		} {
			t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled), func(t *testing.T) {
				component := <-componentChannel

				if sourcePropertiesEnabled {
					assert.EqualValues(t, tc.expectedSP, component)
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

type MockDeploymentAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockDeploymentAPICollectorClient) GetDeployments() ([]appsV1.Deployment, error) {
	deployments := make([]appsV1.Deployment, 0)
	for i := 1; i <= 3; i++ {
		deployment := appsV1.Deployment{
			TypeMeta: v1.TypeMeta{
				Kind: "",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-deployment-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:          types.UID(fmt.Sprintf("test-deployment-%d", i)),
				GenerateName: "",
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
		}

		if i == 3 {
			deployment.TypeMeta.Kind = "some-specified-kind"
			deployment.ObjectMeta.GenerateName = "some-specified-generation"
			deployment.Annotations = map[string]string{
				"another-annotation-1": "should-be-kept",
				"kubectl.kubernetes.io/last-applied-configuration": `{"apiVersion":"apps/v1","kind":"Deployment",
				  "metadata":{"annotations":{},"name":"nginx-deployment","namespace":"default"},
				  "spec":{"minReadySeconds":5,"selector":{"matchLabels":{"app":nginx}},"template":{"metadata":{"labels":{"app":"nginx"}},
				  "spec":{"containers":[{"image":"nginx:1.14.2","name":"nginx",
				  "ports":[{"containerPort":80}]}]}}}}`,
			}
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}
