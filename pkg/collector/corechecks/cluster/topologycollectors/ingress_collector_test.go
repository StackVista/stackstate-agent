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

	netV1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/version"

	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var lastAppliedConfigurationIngress = `{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"annotations":{"kubernetes.io/ingress.class":"ingress-nginx-external","nginx.ingress.kubernetes.io/ingress.class":"ingress-nginx-external"},"labels":{"app.kubernetes.io/managed-by":"Helm","app.kubernetes.io/version":"1.0.0","helm.sh/chart":"1.0.0"},"name":"app","namespace":"tenant"},"spec":{"rules":[{"host":"test.com","http":{"paths":[]}}],"tls":[]}}`

func TestIngressCollector_1_18(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	k8sVersion := version.Info{
		Major: "1",
		Minor: "18",
	}

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			commonClusterCollector := NewTestCommonClusterCollectorWithVersion(MockIngressAPICollectorClient{}, sourcePropertiesEnabled, componentChannel, relationChannel, &k8sVersion, kubernetesStatusEnabled)
			commonClusterCollector.SetUseRelationCache(false)
			ic := NewIngressCollector(commonClusterCollector)
			expectedCollectorName := "Ingress Collector"
			RunCollectorTest(t, ic, expectedCollectorName)

			for _, tc := range []struct {
				testCase   string
				assertions []func(*testing.T, chan *topology.Component, chan *topology.Relation)
			}{
				{
					testCase: "Test Service 1 - Minimal",
					assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
						expectIngress211(sourcePropertiesEnabled, kubernetesStatusEnabled, creationTimeFormatted),
						expectEndpoint211(),
						expectRelationEndpointIPIngress211(),
						expectEndpointAmazon21(),
						expectRelationEndpointAmazonIngress211(),
					},
				},
				{
					testCase: "Test Service 2 - Default Backend",
					assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
						expectIngress212(sourcePropertiesEnabled, kubernetesStatusEnabled, creationTimeFormatted),
						expectRelationIngress212Service(),
						expectEndpointIP212(),
						expectRelationEndpointIP21Ingress212(),
						expectEndpointAmazon212(),
						expectRelationEndpointAmazonIngress212(),
					},
				},
				{
					testCase: "Test Service 3 - Ingress Rules",
					assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
						expectIngress213(sourcePropertiesEnabled, kubernetesStatusEnabled, creationTimeFormatted),
						expectRelationIngress213Service1(),
						expectRelationIngress213Service2(),
						expectRelationIngress213Service3(),
						expectEndpointIP213(),
						expectRelationEndpointIPIngress213(),
						expectEndpointAmazon213(),
						expectRelationEndpointAmazonIngress213(),
					},
				},
			} {
				t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
					for _, a := range tc.assertions {
						a(t, componentChannel, relationChannel)
					}
				})
			}
		}
	}
}

func TestIngressCollector_1_22(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	k8sVersion := version.Info{
		Major: "1",
		Minor: "22",
	}

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			commonClusterCollector := NewTestCommonClusterCollectorWithVersion(MockIngressAPICollectorClient{}, sourcePropertiesEnabled, componentChannel, relationChannel, &k8sVersion, kubernetesStatusEnabled)
			commonClusterCollector.SetUseRelationCache(false)
			ic := NewIngressCollector(commonClusterCollector)
			expectedCollectorName := "Ingress Collector"
			RunCollectorTest(t, ic, expectedCollectorName)

			for _, tc := range []struct {
				testCase   string
				assertions []func(*testing.T, chan *topology.Component, chan *topology.Relation)
			}{
				{
					testCase: "Test Service 1 - Minimal",
					assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
						expectIngress221(sourcePropertiesEnabled, kubernetesStatusEnabled, creationTimeFormatted),
						expectEndpointIP221(),
						expectRelationEndpointIngress221(),
						expectEndpointAmazon22(),
						expectRelationEndpointAmazonIngress221(),
					},
				},
				{
					testCase: "Test Service 2 - Default Backend",
					assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
						expectIngress222(sourcePropertiesEnabled, kubernetesStatusEnabled, creationTimeFormatted),
						expectRelationIngressService222(),
						expectEndpointIP222(),
						expectRelationEndpointIPIngress222(),
						expectEndpointAmazon22(),
						expectRelationEndpointAmazonIngress222(),
					},
				},
				{
					testCase: "Test Service 3 - Ingress Rules",
					assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
						expectIngress223(sourcePropertiesEnabled, kubernetesStatusEnabled, creationTimeFormatted),
						expectRelationIngress223Service1(),
						expectRelationIngress223Service2(),
						expectRelationIngress223Service3(),
						expectEndpointIP223(),
						expectRelationEndpointIP22Ingress223(),
						expectEndpointAmazon22(),
						expectRelationEndpointAmazonIngress223(),
					},
				},
			} {
				t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
					for _, a := range tc.assertions {
						a(t, componentChannel, relationChannel)
					}
				})
			}
		}
	}
}

