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
	"k8s.io/apimachinery/pkg/util/intstr"
)

var lastAppliedConfigurationService = `{"apiVersion":"v1","kind":"Service","metadata":{"annotations":{"argocd.io/tracking-id":"tenant"},"labels":{"app.kubernetes.io/component":"api","app.kubernetes.io/instance":"test","app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/name":"app","app.kubernetes.io/version":"1.0.0","helm.sh/chart":"1.0.0"},"name":"api","namespace":"tenant"},"spec":{"clusterIP":"None","ports":[{"name":"specs","port":8080,"protocol":"TCP","targetPort":"specs"}],"selector":{"app.kubernetes.io/component":"api","app.kubernetes.io/instance":"test"},"type":"ClusterIP"}}`

func TestServiceCollector(t *testing.T) {

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			for testCaseNo, tc := range serviceCollectorTestCases(sourcePropertiesEnabled, kubernetesStatusEnabled, creationTimeFormatted) {
				t.Run(serviceCollectorTestCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
					svcCorrelationChannel := make(chan *ServiceSelectorCorrelation)
					componentChannel := make(chan *topology.Component)
					relationChannel := make(chan *topology.Relation)
					collectorChannel := make(chan bool)

					commonCollector := NewTestCommonClusterCollector(MockServiceAPICollectorClient{testCaseNumber: testCaseNo + 1}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled)
					commonCollector.SetUseRelationCache(false)
					serviceCollector := NewServiceCollector(
						svcCorrelationChannel,
						commonCollector,
					)
					// Mock out DNS resolution function for test
					serviceCollector.(*ServiceCollector).DNS = func(name string) ([]string, error) {
						return []string{"10.10.42.42", "10.10.42.43"}, nil
					}

					assert.Equal(t, "Service Collector", serviceCollector.GetName())

					go func() {
						assert.NoError(t, serviceCollector.CollectorFunction())
						collectorChannel <- true
					}()

					var actualComponents []*topology.Component
					var actualRelations []*topology.Relation
				L:
					for {
						select {
						case component := <-componentChannel:
							actualComponents = append(actualComponents, component)
						case relation := <-relationChannel:
							actualRelations = append(actualRelations, relation)
						case <-svcCorrelationChannel: // ignore
						case <-collectorChannel:
							break L
						}
					}

					assert.EqualValues(t, tc.expectedComponents, actualComponents)
					assert.EqualValues(t, tc.expectedRelations, actualRelations)
				})
			}
		}
	}
}

func serviceCollectorTestCaseName(baseName string, sourcePropertiesEnabled bool, kubernetesStatusEnabled bool) string {
	if sourcePropertiesEnabled {
		baseName = baseName + " w/ sourceProps"
	} else {
		baseName = baseName + " w/o sourceProps"
	}
	if kubernetesStatusEnabled {
		baseName = baseName + " w/ kubeStatus"
	} else {
		baseName = baseName + " w/o kubeStatus"
	}
	return baseName
}

