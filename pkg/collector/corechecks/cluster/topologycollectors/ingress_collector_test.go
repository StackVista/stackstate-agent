// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	netV1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/version"
	"testing"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestIngressCollector_1_21(t *testing.T) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	k8sVersion := version.Info{
		Major: "1",
		Minor: "21",
	}

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		ic := NewIngressCollector(componentChannel, relationChannel, NewTestCommonClusterCollectorWithVersion(MockIngressAPICollectorClient{}, sourcePropertiesEnabled, &k8sVersion))
		expectedCollectorName := "Ingress Collector"
		RunCollectorTest(t, ic, expectedCollectorName)

		for _, tc := range []struct {
			testCase   string
			assertions []func(*testing.T, chan *topology.Component, chan *topology.Relation)
		}{
			{
				testCase: "Test Service 1.21 1 - Minimal",
				assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
					expectIngress211(sourcePropertiesEnabled, creationTimeFormatted),
					expectEndpoint211(),
					expectRelationEndpointIPIngress211(),
					expectEndpointAmazon21(),
					expectRelationEndpointAmazonIngress211(),
				},
			},
			{
				testCase: "Test Service 1.21 2 - Default Backend",
				assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
					expectIngress212(sourcePropertiesEnabled, creationTimeFormatted),
					expectRelationIngress212Service(),
					expectEndpointIP212(),
					expectRelationEndpointIP21Ingress212(),
					expectEndpointAmazon212(),
					expectRelationEndpointAmazonIngress212(),
				},
			},
			{
				testCase: "Test Service 1.21 3 - Ingress Rules",
				assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
					expectIngress213(sourcePropertiesEnabled, creationTimeFormatted),
					expectRelationIngress213Service1(),
					expectRelationIngress213Service2(),
					expectRalationIngress213Service3(),
					expectEndpointIP213(),
					expectRelationEndpointIPIngress213(),
					expectEndpointAmazon213(),
					expectRelationEndpointAmazonIngress213(),
				},
			},
			{
				testCase: "Test Service 1.22 1 - Minimal",
				assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
					expectIngress221(sourcePropertiesEnabled, creationTimeFormatted),
					expectEndpointIP221(),
					expectRelationEndpointIngress221(),
					expectEndpointAmazon22(),
					expectRelationEndpointAmazonIngress221(),
				},
			},
			{
				testCase: "Test Service 1.22 2 - Default Backend",
				assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
					expectIngress222(sourcePropertiesEnabled, creationTimeFormatted),
					expectRelationIngressService222(),
					expectEndpointIP222(),
					expectRelationEndpointIPIngress222(),
					expectEndpointAmazon22(),
					expectRelationEndpointAmazonIngress222(),
				},
			},
			{
				testCase: "Test Service 1.22 3 - Ingress Rules",
				assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
					expectIngress223(sourcePropertiesEnabled, creationTimeFormatted),
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
			t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled), func(t *testing.T) {
				for _, a := range tc.assertions {
					a(t, componentChannel, relationChannel)
				}
			})
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
		ic := NewIngressCollector(componentChannel, relationChannel, NewTestCommonClusterCollectorWithVersion(MockIngressAPICollectorClient{}, sourcePropertiesEnabled, &k8sVersion))
		expectedCollectorName := "Ingress Collector"
		RunCollectorTest(t, ic, expectedCollectorName)

		for _, tc := range []struct {
			testCase   string
			assertions []func(*testing.T, chan *topology.Component, chan *topology.Relation)
		}{
			{
				testCase: "Test Service 1.22 1 - Minimal",
				assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
					expectIngress221(sourcePropertiesEnabled, creationTimeFormatted),
					expectEndpointIP221(),
					expectRelationEndpointIngress221(),
					expectEndpointAmazon22(),
					expectRelationEndpointAmazonIngress221(),
				},
			},
			{
				testCase: "Test Service 1.22 2 - Default Backend",
				assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
					expectIngress222(sourcePropertiesEnabled, creationTimeFormatted),
					expectRelationIngressService222(),
					expectEndpointIP222(),
					expectRelationEndpointIPIngress222(),
					expectEndpointAmazon22(),
					expectRelationEndpointAmazonIngress222(),
				},
			},
			{
				testCase: "Test Service 1.22 3 - Ingress Rules",
				assertions: []func(*testing.T, chan *topology.Component, chan *topology.Relation){
					expectIngress223(sourcePropertiesEnabled, creationTimeFormatted),
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
			t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled), func(t *testing.T) {
				for _, a := range tc.assertions {
					a(t, componentChannel, relationChannel)
				}
			})
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
		},
	})
}