func expectRelationEndpointAmazonIngress213() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
		SourceID:   "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpointAmazon213() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectRelationEndpointIPIngress213() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.15->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
		SourceID:   "urn:endpoint:/test-cluster-name:34.100.200.15",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpointIP213() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.15",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "34.100.200.15",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectRelationIngress213Service3() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3->" +
			"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-3",
		Type:     topology.Type{Name: "routes"},
		SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
		TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-3",
		Data:     map[string]interface{}{},
	})
}

func expectRelationIngress213Service2() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3->" +
			"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-2",
		Type:     topology.Type{Name: "routes"},
		SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
		TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-2",
		Data:     map[string]interface{}{},
	})
}

func expectRelationIngress213Service1() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3->" +
			"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-1",
		Type:     topology.Type{Name: "routes"},
		SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
		TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service-1",
		Data:     map[string]interface{}{},
	})
}

func expectIngress213(sourcePropertiesEnabled bool, kubernetesStatusEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		kubernetesStatusEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress-3",
				"kind":              "Ingress",
				"creationTimestamp": creationTime,
				"tags": map[string]string{
					"test":           "label",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"uid":          types.UID("test-ingress-3"),
				"generateName": "some-specified-generation",
				"identifiers":  []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress-3",
				"tags": map[string]string{
					"test":           "label",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label"},
					"name":              "test-ingress-3",
					"namespace":         "test-namespace",
					"uid":               "test-ingress-3",
					"generateName":      "some-specified-generation",
				},
				"spec": map[string]interface{}{
					"rules": []interface{}{
						map[string]interface{}{
							"host": "host-1",
							"http": map[string]interface{}{
								"paths": []interface{}{
									map[string]interface{}{
										"backend": map[string]interface{}{
											"serviceName": "test-service-1",
											"servicePort": float64(0)},
										"path": "host-1-path-1"},
									map[string]interface{}{
										"backend": map[string]interface{}{
											"serviceName": "test-service-2",
											"servicePort": float64(0)},
										"path": "host-1-path-2"}}}},
						map[string]interface{}{
							"host": "host-2",
							"http": map[string]interface{}{
								"paths": []interface{}{
									map[string]interface{}{
										"backend": map[string]interface{}{
											"serviceName": "test-service-3",
											"servicePort": float64(0)},
										"path": "host-2-path-1"}}}}}},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress-3",
				"tags": map[string]string{
					"test":           "label",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label"},
					"name":              "test-ingress-3",
					"namespace":         "test-namespace",
					"uid":               "test-ingress-3",
					"generateName":      "some-specified-generation",
					"resourceVersion":   "123",
					"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress},
				},
				"spec": map[string]interface{}{
					"rules": []interface{}{
						map[string]interface{}{
							"host": "host-1",
							"http": map[string]interface{}{
								"paths": []interface{}{
									map[string]interface{}{
										"backend": map[string]interface{}{
											"serviceName": "test-service-1",
											"servicePort": float64(0)},
										"path": "host-1-path-1"},
									map[string]interface{}{
										"backend": map[string]interface{}{
											"serviceName": "test-service-2",
											"servicePort": float64(0)},
										"path": "host-1-path-2"}}}},
						map[string]interface{}{
							"host": "host-2",
							"http": map[string]interface{}{
								"paths": []interface{}{
									map[string]interface{}{
										"backend": map[string]interface{}{
											"serviceName": "test-service-3",
											"servicePort": float64(0)},
										"path": "host-2-path-1"}}}}}},
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
	))
}