func serviceCollectorTestCases(sourcePropertiesEnabled bool, kubernetesStatusEnabled bool, creationTimeFormatted string) []serviceCollectorTestCase {
	testCase1 := serviceCollectorTestCase{
		testCase: "Test Service 1 - Service + Pod Relation",
		expectedComponents: []*topology.Component{
			chooseBySourcePropertiesFeature(
				sourcePropertiesEnabled,
				kubernetesStatusEnabled,
				&topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-1",
					Type:       topology.Type{Name: "service"},
					Data: topology.Data{
						"name":              "test-service-1",
						"kind":              "Service",
						"creationTimestamp": creationTime,
						"tags": map[string]string{
							"test":           "label",
							"cluster-name":   "test-cluster-name",
							"cluster-type":   "kubernetes",
							"component-type": "kubernetes-service",
							"namespace":      "test-namespace",
							"service-type":   "ClusterIP",
						},
						"uid":         types.UID("test-service-1"),
						"identifiers": []string{"urn:service:/test-cluster-name:test-namespace:test-service-1"},
					},
				},
				&topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-1",
					Type:       topology.Type{Name: "service"},
					Data: topology.Data{
						"name": "test-service-1",
						"tags": map[string]string{
							"test":           "label",
							"cluster-name":   "test-cluster-name",
							"cluster-type":   "kubernetes",
							"component-type": "kubernetes-service",
							"namespace":      "test-namespace",
							"service-type":   "ClusterIP",
						},
						"identifiers": []string{"urn:service:/test-cluster-name:test-namespace:test-service-1"},
					},
					SourceProperties: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
							"labels":            map[string]interface{}{"test": "label"},
							"name":              "test-service-1",
							"namespace":         "test-namespace",
							"uid":               "test-service-1"},
						"spec": map[string]interface{}{
							"type": "ClusterIP",
							"ports": []interface{}{
								map[string]interface{}{
									"name":       "test-service-port-1",
									"port":       float64(81),
									"targetPort": float64(8081)},
							}},
					},
				},
				&topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-1",
					Type:       topology.Type{Name: "service"},
					Data: topology.Data{
						"name": "test-service-1",
						"tags": map[string]string{
							"test":           "label",
							"cluster-name":   "test-cluster-name",
							"cluster-type":   "kubernetes",
							"component-type": "kubernetes-service",
							"namespace":      "test-namespace",
							"service-type":   "ClusterIP",
						},
						"identifiers": []string{"urn:service:/test-cluster-name:test-namespace:test-service-1"},
					},
					SourceProperties: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
							"labels":            map[string]interface{}{"test": "label"},
							"name":              "test-service-1",
							"namespace":         "test-namespace",
							"uid":               "test-service-1",
							"resourceVersion":   "123",
							"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationService},
						},
						"spec": map[string]interface{}{
							"type": "ClusterIP",
							"ports": []interface{}{
								map[string]interface{}{
									"name":       "test-service-port-1",
									"port":       float64(81),
									"targetPort": float64(8081)},
							},
						},
						"status": map[string]interface{}{
							"loadBalancer": map[string]interface{}{},
						},
					},
				},
			),
		},
		expectedRelations: []*topology.Relation{
			{
				ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
					"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-1",
				Type:     topology.Type{Name: "encloses"},
				SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
				TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-1",
				Data:     map[string]interface{}{},
			},
		},
	}

	testCase6 := serviceCollectorTestCase{
		testCase: "Test Service 6 - LoadBalancer + Ingress Points + Ingress Correlation",
		expectedComponents: []*topology.Component{
			chooseBySourcePropertiesFeature(
				sourcePropertiesEnabled,
				kubernetesStatusEnabled,
				&topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-6",
					Type:       topology.Type{Name: "service"},
					Data: topology.Data{
						"name":              "test-service-6",
						"kind":              "Service",
						"creationTimestamp": creationTime,
						"tags": map[string]string{
							"test":           "label",
							"cluster-name":   "test-cluster-name",
							"cluster-type":   "kubernetes",
							"component-type": "kubernetes-service",
							"namespace":      "test-namespace",
							"service-type":   "LoadBalancer",
						},
						"uid": types.UID("test-service-6"),
						"identifiers": []string{
							"urn:endpoint:/test-cluster-name:10.100.200.23", "urn:ingress-point:/34.100.200.15",
							"urn:ingress-point:/64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
							"urn:service:/test-cluster-name:test-namespace:test-service-6"},
					},
				},
				&topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-6",
					Type:       topology.Type{Name: "service"},
					Data: topology.Data{
						"name": "test-service-6",
						"tags": map[string]string{
							"test":           "label",
							"cluster-name":   "test-cluster-name",
							"cluster-type":   "kubernetes",
							"component-type": "kubernetes-service",
							"namespace":      "test-namespace",
							"service-type":   "LoadBalancer",
						},
						"identifiers": []string{
							"urn:endpoint:/test-cluster-name:10.100.200.23", "urn:ingress-point:/34.100.200.15",
							"urn:ingress-point:/64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
							"urn:service:/test-cluster-name:test-namespace:test-service-6"},
					},
					SourceProperties: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
							"labels":            map[string]interface{}{"test": "label"},
							"name":              "test-service-6",
							"namespace":         "test-namespace",
							"uid":               "test-service-6"},
						"spec": map[string]interface{}{
							"type":           "LoadBalancer",
							"loadBalancerIP": "10.100.200.23",
							"ports": []interface{}{
								map[string]interface{}{
									"name":       "test-service-port-6",
									"port":       float64(86),
									"targetPort": float64(8086)},
								map[string]interface{}{
									"name":       "test-service-node-port-6",
									"nodePort":   float64(10206),
									"port":       float64(86),
									"targetPort": float64(8086)},
							}},
					},
				},
				&topology.Component{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-6",
					Type:       topology.Type{Name: "service"},
					Data: topology.Data{
						"name": "test-service-6",
						"tags": map[string]string{
							"test":           "label",
							"cluster-name":   "test-cluster-name",
							"cluster-type":   "kubernetes",
							"component-type": "kubernetes-service",
							"namespace":      "test-namespace",
							"service-type":   "LoadBalancer",
						},
						"identifiers": []string{
							"urn:endpoint:/test-cluster-name:10.100.200.23", "urn:ingress-point:/34.100.200.15",
							"urn:ingress-point:/64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
							"urn:service:/test-cluster-name:test-namespace:test-service-6"},
					},
					SourceProperties: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": map[string]interface{}{
							"creationTimestamp": creationTimeFormatted,
							"labels":            map[string]interface{}{"test": "label"},
							"name":              "test-service-6",
							"namespace":         "test-namespace",
							"uid":               "test-service-6",
							"resourceVersion":   "123",
							"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationService},
						},
						"spec": map[string]interface{}{
							"type":           "LoadBalancer",
							"loadBalancerIP": "10.100.200.23",
							"ports": []interface{}{
								map[string]interface{}{
									"name":       "test-service-port-6",
									"port":       float64(86),
									"targetPort": float64(8086)},
								map[string]interface{}{
									"name":       "test-service-node-port-6",
									"nodePort":   float64(10206),
									"port":       float64(86),
									"targetPort": float64(8086)},
							}},
						"status": map[string]interface{}{
							"loadBalancer": map[string]interface{}{
								"ingress": []interface{}{
									map[string]interface{}{
										"ip": "34.100.200.15",
									},
									map[string]interface{}{
										"hostname": "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
									},
								},
							},
						},
					},
				},
			),
		},
		expectedRelations: []*topology.Relation{
			{
				ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
					"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-6",
				Type:     topology.Type{Name: "encloses"},
				SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
				TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-6",
				Data:     map[string]interface{}{},
			},
		},
	}

	return []serviceCollectorTestCase{
		testCase1,
		{
			testCase: "Test Service 2 - Minimal - NodePort",
			expectedComponents: []*topology.Component{
				chooseBySourcePropertiesFeature(
					sourcePropertiesEnabled,
					kubernetesStatusEnabled,
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-2",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name":              "test-service-2",
							"kind":              "Service",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "NodePort",
							},
							"uid": types.UID("test-service-2"),
							"identifiers": []string{
								"urn:endpoint:/test-cluster-name:10.100.200.20",
								"urn:endpoint:/test-cluster-name:10.100.200.20:10202",
								"urn:service:/test-cluster-name:test-namespace:test-service-2",
							},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-2",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "NodePort",
							},
							"identifiers": []string{
								"urn:endpoint:/test-cluster-name:10.100.200.20",
								"urn:endpoint:/test-cluster-name:10.100.200.20:10202",
								"urn:service:/test-cluster-name:test-namespace:test-service-2",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-2",
								"namespace":         "test-namespace",
								"uid":               "test-service-2"},
							"spec": map[string]interface{}{
								"type":      "NodePort",
								"clusterIP": "10.100.200.20",
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-node-port-2",
										"nodePort":   float64(10202),
										"port":       float64(82),
										"targetPort": float64(8082)},
								}},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-2",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-2",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "NodePort",
							},
							"identifiers": []string{
								"urn:endpoint:/test-cluster-name:10.100.200.20",
								"urn:endpoint:/test-cluster-name:10.100.200.20:10202",
								"urn:service:/test-cluster-name:test-namespace:test-service-2",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-2",
								"namespace":         "test-namespace",
								"uid":               "test-service-2",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationService},
							},
							"spec": map[string]interface{}{
								"type":      "NodePort",
								"clusterIP": "10.100.200.20",
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-node-port-2",
										"nodePort":   float64(10202),
										"port":       float64(82),
										"targetPort": float64(8082)},
								}},
							"status": map[string]interface{}{
								"loadBalancer": map[string]interface{}{},
							},
						},
					},
				),
			},
			expectedRelations: []*topology.Relation{
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
						"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-2",
					Type:     topology.Type{Name: "encloses"},
					SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
					TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-2",
					Data:     map[string]interface{}{},
				},
			},
		},
		{
			testCase: "Test Service 3 - Minimal - Cluster IP + External IPs",
			expectedComponents: []*topology.Component{
				chooseBySourcePropertiesFeature(
					sourcePropertiesEnabled,
					kubernetesStatusEnabled,
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-3",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name":              "test-service-3",
							"kind":              "Service",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "ClusterIP",
							},
							"uid": types.UID("test-service-3"),
							"identifiers": []string{
								"urn:endpoint:/34.100.200.12:83", "urn:endpoint:/34.100.200.13:83",
								"urn:endpoint:/test-cluster-name:10.100.200.21",
								"urn:service:/test-cluster-name:test-namespace:test-service-3",
							},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-3",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "ClusterIP",
							},
							"identifiers": []string{
								"urn:endpoint:/34.100.200.12:83", "urn:endpoint:/34.100.200.13:83",
								"urn:endpoint:/test-cluster-name:10.100.200.21",
								"urn:service:/test-cluster-name:test-namespace:test-service-3",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-3",
								"namespace":         "test-namespace",
								"uid":               "test-service-3"},
							"spec": map[string]interface{}{
								"type":        "ClusterIP",
								"clusterIP":   "10.100.200.21",
								"externalIPs": []interface{}{"34.100.200.12", "34.100.200.13"},
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-port-3",
										"port":       float64(83),
										"targetPort": float64(8083)},
								}},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-3",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-3",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "ClusterIP",
							},
							"identifiers": []string{
								"urn:endpoint:/34.100.200.12:83", "urn:endpoint:/34.100.200.13:83",
								"urn:endpoint:/test-cluster-name:10.100.200.21",
								"urn:service:/test-cluster-name:test-namespace:test-service-3",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-3",
								"namespace":         "test-namespace",
								"uid":               "test-service-3",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationService},
							},
							"spec": map[string]interface{}{
								"type":        "ClusterIP",
								"clusterIP":   "10.100.200.21",
								"externalIPs": []interface{}{"34.100.200.12", "34.100.200.13"},
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-port-3",
										"port":       float64(83),
										"targetPort": float64(8083),
									},
								},
							},
							"status": map[string]interface{}{
								"loadBalancer": map[string]interface{}{},
							},
						},
					},
				),
			},
			expectedRelations: []*topology.Relation{
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
						"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-3",
					Type:     topology.Type{Name: "encloses"},
					SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
					TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-3",
					Data:     map[string]interface{}{},
				},
			},
		},
		{
			testCase: "Test Service 4 - Minimal - Cluster IP",
			expectedComponents: []*topology.Component{
				chooseBySourcePropertiesFeature(
					sourcePropertiesEnabled,
					kubernetesStatusEnabled,
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-4",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name":              "test-service-4",
							"kind":              "Service",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "ClusterIP",
							},
							"uid": types.UID("test-service-4"),
							"identifiers": []string{
								"urn:endpoint:/test-cluster-name:10.100.200.22",
								"urn:service:/test-cluster-name:test-namespace:test-service-4",
							},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-4",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-4",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "ClusterIP",
							},
							"identifiers": []string{
								"urn:endpoint:/test-cluster-name:10.100.200.22",
								"urn:service:/test-cluster-name:test-namespace:test-service-4",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-4",
								"namespace":         "test-namespace",
								"uid":               "test-service-4"},
							"spec": map[string]interface{}{
								"type":      "ClusterIP",
								"clusterIP": "10.100.200.22",
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-port-4",
										"port":       float64(84),
										"targetPort": float64(8084)},
								}},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-4",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-4",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "ClusterIP",
							},
							"identifiers": []string{
								"urn:endpoint:/test-cluster-name:10.100.200.22",
								"urn:service:/test-cluster-name:test-namespace:test-service-4",
							},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-4",
								"namespace":         "test-namespace",
								"uid":               "test-service-4",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationService},
							},
							"spec": map[string]interface{}{
								"type":      "ClusterIP",
								"clusterIP": "10.100.200.22",
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-port-4",
										"port":       float64(84),
										"targetPort": float64(8084)},
								},
							},
							"status": map[string]interface{}{
								"loadBalancer": map[string]interface{}{},
							},
						},
					},
				),
			},
			expectedRelations: []*topology.Relation{
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
						"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-4",
					Type:     topology.Type{Name: "encloses"},
					SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
					TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-4",
					Data:     map[string]interface{}{},
				},
			},
		},
		{
			testCase: "Test Service 5 - Minimal - Cluster IP - None",
			expectedComponents: []*topology.Component{
				chooseBySourcePropertiesFeature(
					sourcePropertiesEnabled,
					kubernetesStatusEnabled,
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-5",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name":              "test-service-5",
							"kind":              "Service",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service":        "headless",
								"service-type":   "ClusterIP",
							},
							"uid":         types.UID("test-service-5"),
							"identifiers": []string{"urn:service:/test-cluster-name:test-namespace:test-service-5"},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-5",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-5",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service":        "headless",
								"service-type":   "ClusterIP",
							},
							"identifiers": []string{"urn:service:/test-cluster-name:test-namespace:test-service-5"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-5",
								"namespace":         "test-namespace",
								"uid":               "test-service-5"},
							"spec": map[string]interface{}{
								"type":      "ClusterIP",
								"clusterIP": "None",
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-port-5",
										"port":       float64(85),
										"targetPort": float64(8085)},
								}},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-5",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-5",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service":        "headless",
								"service-type":   "ClusterIP",
							},
							"identifiers": []string{"urn:service:/test-cluster-name:test-namespace:test-service-5"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-5",
								"namespace":         "test-namespace",
								"uid":               "test-service-5",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationService},
							},
							"spec": map[string]interface{}{
								"type":      "ClusterIP",
								"clusterIP": "None",
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-port-5",
										"port":       float64(85),
										"targetPort": float64(8085)},
								},
							},
							"status": map[string]interface{}{
								"loadBalancer": map[string]interface{}{},
							},
						},
					},
				),
			},
			expectedRelations: []*topology.Relation{
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
						"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-5",
					Type:     topology.Type{Name: "encloses"},
					SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
					TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-5",
					Data:     map[string]interface{}{},
				},
			},
		},
		testCase6,
		{
			testCase: "Test Service 7 - ExternalName Service",
			expectedComponents: []*topology.Component{
				chooseBySourcePropertiesFeature(
					sourcePropertiesEnabled,
					kubernetesStatusEnabled,
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-7",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name":              "test-service-7",
							"kind":              "Service",
							"creationTimestamp": creationTime,
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "ExternalName",
							},
							"uid":         types.UID("test-service-7"),
							"identifiers": []string{"urn:service:/test-cluster-name:test-namespace:test-service-7"},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-7",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-7",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "ExternalName",
							},
							"identifiers": []string{"urn:service:/test-cluster-name:test-namespace:test-service-7"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-7",
								"namespace":         "test-namespace",
								"uid":               "test-service-7"},
							"spec": map[string]interface{}{
								"type":         "ExternalName",
								"externalName": "mysql-db.host.example.com",
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-port-7",
										"port":       float64(87),
										"targetPort": float64(8087)},
								}},
						},
					},
					&topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-7",
						Type:       topology.Type{Name: "service"},
						Data: topology.Data{
							"name": "test-service-7",
							"tags": map[string]string{
								"test":           "label",
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-service",
								"namespace":      "test-namespace",
								"service-type":   "ExternalName",
							},
							"identifiers": []string{"urn:service:/test-cluster-name:test-namespace:test-service-7"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Service",
							"metadata": map[string]interface{}{
								"creationTimestamp": creationTimeFormatted,
								"labels":            map[string]interface{}{"test": "label"},
								"name":              "test-service-7",
								"namespace":         "test-namespace",
								"uid":               "test-service-7",
								"resourceVersion":   "123",
								"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationService},
							},
							"spec": map[string]interface{}{
								"type":         "ExternalName",
								"externalName": "mysql-db.host.example.com",
								"ports": []interface{}{
									map[string]interface{}{
										"name":       "test-service-port-7",
										"port":       float64(87),
										"targetPort": float64(8087)},
								}},
							"status": map[string]interface{}{
								"loadBalancer": map[string]interface{}{},
							},
						},
					},
				),
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:external-service/test-service-7",
					Type:       topology.Type{Name: "external-service"},
					Data: topology.Data{
						"name":              "test-service-7",
						"kind":              "ExternalService",
						"creationTimestamp": creationTime,
						"tags": map[string]string{
							"test":           "label",
							"cluster-name":   "test-cluster-name",
							"cluster-type":   "kubernetes",
							"component-type": "kubernetes-externalservice",
							"namespace":      "test-namespace",
						},
						"uid": types.UID("test-service-7"),
						"identifiers": []string{
							"urn:endpoint:/mysql-db.host.example.com",
							"urn:endpoint:/test-cluster-name:mysql-db.host.example.com:87",
							"urn:endpoint:/test-cluster-name:10.10.42.42",
							"urn:endpoint:/test-cluster-name:10.10.42.42:87",
							"urn:endpoint:/test-cluster-name:10.10.42.43",
							"urn:endpoint:/test-cluster-name:10.10.42.43:87",
							"urn:external-service:/test-cluster-name:test-namespace:test-service-7",
						},
					},
				},
			},
			expectedRelations: []*topology.Relation{
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-7->" +
						"urn:kubernetes:/test-cluster-name:test-namespace:external-service/test-service-7",
					Type:     topology.Type{Name: "uses"},
					SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-7",
					TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:external-service/test-service-7",
					Data:     map[string]interface{}{},
				},
				{
					ExternalID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace->" +
						"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-7",
					Type:     topology.Type{Name: "encloses"},
					SourceID: "urn:kubernetes:/test-cluster-name:namespace/test-namespace",
					TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-7",
					Data:     map[string]interface{}{},
				},
			},
		},
	}
}