func expectRalationIngress213Service3() func(*testing.T, chan *topology.Component, chan *topology.Relation) {
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

func expectIngress213(sourcePropertiesEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress-3",
				"creationTimestamp": creationTime,
				"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"uid":               types.UID("test-ingress-3"),
				"kind":              "some-specified-kind",
				"generateName":      "some-specified-generation",
				"identifiers":       []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":        "test-ingress-3",
				"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
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
							"ingressRuleValue": map[string]interface{}{
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
											"path": "host-1-path-2"}}}}},
						map[string]interface{}{
							"host": "host-2",
							"ingressRuleValue": map[string]interface{}{
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"backend": map[string]interface{}{
												"serviceName": "test-service-3",
												"servicePort": float64(0)},
											"path": "host-2-path-1"}}}}}}},
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
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

func expectIngress212(sourcePropertiesEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress-2",
				"creationTimestamp": creationTime,
				"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"uid":               types.UID("test-ingress-2"),
				"identifiers":       []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":        "test-ingress-2",
				"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
		},
	})
}

func expectIngress211(sourcePropertiesEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress-1",
				"creationTimestamp": creationTime,
				"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"uid":               types.UID("test-ingress-1"),
				"identifiers":       []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":        "test-ingress-1",
				"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
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

func expectIngress223(sourcePropertiesEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress22-3",
				"creationTimestamp": creationTime,
				"tags":              map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"uid":               types.UID("test-ingress22-3"),
				"kind":              "some-specified-kind",
				"generateName":      "some-specified-generation",
				"identifiers":       []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-3",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":        "test-ingress22-3",
				"tags":        map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
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
							"ingressRuleValue": map[string]interface{}{
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"path": "host-1-path-1",
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
											"path": "host-1-path-2"}}}}},
						map[string]interface{}{
							"host": "host22-2",
							"ingressRuleValue": map[string]interface{}{
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"path": "host-2-path-1",
											"backend": map[string]interface{}{
												"service": map[string]interface{}{
													"name": "test-service22-3",
													"port": map[string]interface{}{
														"number": float64(22),
													}}}}}}}}}},
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
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

func expectIngress222(sourcePropertiesEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress22-2",
				"creationTimestamp": creationTime,
				"tags":              map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"uid":               types.UID("test-ingress22-2"),
				"identifiers":       []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-2",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":        "test-ingress22-2",
				"tags":        map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
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
			"creationTimestamp": creationTime,
			"tags":              map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
			"identifiers":       []string{},
		},
	})
}

func expectIngress221(sourcePropertiesEnabled bool, creationTimeFormatted string) func(*testing.T, chan *topology.Component, chan *topology.Relation) {
	return expectComponent(chooseBySourcePropertiesFeature(
		sourcePropertiesEnabled,
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":              "test-ingress22-1",
				"creationTimestamp": creationTime,
				"tags":              map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"uid":               types.UID("test-ingress22-1"),
				"identifiers":       []string{},
			},
		},
		&topology.Component{
			ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress22-1",
			Type:       topology.Type{Name: "ingress"},
			Data: topology.Data{
				"name":        "test-ingress22-1",
				"tags":        map[string]string{"test": "label22", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
				"identifiers": []string{},
			},
			SourceProperties: map[string]interface{}{
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
				Kind: "",
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
				LoadBalancer: coreV1.LoadBalancerStatus{
					Ingress: []coreV1.LoadBalancerIngress{
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
			ingress.TypeMeta.Kind = "some-specified-kind"
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
				Kind: "",
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
				LoadBalancer: coreV1.LoadBalancerStatus{
					Ingress: []coreV1.LoadBalancerIngress{
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
			ingress.TypeMeta.Kind = "some-specified-kind"
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
			Kind: "",
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
			LoadBalancer: coreV1.LoadBalancerStatus{
				Ingress: []coreV1.LoadBalancerIngress{
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
		Minor: "21",
	}

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		ic := NewIngressCollector(componentChannel, relationChannel, NewTestCommonClusterCollectorWithVersion(MockIngressAPICollectorClientNoHTTPRule{}, sourcePropertiesEnabled, &versionInfo))
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
						&topology.Component{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress",
							Type:       topology.Type{Name: "ingress"},
							Data: topology.Data{
								"name":              "test-ingress",
								"creationTimestamp": creationTime,
								"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
								"uid":               types.UID("test-ingress"),
								"identifiers":       []string{},
							},
						},
						&topology.Component{
							ExternalID: "urn:kubernetes:/test-cluster-name:test-namespace:ingress/test-ingress",
							Type:       topology.Type{Name: "ingress"},
							Data: topology.Data{
								"name":        "test-ingress",
								"tags":        map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
								"identifiers": []string{},
							},
							SourceProperties: map[string]interface{}{
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
											"host":             "host-1",
											"ingressRuleValue": map[string]interface{}{},
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
							"creationTimestamp": creationTime,
							"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
							"identifiers":       []string{},
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
							"creationTimestamp": creationTime,
							"tags":              map[string]string{"test": "label", "cluster-name": "test-cluster-name", "namespace": "test-namespace"},
							"identifiers":       []string{},
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
			t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled), func(t *testing.T) {
				for _, a := range tc.assertions {
					a(t, componentChannel, relationChannel)
				}
			})
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