func expectRelationEndpointAmazonIngress212() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
		SourceID:   "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpointAmazon212() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectRelationEndpointIP21Ingress212() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.15->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
		SourceID:   "urn:endpoint:/test-cluster-name:34.100.200.15",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpointIP212() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.15",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "34.100.200.15",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectRelationIngress212Service() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2->" +
			"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service",
		Type:     topology.Type{Name: "routes"},
		SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
		TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service",
		Data:     map[string]interface{}{},
	})
}

func expectIngress212(sourcePropertiesEnabled bool, kubernetesStatusEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		kubernetesStatusEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress-2",
				"kind":              "Ingress",
				"creationTimestamp": creationTime,
				"tags": map[string]string{
					"test":           "label",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"uid":         types.UID("test-ingress-2"),
				"identifiers": []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress-2",
				"tags": map[string]string{
					"test":           "label",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label"},
					"name":              "test-ingress-2",
					"namespace":         "test-namespace",
					"uid":               "test-ingress-2",
				},
				"spec": map[string]interface{}{
					"backend": map[string]interface{}{
						"serviceName": "test-service",
						"servicePort": float64(0),
					},
				},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress-2",
				"tags": map[string]string{
					"test":           "label",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label"},
					"name":              "test-ingress-2",
					"namespace":         "test-namespace",
					"uid":               "test-ingress-2",
					"resourceVersion":   "123",
					"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress},
				},
				"spec": map[string]interface{}{
					"backend": map[string]interface{}{
						"serviceName": "test-service",
						"servicePort": float64(0),
					},
				},
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
	))
}

func expectRelationEndpointAmazonIngress211() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-1",
		SourceID:   "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-1",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpointAmazon21() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectRelationEndpointIPIngress211() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.15->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-1",
		SourceID:   "urn:endpoint:/test-cluster-name:34.100.200.15",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-1",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpoint211() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.15",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "34.100.200.15",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectIngress211(sourcePropertiesEnabled bool, kubernetesStatusEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		kubernetesStatusEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress-1",
				"kind":              "Ingress",
				"creationTimestamp": creationTime,
				"tags": map[string]string{
					"test":           "label",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"uid":         types.UID("test-ingress-1"),
				"identifiers": []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress-1",
				"tags": map[string]string{
					"test":           "label",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label"},
					"name":              "test-ingress-1",
					"namespace":         "test-namespace",
					"uid":               "test-ingress-1",
				},
				"spec": map[string]interface{}{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress-1",
				"tags": map[string]string{
					"test":           "label",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label"},
					"name":              "test-ingress-1",
					"namespace":         "test-namespace",
					"uid":               "test-ingress-1",
					"resourceVersion":   "123",
					"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress},
				},
				"spec": map[string]interface{}{},
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
	))
}

func expectRelationEndpointAmazonIngress223() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
		SourceID:   "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpointAmazon22() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label22",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectRelationEndpointIP22Ingress223() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.22->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
		SourceID:   "urn:endpoint:/test-cluster-name:34.100.200.22",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpointIP223() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.22",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "34.100.200.22",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label22",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectRelationIngress223Service3() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3->" +
			"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service22-3",
		Type:     topology.Type{Name: "routes"},
		SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
		TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service22-3",
		Data:     map[string]interface{}{},
	})
}

func expectRelationIngress223Service2() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3->" +
			"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service22-2",
		Type:     topology.Type{Name: "routes"},
		SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
		TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service22-2",
		Data:     map[string]interface{}{},
	})
}