type serviceCollectorTestCase struct {
	testCase           string
	expectedComponents []*topology.Component
	expectedRelations  []*topology.Relation
}

type MockServiceAPICollectorClient struct {
	apiserver.APICollectorClient
	testCaseNumber int
}

func (m MockServiceAPICollectorClient) GetServices() ([]coreV1.Service, error) {
	services := make([]coreV1.Service, 0)
	i := m.testCaseNumber
	service := coreV1.Service{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              fmt.Sprintf("test-service-%d", i),
			CreationTimestamp: creationTime,
			Namespace:         "test-namespace",
			Labels: map[string]string{
				"test": "label",
			},
			UID:             types.UID(fmt.Sprintf("test-service-%d", i)),
			GenerateName:    "",
			ResourceVersion: "123",
			Annotations: map[string]string{
				"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationService,
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
		Spec: coreV1.ServiceSpec{
			Ports: []coreV1.ServicePort{
				{Name: fmt.Sprintf("test-service-port-%d", i), Port: int32(80 + i), TargetPort: intstr.FromInt(8080 + i)},
			},
			Type: coreV1.ServiceTypeClusterIP,
		},
		Status: coreV1.ServiceStatus{
			LoadBalancer: coreV1.LoadBalancerStatus{},
		},
	}

	if i == 2 {
		service.Spec.Type = coreV1.ServiceTypeNodePort
		service.Spec.Ports = []coreV1.ServicePort{
			{
				Name:       fmt.Sprintf("test-service-node-port-%d", i),
				Port:       int32(80 + i),
				TargetPort: intstr.FromInt(8080 + i),
				NodePort:   int32(10200 + i),
			},
		}
		service.Spec.ClusterIP = "10.100.200.20"
	}

	if i == 3 {
		service.Spec.Type = coreV1.ServiceTypeClusterIP
		service.Spec.ExternalIPs = []string{"34.100.200.12", "34.100.200.13"}
		service.Spec.ClusterIP = "10.100.200.21"
	}

	if i == 4 {
		service.Spec.Type = coreV1.ServiceTypeClusterIP
		service.Spec.ClusterIP = "10.100.200.22"
	}

	if i == 5 {
		service.Spec.Type = coreV1.ServiceTypeClusterIP
		service.Spec.ClusterIP = "None"
	}

	if i == 6 {
		service.Spec.Type = coreV1.ServiceTypeLoadBalancer
		service.Spec.Ports = []coreV1.ServicePort{
			{
				Name:       fmt.Sprintf("test-service-port-%d", i),
				Port:       int32(80 + i),
				TargetPort: intstr.FromInt(8080 + i),
			},
			{
				Name:       fmt.Sprintf("test-service-node-port-%d", i),
				Port:       int32(80 + i),
				TargetPort: intstr.FromInt(8080 + i),
				NodePort:   int32(10200 + i),
			},
		}
		service.Status.LoadBalancer = coreV1.LoadBalancerStatus{
			Ingress: []coreV1.LoadBalancerIngress{
				{IP: "34.100.200.15"},
				{Hostname: "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com"},
			},
		}
		service.Spec.LoadBalancerIP = "10.100.200.23"
	}

	if i == 7 {
		service.Spec.Type = coreV1.ServiceTypeExternalName
		service.Spec.Ports = []coreV1.ServicePort{
			{
				Name:       fmt.Sprintf("test-service-port-%d", i),
				Port:       int32(80 + i),
				TargetPort: intstr.FromInt(8080 + i),
			},
		}
		service.Spec.ExternalName = "mysql-db.host.example.com"
	}

	services = append(services, service)

	return services, nil
}

func (m MockServiceAPICollectorClient) GetEndpoints() ([]coreV1.Endpoints, error) {
	endpoints := make([]coreV1.Endpoints, 0)

	return endpoints, nil
}
