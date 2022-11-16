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

func TestReplicaSetCollector(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)
	componentIDChannel := make(chan string)
	defer close(componentIDChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)
	replicas = 1

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		ic := NewReplicaSetCollector(relationChannel, NewTestCommonClusterCollector(MockReplicaSetAPICollectorClient{}, componentChannel, componentIDChannel, sourcePropertiesEnabled))
		expectedCollectorName := "ReplicaSet Collector"
		RunCollectorTest(t, ic, expectedCollectorName)

		for _, tc := range []struct {
			testCase              string
			expectedComponentSP   *topology.Component
			expectedComponentNoSP *topology.Component
			expectedRelations     []*topology.Relation
		}{
			{
				testCase: "Test ReplicaSet 1 - Minimal",
				expectedComponentNoSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-1",
					Type:       topology.Type{Name: "replicaset"},
					Data: topology.Data{
						"name":              "test-replicaset-1",
						"creationTimestamp": creationTime,
						"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
						"uid":               types.UID("test-replicaset-1"),
						"desiredReplicas":   &replicas,
					},
				},
				expectedComponentSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-1",
					Type:       topology.Type{Name: "replicaset"},
					Data: topology.Data{
						"name": "test-replicaset-1",
						"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					},
					SourceProperties: map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
							"labels":            map[string]interface{}{"test": "label"},
							"name":              "test-replicaset-1",
							"namespace":         "test-namespace",
							"uid":               "test-replicaset-1"},
						"spec": map[string]interface{}{
							"replicas": float64(1),
							"template": map[string]interface{}{
								"metadata": map[string]interface{}{
									"creationTimestamp": nil},
								"spec": map[string]interface{}{},
							}},
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
						"creationTimestamp": creationTime,
						"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
						"uid":               types.UID("test-replicaset-2"),
						"desiredReplicas":   &replicas,
						"kind":              "some-specified-kind",
						"generateName":      "some-specified-generation",
					},
				},
				expectedComponentSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-2",
					Type:       topology.Type{Name: "replicaset"},
					Data: topology.Data{
						"name": "test-replicaset-2",
						"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					},
					SourceProperties: map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
							"labels":            map[string]interface{}{"test": "label"},
							"name":              "test-replicaset-2",
							"namespace":         "test-namespace",
							"generateName":      "some-specified-generation",
							"uid":               "test-replicaset-2"},
						"spec": map[string]interface{}{
							"replicas": float64(1),
							"template": map[string]interface{}{
								"metadata": map[string]interface{}{
									"creationTimestamp": nil},
								"spec": map[string]interface{}{},
							}},
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
						"creationTimestamp": creationTime,
						"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
						"uid":               types.UID("test-replicaset-3"),
						"desiredReplicas":   &replicas,
						"kind":              "some-specified-kind",
						"generateName":      "some-specified-generation",
					},
				},
				expectedComponentSP: &topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/test-replicaset-3",
					Type:       topology.Type{Name: "replicaset"},
					Data: topology.Data{
						"name": "test-replicaset-3",
						"tags": map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
					},
					SourceProperties: map[string]interface{}{
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
							"labels":            map[string]interface{}{"test": "label"},
							"name":              "test-replicaset-3",
							"namespace":         "test-namespace",
							"generateName":      "some-specified-generation",

							"ownerReferences": []interface{}{map[string]interface{}{"kind": "Deployment", "name": "test-deployment-3"}},
							"uid":             "test-replicaset-3"},
						"spec": map[string]interface{}{
							"replicas": float64(1),
							"template": map[string]interface{}{
								"metadata": map[string]interface{}{
									"creationTimestamp": nil},
								"spec": map[string]interface{}{},
							}},
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
			t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled), func(t *testing.T) {
				service := <-componentChannel
				<-componentIDChannel
				if sourcePropertiesEnabled {
					assert.EqualValues(t, tc.expectedComponentSP, service)
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

type MockReplicaSetAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockReplicaSetAPICollectorClient) GetReplicaSets() ([]appsV1.ReplicaSet, error) {
	replicaSets := make([]appsV1.ReplicaSet, 0)
	for i := 1; i <= 3; i++ {
		replicaSet := appsV1.ReplicaSet{
			TypeMeta: v1.TypeMeta{
				Kind: "",
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
		}

		if i > 1 {
			replicaSet.TypeMeta.Kind = "some-specified-kind"
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