func expectRelationIngress223Service1() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3->" +
			"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service22-1",
		Type:     topology.Type{Name: "routes"},
		SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
		TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service22-1",
		Data:     map[string]interface{}{},
	})
}

func expectIngress223(sourcePropertiesEnabled bool, kubernetesStatusEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		kubernetesStatusEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress22-3",
				"creationTimestamp": creationTime,
				"tags": map[string]string{
					"test":           "label22",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"uid":          types.UID("test-ingress22-3"),
				"kind":         "Ingress",
				"generateName": "some-specified-generation",
				"identifiers":  []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress22-3",
				"tags": map[string]string{
					"test":           "label22",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "networking.k8s.io/v1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label22"},
					"name":              "test-ingress22-3",
					"namespace":         "test-namespace",
					"uid":               "test-ingress22-3",
					"generateName":      "some-specified-generation",
				},
				"spec": map[string]interface{}{
					"rules": []interface{}{
						map[string]interface{}{
							"host": "host22-1",
							"http": map[string]interface{}{
								"paths": []interface{}{
									map[string]interface{}{
										"path":     "host-1-path-1",
										"pathType": interface{}(nil),
										"backend": map[string]interface{}{
											"service": map[string]interface{}{
												"name": "test-service22-1",
												"port": map[string]interface{}{},
											},
										}},
									map[string]interface{}{
										"backend": map[string]interface{}{
											"service": map[string]interface{}{
												"name": "test-service22-2",
												"port": map[string]interface{}{},
											},
										},
										"path":     "host-1-path-2",
										"pathType": interface{}(nil)}}}},
						map[string]interface{}{
							"host": "host22-2",
							"http": map[string]interface{}{
								"paths": []interface{}{
									map[string]interface{}{
										"path":     "host-2-path-1",
										"pathType": interface{}(nil),
										"backend": map[string]interface{}{
											"service": map[string]interface{}{
												"name": "test-service22-3",
												"port": map[string]interface{}{
													"number": float64(22),
												}}}}}}}}},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress22-3",
				"tags": map[string]string{
					"test":           "label22",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "networking.k8s.io/v1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label22"},
					"name":              "test-ingress22-3",
					"namespace":         "test-namespace",
					"uid":               "test-ingress22-3",
					"generateName":      "some-specified-generation",
					"resourceVersion":   "123",
					"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress},
				},
				"spec": map[string]interface{}{
					"rules": []interface{}{
						map[string]interface{}{
							"host": "host22-1",
							"http": map[string]interface{}{
								"paths": []interface{}{
									map[string]interface{}{
										"path":     "host-1-path-1",
										"pathType": interface{}(nil),
										"backend": map[string]interface{}{
											"service": map[string]interface{}{
												"name": "test-service22-1",
												"port": map[string]interface{}{},
											},
										}},
									map[string]interface{}{
										"backend": map[string]interface{}{
											"service": map[string]interface{}{
												"name": "test-service22-2",
												"port": map[string]interface{}{},
											},
										},
										"path":     "host-1-path-2",
										"pathType": interface{}(nil)}}}},
						map[string]interface{}{
							"host": "host22-2",
							"http": map[string]interface{}{
								"paths": []interface{}{
									map[string]interface{}{
										"path":     "host-2-path-1",
										"pathType": interface{}(nil),
										"backend": map[string]interface{}{
											"service": map[string]interface{}{
												"name": "test-service22-3",
												"port": map[string]interface{}{
													"number": float64(22),
												}}}}}}}}},
				"status": map[string]interface{}{
					"loadBalancer": map[string]interface{}{
						"ingress": []interface{}{
							map[string]interface{}{
								"ip": "34.100.200.22",
							},
							map[string]interface{}{
								"hostname": "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
							},
						},
					},
				},
			},
		},
	))
}

