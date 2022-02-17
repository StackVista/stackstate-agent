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

var configMap coreV1.ConfigMapVolumeSource
var secret coreV1.SecretVolumeSource

func TestPodCollector(t *testing.T) {

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)
	pathType = coreV1.HostPathFileOrCreate
	gcePersistentDisk = coreV1.GCEPersistentDiskVolumeSource{
		PDName: "name-of-the-gce-persistent-disk",
	}
	awsElasticBlockStore = coreV1.AWSElasticBlockStoreVolumeSource{
		VolumeID: "id-of-the-aws-block-store",
	}
	configMap = coreV1.ConfigMapVolumeSource{
		LocalObjectReference: coreV1.LocalObjectReference{
			Name: "name-of-the-config-map",
		},
	}
	secret = coreV1.SecretVolumeSource{
		SecretName: "name-of-the-secret",
	}
	hostPath = coreV1.HostPathVolumeSource{
		Path: "some/path/to/the/volume",
		Type: &pathType,
	}

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		componentChannel := make(chan *topology.Component)
		relationChannel := make(chan *topology.Relation)
		containerCorrelationChannel := make(chan *ContainerCorrelation)
		volumeCorrelationChannel := make(chan *VolumeCorrelation)

		ic := NewPodCollector(componentChannel, relationChannel, containerCorrelationChannel, volumeCorrelationChannel, NewTestCommonClusterCollector(MockPodAPICollectorClient{}, sourcePropertiesEnabled))
		expectedCollectorName := "Pod Collector"
		RunCollectorTest(t, ic, expectedCollectorName)

		for _, tc := range []struct {
			testCase   string
			assertions []func()
		}{
			{
				testCase: "Test Pod 1 - Minimal",
				assertions: []func(){
					func() {
						component := <-componentChannel
						expectedComponent := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-1",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":              "test-pod-1",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"uid":               types.UID("test-pod-1"),
									"identifiers":       []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-1:10.0.0.1"},
									"restartPolicy":     coreV1.RestartPolicyAlways,
									"status": coreV1.PodStatus{
										Phase:     coreV1.PodRunning,
										StartTime: &creationTime,
										PodIP:     "10.0.0.1",
									}},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-1",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":        "test-pod-1",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"identifiers": []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-1:10.0.0.1"},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-pod-1",
										"namespace":         "test-namespace",
										"uid":               "test-pod-1"},
									"spec": map[string]interface{}{
										"nodeName":      "test-node",
										"restartPolicy": "Always"},
									"status": map[string]interface{}{
										"phase":     "Running",
										"podIP":     "10.0.0.1",
										"startTime": creationTimeFormatted,
									},
								},
							},
						)
						assert.EqualValues(t, expectedComponent, component)
					},
					expectPodNodeRelation(t, relationChannel, "test-pod-1"),
					expectNamespaceRelation(t, relationChannel, "test-pod-1"),
				},
			},
			{
				testCase: "Test Pod 2 - All Metadata",
				assertions: []func(){
					func() {
						component := <-componentChannel
						expectedComponent := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-2",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":              "test-pod-2",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace", "service-account": "some-service-account-name"},
									"uid":               types.UID("test-pod-2"),
									"identifiers":       []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-2:10.0.0.2"},
									"restartPolicy":     coreV1.RestartPolicyAlways,
									"kind":              "some-specified-kind",
									"generateName":      "some-specified-generation",
									"status": coreV1.PodStatus{
										Phase:             coreV1.PodRunning,
										StartTime:         &creationTime,
										PodIP:             "10.0.0.2",
										Message:           "some longer readable message for the phase",
										Reason:            "some-short-reason",
										NominatedNodeName: "some-nominated-node-name",
										QOSClass:          "some-qos-class",
									}},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-2",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":        "test-pod-2",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace", "service-account": "some-service-account-name"},
									"identifiers": []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-2:10.0.0.2"},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-pod-2",
										"namespace":         "test-namespace",
										"generateName":      "some-specified-generation",
										"uid":               "test-pod-2"},
									"spec": map[string]interface{}{
										"hostNetwork":        true,
										"nodeName":           "test-node",
										"serviceAccountName": "some-service-account-name",
										"restartPolicy":      "Always"},
									"status": map[string]interface{}{
										"phase":             "Running",
										"startTime":         creationTimeFormatted,
										"podIP":             "10.0.0.2",
										"message":           "some longer readable message for the phase",
										"reason":            "some-short-reason",
										"nominatedNodeName": "some-nominated-node-name",
										"qosClass":          "some-qos-class",
									},
								},
							},
						)
						assert.EqualValues(t, expectedComponent, component)
					},
					expectPodNodeRelation(t, relationChannel, "test-pod-2"),
					expectNamespaceRelation(t, relationChannel, "test-pod-2"),
				},
			},
			{
				testCase: "Test Pod 3 - All Controllers: Daemonset, Deployment, Job, ReplicaSet, StatefulSet",
				assertions: []func(){
					func() {
						component := <-componentChannel
						expectedComponent := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":              "test-pod-3",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"uid":               types.UID("test-pod-3"),
									"identifiers":       []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-3:10.0.0.1"},
									"restartPolicy":     coreV1.RestartPolicyAlways,
									"status": coreV1.PodStatus{
										Phase:     coreV1.PodRunning,
										StartTime: &creationTime,
										PodIP:     "10.0.0.1",
									}},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":        "test-pod-3",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"identifiers": []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-3:10.0.0.1"},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-pod-3",
										"namespace":         "test-namespace",
										"uid":               "test-pod-3",
										"ownerReferences": []interface{}{
											map[string]interface{}{"kind": "DaemonSet", "name": "daemonset-v"},
											map[string]interface{}{"kind": "Deployment", "name": "deployment-w"},
											map[string]interface{}{"kind": "Job", "name": "job-x"},
											map[string]interface{}{"kind": "ReplicaSet", "name": "replicaset-y"},
											map[string]interface{}{"kind": "StatefulSet", "name": "statefulset-z"},
										}},
									"spec": map[string]interface{}{
										"nodeName":      "test-node",
										"restartPolicy": "Always"},
									"status": map[string]interface{}{
										"phase":     "Running",
										"startTime": creationTimeFormatted,
										"podIP":     "10.0.0.1",
									},
								},
							},
						)
						assert.EqualValues(t, expectedComponent, component)
					},
					expectPodNodeRelation(t, relationChannel, "test-pod-3"),
					func() {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/daemonset-v->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Type:     topology.Type{Name: "controls"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:daemonset/daemonset-v",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
					func() {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/deployment-w->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Type:     topology.Type{Name: "controls"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:deployment/deployment-w",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
					func() {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:job/job-x->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Type:     topology.Type{Name: "controls"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:job/job-x",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
					func() {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/replicaset-y->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Type:     topology.Type{Name: "controls"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:replicaset/replicaset-y",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
					func() {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/statefulset-z->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Type:     topology.Type{Name: "controls"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:statefulset/statefulset-z",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-3",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
				},
			},
			{
				testCase: "Test Pod 4 - Volumes + Persistent Volumes + HostPath",
				assertions: []func(){
					func() {
						component := <-componentChannel
						expectedComponent := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-4",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":              "test-pod-4",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"uid":               types.UID("test-pod-4"),
									"identifiers":       []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-4:10.0.0.1"},
									"restartPolicy":     coreV1.RestartPolicyAlways,
									"status": coreV1.PodStatus{
										Phase:     coreV1.PodRunning,
										StartTime: &creationTime,
										PodIP:     "10.0.0.1",
									},
								},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-4",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":        "test-pod-4",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"identifiers": []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-4:10.0.0.1"},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-pod-4",
										"namespace":         "test-namespace",
										"uid":               "test-pod-4"},
									"spec": map[string]interface{}{
										"nodeName":      "test-node",
										"restartPolicy": "Always",
										"volumes": []interface{}{
											map[string]interface{}{
												"name": "test-volume-1",
												"volumeSource": map[string]interface{}{
													"awsElasticBlockStore": map[string]interface{}{
														"volumeID": "id-of-the-aws-block-store"}}},
											map[string]interface{}{
												"name": "test-volume-2",
												"volumeSource": map[string]interface{}{
													"gcePersistentDisk": map[string]interface{}{
														"pdName": "name-of-the-gce-persistent-disk"}}},
											map[string]interface{}{
												"name": "test-volume-3",
												"volumeSource": map[string]interface{}{
													"configMap": map[string]interface{}{
														"localObjectReference": map[string]interface{}{
															"name": "name-of-the-config-map"}}}},
											map[string]interface{}{
												"name": "test-volume-4",
												"volumeSource": map[string]interface{}{
													"hostPath": map[string]interface{}{
														"path": "some/path/to/the/volume",
														"type": "FileOrCreate"}}},
											map[string]interface{}{
												"name": "test-volume-5",
												"volumeSource": map[string]interface{}{
													"secret": map[string]interface{}{
														"secretName": "name-of-the-secret"}}},
										}},
									"status": map[string]interface{}{
										"phase":     "Running",
										"startTime": creationTimeFormatted,
										"podIP":     "10.0.0.1",
									},
								},
							},
						)
						assert.EqualValues(t, expectedComponent, component)
					},
					expectPodNodeRelation(t, relationChannel, "test-pod-4"),
					expectNamespaceRelation(t, relationChannel, "test-pod-4"),
					func() {
						correlation := <-volumeCorrelationChannel
						assert.Len(t, correlation.Volumes, 5)
						assert.Equal(t, correlation.Pod.ExternalID, "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-4")
						assert.Equal(t, correlation.Pod.Name, "test-pod-4")
						assert.Equal(t, correlation.Pod.Namespace, "test-namespace")
					},
				},
			},
			{
				testCase: "Test Pod 5 - Containers + Config Maps",
				assertions: []func(){
					func() {
						component := <-componentChannel
						expectedComponent := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-5",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":              "test-pod-5",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"uid":               types.UID("test-pod-5"),
									"identifiers":       []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-5:10.0.0.1"},
									"restartPolicy":     coreV1.RestartPolicyAlways,
									"status": coreV1.PodStatus{
										Phase:     coreV1.PodRunning,
										StartTime: &creationTime,
										PodIP:     "10.0.0.1",
									},
								},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-5",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":        "test-pod-5",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"identifiers": []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-5:10.0.0.1"},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-pod-5",
										"namespace":         "test-namespace",
										"uid":               "test-pod-5"},
									"spec": map[string]interface{}{
										"nodeName":      "test-node",
										"restartPolicy": "Always",
										"containers": []interface{}{
											map[string]interface{}{
												"env": []interface{}{
													map[string]interface{}{
														"name": "env-var",
														"valueFrom": map[string]interface{}{
															"configMapKeyRef": map[string]interface{}{
																"localObjectReference": map[string]interface{}{
																	"name": "name-of-the-env-config-map"}}}}},
												"envFrom": []interface{}{
													map[string]interface{}{
														"configMapRef": map[string]interface{}{
															"localObjectReference": map[string]interface{}{
																"name": "name-of-the-config-map"}}}},
												"image":     "docker/image/repo/container:latest",
												"name":      "container-1",
												"resources": map[string]interface{}{}}},
									},
									"status": map[string]interface{}{
										"phase":     "Running",
										"startTime": creationTimeFormatted,
										"podIP":     "10.0.0.1",
									},
								},
							},
						)
						assert.EqualValues(t, expectedComponent, component)
					},
					expectPodNodeRelation(t, relationChannel, "test-pod-5"),
					expectNamespaceRelation(t, relationChannel, "test-pod-5"),
					func() {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-5->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:configmap/name-of-the-config-map",
							Type:     topology.Type{Name: "uses"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-5",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/name-of-the-config-map",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
					func() {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-5->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:configmap/name-of-the-env-config-map",
							Type:     topology.Type{Name: "uses_value"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-5",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:configmap/name-of-the-env-config-map",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
				},
			},
			{
				testCase: "Test Pod 6 - Containers + Config Maps",
				assertions: []func(){
					func() {
						component := <-componentChannel
						expectedComponent := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-6",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":              "test-pod-6",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"uid":               types.UID("test-pod-6"),
									"identifiers":       []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-6:10.0.0.1"},
									"restartPolicy":     coreV1.RestartPolicyAlways,
									"status": coreV1.PodStatus{
										Phase:     coreV1.PodRunning,
										StartTime: &creationTime,
										PodIP:     "10.0.0.1",
									},
								},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-6",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":        "test-pod-6",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"identifiers": []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-6:10.0.0.1"},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-pod-6",
										"namespace":         "test-namespace",
										"uid":               "test-pod-6"},
									"spec": map[string]interface{}{
										"nodeName":      "test-node",
										"restartPolicy": "Always",
										"containers": []interface{}{
											map[string]interface{}{
												"env": []interface{}{
													map[string]interface{}{
														"name": "env-var",
														"valueFrom": map[string]interface{}{
															"secretKeyRef": map[string]interface{}{
																"localObjectReference": map[string]interface{}{
																	"name": "name-of-the-env-secret"}}}}},
												"envFrom": []interface{}{
													map[string]interface{}{
														"secretRef": map[string]interface{}{
															"localObjectReference": map[string]interface{}{
																"name": "name-of-the-secret"}}}},
												"image":     "docker/image/repo/container:latest",
												"name":      "container-1",
												"resources": map[string]interface{}{}}},
									},
									"status": map[string]interface{}{
										"phase":     "Running",
										"startTime": creationTimeFormatted,
										"podIP":     "10.0.0.1",
									},
								},
							},
						)
						assert.EqualValues(t, expectedComponent, component)
					},
					expectPodNodeRelation(t, relationChannel, "test-pod-6"),
					expectNamespaceRelation(t, relationChannel, "test-pod-6"),
					func() {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-6->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:secret/name-of-the-secret",
							Type:     topology.Type{Name: "uses"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-6",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/name-of-the-secret",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
					func() {
						relation := <-relationChannel
						expectedRelation := &topology.Relation{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-6->" +
								"urn:kubernetes:/test-cluster-name:test-namespace:secret/name-of-the-env-secret",
							Type:     topology.Type{Name: "uses_value"},
							SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-6",
							TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:secret/name-of-the-env-secret",
							Data:     map[string]interface{}{},
						}
						assert.EqualValues(t, expectedRelation, relation)
					},
				},
			},
			{
				testCase: "Test Pod 7 - Containers + Container Correlation",
				assertions: []func(){
					func() {
						component := <-componentChannel
						expectedComponent := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-7",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":              "test-pod-7",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"uid":               types.UID("test-pod-7"),
									"identifiers":       []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-7:10.0.0.1"},
									"restartPolicy":     coreV1.RestartPolicyAlways,
									"status": coreV1.PodStatus{
										Phase:     coreV1.PodRunning,
										StartTime: &creationTime,
										PodIP:     "10.0.0.1",
									},
								},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-7",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":        "test-pod-7",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"identifiers": []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-7:10.0.0.1"},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-pod-7",
										"namespace":         "test-namespace",
										"uid":               "test-pod-7",
									},
									"spec": map[string]interface{}{
										"nodeName":      "test-node",
										"restartPolicy": "Always"},
									"status": map[string]interface{}{
										"phase":     "Running",
										"startTime": creationTimeFormatted,
										"podIP":     "10.0.0.1",
									},
								},
							},
						)
						assert.EqualValues(t, expectedComponent, component)
					},
					expectPodNodeRelation(t, relationChannel, "test-pod-7"),
					expectNamespaceRelation(t, relationChannel, "test-pod-7"),
					func() {
						correlation := <-containerCorrelationChannel
						expectedCorrelation := &ContainerCorrelation{
							Pod: ContainerPod{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-7",
								Name:       "test-pod-7",
								Labels:     map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
								PodIP:      "10.0.0.1",
								Namespace:  "test-namespace",
								NodeName:   "test-node",
								Phase:      "Running",
							},
							ContainerStatuses: []coreV1.ContainerStatus{
								{
									Name:  "container-1",
									Image: "docker/image/repo/container-1:latest",
								},
								{
									Name:  "container-2",
									Image: "docker/image/repo/container-2:latest",
								},
							},
						}
						assert.EqualValues(t, expectedCorrelation, correlation)
					},
				},
			},
			{
				testCase: "Test Pod 8 - Pod Phase Succeeded - no Job relation created",
				assertions: []func(){
					func() {
						component := <-componentChannel
						expectedComponent := chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-8",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":              "test-pod-8",
									"creationTimestamp": creationTime,
									"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"uid":               types.UID("test-pod-8"),
									"identifiers":       []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-8:10.0.0.1"},
									"restartPolicy":     coreV1.RestartPolicyAlways,
									"status": coreV1.PodStatus{
										Phase:     coreV1.PodSucceeded,
										StartTime: &creationTime,
										PodIP:     "10.0.0.1",
									},
								},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:pod/test-pod-8",
								Type:       topology.Type{Name: "pod"},
								Data: topology.Data{
									"name":        "test-pod-8",
									"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
									"identifiers": []string{"urn:ip:/test-cluster-name:test-namespace:test-pod-8:10.0.0.1"},
								},
								SourceProperties: map[string]interface{}{
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-pod-8",
										"namespace":         "test-namespace",
										"ownerReferences":   []interface{}{map[string]interface{}{"kind": "Job", "name": "test-job-8"}},
										"uid":               "test-pod-8"},
									"spec": map[string]interface{}{
										"nodeName":      "test-node",
										"restartPolicy": "Always",
									},
									"status": map[string]interface{}{
										"phase":     "Succeeded",
										"startTime": creationTimeFormatted,
										"podIP":     "10.0.0.1",
									},
								},
							},
						)
						assert.EqualValues(t, expectedComponent, component)
					},
					expectPodNodeRelation(t, relationChannel, "test-pod-8"),
					expectNamespaceRelation(t, relationChannel, "test-pod-8"),
					func() {
						// there should be no relations created for skipped pod
						assert.Empty(t, relationChannel)
					},
				},
			},
		} {
			t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled), func(t *testing.T) {
				for _, assertion := range tc.assertions {
					assertion()
				}
			})
		}
	}
}

type MockPodAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockPodAPICollectorClient) GetPods() ([]coreV1.Pod, error) {
	pods := make([]coreV1.Pod, 0)
	for i := 1; i <= 8; i++ {
		pod := coreV1.Pod{
			TypeMeta: v1.TypeMeta{
				Kind: "",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-pod-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:          types.UID(fmt.Sprintf("test-pod-%d", i)),
				GenerateName: "",
			},
			Status: coreV1.PodStatus{
				Phase:     coreV1.PodRunning,
				PodIP:     "10.0.0.1",
				StartTime: &creationTime,
			},
			Spec: coreV1.PodSpec{
				RestartPolicy: coreV1.RestartPolicyAlways,
				NodeName:      "test-node",
			},
		}

		if i == 2 {
			pod.Spec.HostNetwork = true
			pod.Status.PodIP = "10.0.0.2"
			pod.Spec.ServiceAccountName = "some-service-account-name"
			pod.Status.Message = "some longer readable message for the phase"
			pod.Status.Reason = "some-short-reason"
			pod.Status.NominatedNodeName = "some-nominated-node-name"
			pod.Status.QOSClass = "some-qos-class"
			pod.TypeMeta.Kind = "some-specified-kind"
			pod.ObjectMeta.GenerateName = "some-specified-generation"
		}

		if i == 3 {
			pod.OwnerReferences = []v1.OwnerReference{
				{Kind: "DaemonSet", Name: "daemonset-v"},
				{Kind: "Deployment", Name: "deployment-w"},
				{Kind: "Job", Name: "job-x"},
				{Kind: "ReplicaSet", Name: "replicaset-y"},
				{Kind: "StatefulSet", Name: "statefulset-z"},
			}
		}

		if i == 4 {
			pod.Spec.Volumes = []coreV1.Volume{
				{Name: "test-volume-1", VolumeSource: coreV1.VolumeSource{AWSElasticBlockStore: &awsElasticBlockStore}},
				{Name: "test-volume-2", VolumeSource: coreV1.VolumeSource{GCEPersistentDisk: &gcePersistentDisk}},
				{Name: "test-volume-3", VolumeSource: coreV1.VolumeSource{ConfigMap: &configMap}},
				{Name: "test-volume-4", VolumeSource: coreV1.VolumeSource{HostPath: &hostPath}},
				{Name: "test-volume-5", VolumeSource: coreV1.VolumeSource{Secret: &secret}},
			}
		}

		if i == 5 {
			pod.Spec.Containers = []coreV1.Container{
				{
					Name:  "container-1",
					Image: "docker/image/repo/container:latest",
					Env: []coreV1.EnvVar{
						{
							Name: "env-var",
							ValueFrom: &coreV1.EnvVarSource{
								ConfigMapKeyRef: &coreV1.ConfigMapKeySelector{
									LocalObjectReference: coreV1.LocalObjectReference{Name: "name-of-the-env-config-map"},
								},
							},
						},
					},
					EnvFrom: []coreV1.EnvFromSource{
						{
							ConfigMapRef: &coreV1.ConfigMapEnvSource{
								LocalObjectReference: coreV1.LocalObjectReference{Name: "name-of-the-config-map"},
							},
						},
					},
				},
			}
		}
		if i == 6 {
			pod.Spec.Containers = []coreV1.Container{
				{
					Name:  "container-1",
					Image: "docker/image/repo/container:latest",
					Env: []coreV1.EnvVar{
						{
							Name: "env-var",
							ValueFrom: &coreV1.EnvVarSource{
								SecretKeyRef: &coreV1.SecretKeySelector{
									LocalObjectReference: coreV1.LocalObjectReference{Name: "name-of-the-env-secret"},
								},
							},
						},
					},
					EnvFrom: []coreV1.EnvFromSource{
						{
							SecretRef: &coreV1.SecretEnvSource{
								LocalObjectReference: coreV1.LocalObjectReference{Name: "name-of-the-secret"},
							},
						},
					},
				},
			}
		}

		if i == 7 {
			pod.Status.ContainerStatuses = []coreV1.ContainerStatus{
				{
					Name:  "container-1",
					Image: "docker/image/repo/container-1:latest",
				},
				{
					Name:  "container-2",
					Image: "docker/image/repo/container-2:latest",
				},
			}
		}

		if i == 8 {
			pod.Status.Phase = coreV1.PodSucceeded
			pod.OwnerReferences = []v1.OwnerReference{
				{Kind: "Job", Name: "test-job-8"},
			}
		}

		pods = append(pods, pod)
	}

	return pods, nil
}

func expectNamespaceRelation(t *testing.T, ch chan *topology.Relation, podName string) func() {
	return func() {
		relation := <-ch
		expected := &topology.Relation{
			ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
				fmt.Sprintf("urn:kubernetes:/test-cluster-name:test-namespace:pod/%s", podName),
			Type:     topology.Type{Name: "encloses"},
			SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
			TargetID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:test-namespace:pod/%s", podName),
			Data:     map[string]interface{}{},
		}
		assert.EqualValues(t, expected, relation)
	}
}

func expectPodNodeRelation(t *testing.T, ch chan *topology.Relation, podName string) func() {
	return func() {
		relation := <-ch
		expectedRelation := &topology.Relation{
			ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:test-namespace:pod/%s->", podName) +
				"urn:kubernetes:/test-cluster-name:node/test-node",
			Type:     topology.Type{Name: "scheduled_on"},
			SourceID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:test-namespace:pod/%s", podName),
			TargetID: "urn:kubernetes:/test-cluster-name:node/test-node",
			Data:     map[string]interface{}{},
		}
		assert.EqualValues(t, expectedRelation, relation)
	}

}