func expectRelationEndpointAmazonIngress222() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
		SourceID:   "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectRelationEndpointIPIngress222() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.22->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
		SourceID:   "urn:endpoint:/test-cluster-name:34.100.200.22",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpointIP222() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.22",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "34.100.200.22",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label22",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectRelationIngressService222() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2->" +
			"urn:kubernetes:/test-cluster-name:test-namespace:service/test-service22",
		Type:     topology.Type{Name: "routes"},
		SourceID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
		TargetID: "urn:kubernetes:/test-cluster-name:test-namespace:service/test-service22",
		Data:     map[string]interface{}{},
	})
}

func expectIngress222(sourcePropertiesEnabled bool, kubernetesStatusEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		kubernetesStatusEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress22-2",
				"kind":              "Ingress",
				"creationTimestamp": creationTime,
				"tags": map[string]string{
					"test":           "label22",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"uid":         types.UID("test-ingress22-2"),
				"identifiers": []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress22-2",
				"tags": map[string]string{
					"test":           "label22",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "networking.k8s.io/v1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label22"},
					"name":              "test-ingress22-2",
					"namespace":         "test-namespace",
					"uid":               "test-ingress22-2",
				},
				"spec": map[string]interface{}{
					"defaultBackend": map[string]interface{}{
						"service": map[string]interface{}{
							"name": "test-service22",
							"port": map[string]interface{}{},
						},
					},
				},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress22-2",
				"tags": map[string]string{
					"test":           "label22",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "networking.k8s.io/v1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label22"},
					"name":              "test-ingress22-2",
					"namespace":         "test-namespace",
					"uid":               "test-ingress22-2",
					"resourceVersion":   "123",
					"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress},
				},
				"spec": map[string]interface{}{
					"defaultBackend": map[string]interface{}{
						"service": map[string]interface{}{
							"name": "test-service22",
							"port": map[string]interface{}{},
						},
					},
				},
				"status": map[string]interface{}{
					"loadBalancer": map[string]interface{}{
						"ingress": []interface{}{
							map[string]interface{}{
								"ip": "34.100.200.22",
							},
							map[string]interface{}{
								"hostname": "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
							},
						},
					},
				},
			},
		},
	))
}

func expectRelationEndpointAmazonIngress221() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-1",
		SourceID:   "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-1",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectRelationEndpointIngress221() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectRelation(&topology.Relation{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.22->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-1",
		SourceID:   "urn:endpoint:/test-cluster-name:34.100.200.22",
		TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-1",
		Type:       topology.Type{Name: "routes"},
		Data:       map[string]interface{}{},
	})
}

func expectEndpointIP221() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(&topology.Component{
		ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.22",
		Type:       topology.Type{Name: "endpoint"},
		Data: topology.Data{
			"name":              "34.100.200.22",
			"kind":              "Endpoint",
			"creationTimestamp": creationTime,
			"tags": map[string]string{
				"test":           "label22",
				"cluster-name":   "test-cluster-name",
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-endpoint",
				"namespace":      "test-namespace",
			},
			"identifiers": []string{},
		},
	})
}

func expectIngress221(sourcePropertiesEnabled bool, kubernetesStatusEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		kubernetesStatusEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress22-1",
				"kind":              "Ingress",
				"creationTimestamp": creationTime,
				"tags": map[string]string{
					"test":           "label22",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"uid":         types.UID("test-ingress22-1"),
				"identifiers": []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress22-1",
				"tags": map[string]string{
					"test":           "label22",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "networking.k8s.io/v1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label22"},
					"name":              "test-ingress22-1",
					"namespace":         "test-namespace",
					"uid":               "test-ingress22-1",
				},
				"spec": map[string]interface{}{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name": "test-ingress22-1",
				"tags": map[string]string{
					"test":           "label22",
					"cluster-name":   "test-cluster-name",
					"cluster-type":   "kubernetes",
					"component-type": "kubernetes-ingress",
					"namespace":      "test-namespace",
				},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
				"apiVersion": "networking.k8s.io/v1",
				"kind":       "Ingress",
				"metadata": map[string]interface{}{
					"creationTimestamp": creationTimeFormatted,
					"labels":            map[string]interface{}{"test": "label22"},
					"name":              "test-ingress22-1",
					"namespace":         "test-namespace",
					"uid":               "test-ingress22-1",
					"resourceVersion":   "123",
					"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress},
				},
				"spec": map[string]interface{}{},
				"status": map[string]interface{}{
					"loadBalancer": map[string]interface{}{
						"ingress": []interface{}{
							map[string]interface{}{
								"ip": "34.100.200.22",
							},
							map[string]interface{}{
								"hostname": "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
							},
						},
					},
				},
			},
		},
	))
}

type MockIngressAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockIngressAPICollectorClient) GetIngressesNetV1() ([]netV1.Ingress, error) {
	ingresses := make([]netV1.Ingress, 0)
	for i := 1; i <= 3; i++ {
		ingress := netV1.Ingress{
			TypeMeta: v1.TypeMeta{
				Kind: "Ingress",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-ingress22-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label22",
				},
				UID:             types.UID(fmt.Sprintf("test-ingress22-%d", i)),
				GenerateName:    "",
				ResourceVersion: "123",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress,
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
			Status: netV1.IngressStatus{
				LoadBalancer: netV1.IngressLoadBalancerStatus{
					Ingress: []netV1.IngressLoadBalancerIngress{
						{IP: "34.100.200.22"},
						{Hostname: "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com"},
					},
				},
			},
		}

		if i == 2 {
			ingress.Spec.DefaultBackend = &netV1.IngressBackend{Service: &netV1.IngressServiceBackend{
				Name: "test-service22",
			}}
		}

		if i == 3 {
			ingress.TypeMeta.Kind = "Ingress"
			ingress.ObjectMeta.GenerateName = "some-specified-generation"
			ingress.Spec.Rules = []netV1.IngressRule{
				{
					Host: "host22-1",
					IngressRuleValue: netV1.IngressRuleValue{
						HTTP: &netV1.HTTPIngressRuleValue{
							Paths: []netV1.HTTPIngressPath{
								{Path: "host-1-path-1", Backend: netV1.IngressBackend{Service: &netV1.IngressServiceBackend{
									Name: "test-service22-1",
								}}},
								{Path: "host-1-path-2", Backend: netV1.IngressBackend{Service: &netV1.IngressServiceBackend{
									Name: "test-service22-2",
								}}},
							},
						},
					},
				},
				{
					Host: "host22-2",
					IngressRuleValue: netV1.IngressRuleValue{
						HTTP: &netV1.HTTPIngressRuleValue{
							Paths: []netV1.HTTPIngressPath{
								{Path: "host-2-path-1", Backend: netV1.IngressBackend{Service: &netV1.IngressServiceBackend{
									Name: "test-service22-3",
									Port: netV1.ServiceBackendPort{
										Number: 22,
									},
								}}},
							},
						},
					},
				},
			}
		}

		ingresses = append(ingresses, ingress)
	}

	return ingresses, nil
}

func (m MockIngressAPICollectorClient) GetIngressesExtV1B1() ([]v1beta1.Ingress, error) {
	ingresses := make([]v1beta1.Ingress, 0)
	for i := 1; i <= 3; i++ {
		ingress := v1beta1.Ingress{
			TypeMeta: v1.TypeMeta{
				Kind: "Ingress",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:              fmt.Sprintf("test-ingress-%d", i),
				CreationTimestamp: creationTime,
				Namespace:         "test-namespace",
				Labels: map[string]string{
					"test": "label",
				},
				UID:             types.UID(fmt.Sprintf("test-ingress-%d", i)),
				GenerateName:    "",
				ResourceVersion: "123",
				Annotations: map[string]string{
					"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress,
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
			Status: v1beta1.IngressStatus{
				LoadBalancer: v1beta1.IngressLoadBalancerStatus{
					Ingress: []v1beta1.IngressLoadBalancerIngress{
						{IP: "34.100.200.15"},
						{Hostname: "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com"},
					},
				},
			},
		}

		if i == 2 {
			ingress.Spec.Backend = &v1beta1.IngressBackend{ServiceName: "test-service"}
		}

		if i == 3 {
			ingress.TypeMeta.Kind = "Ingress"
			ingress.ObjectMeta.GenerateName = "some-specified-generation"
			ingress.Spec.Rules = []v1beta1.IngressRule{
				{
					Host: "host-1",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{Path: "host-1-path-1", Backend: v1beta1.IngressBackend{ServiceName: "test-service-1"}},
								{Path: "host-1-path-2", Backend: v1beta1.IngressBackend{ServiceName: "test-service-2"}},
							},
						},
					},
				},
				{
					Host: "host-2",
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{Path: "host-2-path-1", Backend: v1beta1.IngressBackend{ServiceName: "test-service-3"}},
							},
						},
					},
				},
			}
		}

		ingresses = append(ingresses, ingress)
	}

	return ingresses, nil
}

type MockIngressAPICollectorClientNoHTTPRule struct {
	apiserver.APICollectorClient
}

func (m MockIngressAPICollectorClientNoHTTPRule) GetIngressesNetV1() ([]netV1.Ingress, error) {
	ingresses := make([]netV1.Ingress, 0)
	return ingresses, nil
}

func (m MockIngressAPICollectorClientNoHTTPRule) GetIngressesExtV1B1() ([]v1beta1.Ingress, error) {
	ingresses := make([]v1beta1.Ingress, 0)
	ingress := v1beta1.Ingress{
		TypeMeta: v1.TypeMeta{
			Kind: "Ingress",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              "test-ingress",
			CreationTimestamp: creationTime,
			Namespace:         "test-namespace",
			Labels: map[string]string{
				"test": "label",
			},
			UID:             types.UID("test-ingress"),
			GenerateName:    "",
			ResourceVersion: "123",
			Annotations: map[string]string{
				"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress,
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
		Status: v1beta1.IngressStatus{
			LoadBalancer: v1beta1.IngressLoadBalancerStatus{
				Ingress: []v1beta1.IngressLoadBalancerIngress{
					{IP: "34.100.200.15"},
					{Hostname: "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com"},
				},
			},
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: "host-1",
				},
			},
		},
	}

	ingresses = append(ingresses, ingress)
	return ingresses, nil
}

// Test for bug STAC-17811
func TestIngressCollector_NoHttpRule(t *testing.T) {
	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	versionInfo := version.Info{
		Major: "1",
		Minor: "18",
	}

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			commonClusterCollector := NewTestCommonClusterCollectorWithVersion(MockIngressAPICollectorClientNoHTTPRule{}, sourcePropertiesEnabled, componentChannel, relationChannel, &versionInfo, kubernetesStatusEnabled)
			commonClusterCollector.SetUseRelationCache(false)
			ic := NewIngressCollector(commonClusterCollector)
			expectedCollectorName := "Ingress Collector"
			RunCollectorTest(t, ic, expectedCollectorName)

			for _, tc := range []struct {
				testCase   string
				assertions []func(*testing.T, chan *topology.Component, chan *topology.Relation)
			}{
				{
					testCase: "Test Service 1.21 1 - Minimal",
					assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
						expectComponent(chooseBySourcePropertiesFeature(
							sourcePropertiesEnabled,
							kubernetesStatusEnabled,
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress",
								Type:       topology.Type{Name: "ingress"},
								Data: topology.Data{
									"name":              "test-ingress",
									"kind":              "Ingress",
									"creationTimestamp": creationTime,
									"tags": map[string]string{
										"test":           "label",
										"cluster-name":   "test-cluster-name",
										"cluster-type":   "kubernetes",
										"component-type": "kubernetes-ingress",
										"namespace":      "test-namespace",
									},
									"uid":         types.UID("test-ingress"),
									"identifiers": []string{},
								},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress",
								Type:       topology.Type{Name: "ingress"},
								Data: topology.Data{
									"name": "test-ingress",
									"tags": map[string]string{
										"test":           "label",
										"cluster-name":   "test-cluster-name",
										"cluster-type":   "kubernetes",
										"component-type": "kubernetes-ingress",
										"namespace":      "test-namespace",
									},
									"identifiers": []string{},
								},
								SourceProperties: map[string]interface{}{
									"apiVersion": "extensions/v1beta1",
									"kind":       "Ingress",
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-ingress",
										"namespace":         "test-namespace",
										"uid":               "test-ingress",
									},
									"spec": map[string]interface{}{
										"rules": []interface{}{
											map[string]interface{}{
												"host": "host-1",
											},
										},
									},
								},
							},
							&topology.Component{
								ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress",
								Type:       topology.Type{Name: "ingress"},
								Data: topology.Data{
									"name": "test-ingress",
									"tags": map[string]string{
										"test":           "label",
										"cluster-name":   "test-cluster-name",
										"cluster-type":   "kubernetes",
										"component-type": "kubernetes-ingress",
										"namespace":      "test-namespace",
									},
									"identifiers": []string{},
								},
								SourceProperties: map[string]interface{}{
									"apiVersion": "extensions/v1beta1",
									"kind":       "Ingress",
									"metadata": map[string]interface{}{
										"creationTimestamp": creationTimeFormatted,
										"labels":            map[string]interface{}{"test": "label"},
										"name":              "test-ingress",
										"namespace":         "test-namespace",
										"uid":               "test-ingress",
										"resourceVersion":   "123",
										"annotations":       map[string]interface{}{"kubectl.kubernetes.io/last-applied-configuration": lastAppliedConfigurationIngress},
									},
									"spec": map[string]interface{}{
										"rules": []interface{}{
											map[string]interface{}{
												"host": "host-1",
											},
										},
									},
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
						)),
						expectComponent(&topology.Component{
							ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.15",
							Type:       topology.Type{Name: "endpoint"},
							Data: topology.Data{
								"name":              "34.100.200.15",
								"kind":              "Endpoint",
								"creationTimestamp": creationTime,
								"tags": map[string]string{
									"test":           "label",
									"cluster-name":   "test-cluster-name",
									"cluster-type":   "kubernetes",
									"component-type": "kubernetes-endpoint",
									"namespace":      "test-namespace",
								},
								"identifiers": []string{},
							},
						}),
						expectRelation(&topology.Relation{
							ExternalID: "urn:endpoint:/test-cluster-name:34.100.200.15->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress",
							SourceID:   "urn:endpoint:/test-cluster-name:34.100.200.15",
							TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress",
							Type:       topology.Type{Name: "routes"},
							Data:       map[string]interface{}{},
						}),
						expectComponent(&topology.Component{
							ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
							Type:       topology.Type{Name: "endpoint"},
							Data: topology.Data{
								"name":              "64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
								"kind":              "Endpoint",
								"creationTimestamp": creationTime,
								"tags": map[string]string{
									"test":           "label",
									"cluster-name":   "test-cluster-name",
									"cluster-type":   "kubernetes",
									"component-type": "kubernetes-endpoint",
									"namespace":      "test-namespace",
								},
								"identifiers": []string{},
							},
						}),
						expectRelation(&topology.Relation{
							ExternalID: "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com->urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress",
							SourceID:   "urn:endpoint:/test-cluster-name:64047e8f24bb48e9a406ac8286ee8b7d.eu-west-1.elb.amazonaws.com",
							TargetID:   "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress",
							Type:       topology.Type{Name: "routes"},
							Data:       map[string]interface{}{},
						}),
					},
				},
			} {
				t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
					for _, a := range tc.assertions {
						a(t, componentChannel, relationChannel)
					}
				})
			}
		}
	}
}

func expectComponent(expected *topology.Component) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return func(t *testing.T, componentChan chan *topology.Component, _ chan *topology.Relation) {
		c := <-componentChan
		assert.EqualValues(t, expected, c)
	}
}

func expectRelation(expected *topology.Relation) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return func(t *testing.T, _ chan *topology.Component, relationChan chan *topology.Relation) {
		r := <-relationChan
		assert.EqualValues(t, expected, r)
	}
}
