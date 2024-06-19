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

	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestNodeCollector(t *testing.T) {
	mockConfig := config.Mock(t)
	var testClusterName = "test-cluster-name"
	mockConfig.SetWithoutSource("cluster_name", testClusterName)

	creationTime = v1.Time{Time: time.Now().Add(-1 * time.Hour)}
	creationTimeFormatted := creationTime.UTC().Format(time.RFC3339)

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {
			(func() {
				componentChannel := make(chan *topology.Component)
				defer close(componentChannel)
				relationChannel := make(chan *topology.Relation)
				defer close(relationChannel)
				nodeIdentifierCorrelationChannel := make(chan *NodeIdentifierCorrelation)

				commonClusterCollector := NewTestCommonClusterCollector(MockNodeAPICollectorClient{}, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled)
				commonClusterCollector.SetUseRelationCache(false)
				ic := NewNodeCollector(nodeIdentifierCorrelationChannel, commonClusterCollector)
				expectedCollectorName := "Node Collector"
				RunCollectorTest(t, ic, expectedCollectorName)

				for _, tc := range []struct {
					testCase   string
					assertions []func()
				}{
					{
						testCase: "Test Node 1 - NodeInternalIP",
						assertions: []func(){
							func() {
								component := <-componentChannel
								expectedComponent := chooseBySourcePropertiesFeature(
									sourcePropertiesEnabled,
									kubernetesStatusEnabled,
									&topology.Component{
										ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-1",
										Type:       topology.Type{Name: "node"},
										Data: topology.Data{
											"name":              "test-node-1",
											"kind":              "Node",
											"creationTimestamp": creationTime,
											"tags": map[string]string{
												"test":           "label",
												"cluster-name":   "test-cluster-name",
												"cluster-type":   "kubernetes",
												"component-type": "kubernetes-node",
												"namespace":      "test-namespace",
											},
											"uid":        types.UID("test-node-1"),
											"instanceId": "test-node-1-test-cluster-name",
											"sts_host":   "test-node-1-test-cluster-name",
											"status": NodeStatus{
												Phase: coreV1.NodeRunning,
												NodeInfo: coreV1.NodeSystemInfo{
													MachineID:     "test-machine-id-1",
													KernelVersion: "4.19.0",
													Architecture:  "x86_64",
												},
												KubeletEndpoint: coreV1.DaemonEndpoint{Port: 5000},
											},
											"identifiers": []string{
												"urn:ip:/test-cluster-name:test-node-1:10.20.01.01",
												"urn:host:/test-node-1-test-cluster-name",
											},
										},
									},
									&topology.Component{
										ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-1",
										Type:       topology.Type{Name: "node"},
										Data: topology.Data{
											"name": "test-node-1",
											"tags": map[string]string{
												"test":           "label",
												"cluster-name":   "test-cluster-name",
												"cluster-type":   "kubernetes",
												"component-type": "kubernetes-node",
												"namespace":      "test-namespace",
											},
											"instanceId": "test-node-1-test-cluster-name",
											"sts_host":   "test-node-1-test-cluster-name",
											"identifiers": []string{
												"urn:ip:/test-cluster-name:test-node-1:10.20.01.01",
												"urn:host:/test-node-1-test-cluster-name",
											},
										},
										SourceProperties: map[string]interface{}{
											"apiVersion": "v1",
											"kind":       "Node",
											"metadata": map[string]interface{}{
												"creationTimestamp": creationTimeFormatted,
												"labels":            map[string]interface{}{"test": "label"},
												"name":              "test-node-1",
												"namespace":         "test-namespace",
												"uid":               "test-node-1",
											},
											"spec": map[string]interface{}{},
											"status": map[string]interface{}{
												"phase": "Running",
												"nodeInfo": map[string]interface{}{
													"machineID":               "test-machine-id-1",
													"bootID":                  "",
													"containerRuntimeVersion": "",
													"kernelVersion":           "4.19.0",
													"architecture":            "x86_64",
													"kubeProxyVersion":        "",
													"kubeletVersion":          "",
													"operatingSystem":         "",
													"osImage":                 "",
													"systemUUID":              "",
												},
												"daemonEndpoints": map[string]interface{}{
													"kubeletEndpoint": map[string]interface{}{
														"Port": float64(5000),
													},
												},
											},
										},
									},
									&topology.Component{
										ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-1",
										Type:       topology.Type{Name: "node"},
										Data: topology.Data{
											"name": "test-node-1",
											"tags": map[string]string{
												"test":           "label",
												"cluster-name":   "test-cluster-name",
												"cluster-type":   "kubernetes",
												"component-type": "kubernetes-node",
												"namespace":      "test-namespace",
											},
											"instanceId": "test-node-1-test-cluster-name",
											"sts_host":   "test-node-1-test-cluster-name",
											"identifiers": []string{
												"urn:ip:/test-cluster-name:test-node-1:10.20.01.01",
												"urn:host:/test-node-1-test-cluster-name",
											},
										},
										SourceProperties: map[string]interface{}{
											"apiVersion": "v1",
											"kind":       "Node",
											"metadata": map[string]interface{}{
												"creationTimestamp": creationTimeFormatted,
												"labels":            map[string]interface{}{"test": "label"},
												"name":              "test-node-1",
												"namespace":         "test-namespace",
												"uid":               "test-node-1",
												"resourceVersion":   "123",
											},
											"spec": map[string]interface{}{},
											"status": map[string]interface{}{
												"addresses": []interface{}{
													map[string]interface{}{
														"address": "10.20.01.01",
														"type":    "InternalIP",
													},
												},
												"phase": "Running",
												"nodeInfo": map[string]interface{}{
													"machineID":               "test-machine-id-1",
													"bootID":                  "",
													"containerRuntimeVersion": "",
													"kernelVersion":           "4.19.0",
													"architecture":            "x86_64",
													"kubeProxyVersion":        "",
													"kubeletVersion":          "",
													"operatingSystem":         "",
													"osImage":                 "",
													"systemUUID":              "",
												},
												"daemonEndpoints": map[string]interface{}{
													"kubeletEndpoint": map[string]interface{}{
														"Port": float64(5000),
													},
												},
											},
										},
									},
								)
								assert.EqualValues(t, expectedComponent, component)
							},
							func() {
								relation := <-relationChannel
								expectedRelation := &topology.Relation{
									ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-1->urn:cluster:/kubernetes:test-cluster-name",
									Type:       topology.Type{Name: "belongs_to"},
									SourceID:   "urn:kubernetes:/test-cluster-name:node/test-node-1",
									TargetID:   "urn:cluster:/kubernetes:test-cluster-name",
									Data:       map[string]interface{}{},
								}
								assert.EqualValues(t, expectedRelation, relation)
							},
							func() {
								nodeIdentifier := <-nodeIdentifierCorrelationChannel
								expectedNodeIdentifier := &NodeIdentifierCorrelation{
									NodeName:       "test-node-1",
									NodeIdentifier: "test-node-1-test-cluster-name",
									NodeExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-1",
								}
								assert.EqualValues(t, expectedNodeIdentifier, nodeIdentifier)
							},
						},
					},
					{
						testCase: "Test Node 2 - NodeInternalIP + NodeExternalIP + Kind + Generate Name",
						assertions: []func(){
							func() {
								component := <-componentChannel
								expectedComponent :=
									chooseBySourcePropertiesFeature(
										sourcePropertiesEnabled,
										kubernetesStatusEnabled,
										&topology.Component{
											ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-2",
											Type:       topology.Type{Name: "node"},
											Data: topology.Data{
												"name":              "test-node-2",
												"kind":              "Node",
												"creationTimestamp": creationTime,
												"tags": map[string]string{
													"test":           "label",
													"cluster-name":   "test-cluster-name",
													"cluster-type":   "kubernetes",
													"component-type": "kubernetes-node",
													"namespace":      "test-namespace",
												},
												"uid":        types.UID("test-node-2"),
												"instanceId": "test-node-2-test-cluster-name",
												"sts_host":   "test-node-2-test-cluster-name",
												"status": NodeStatus{
													Phase: coreV1.NodeRunning,
													NodeInfo: coreV1.NodeSystemInfo{
														MachineID:     "test-machine-id-2",
														KernelVersion: "4.19.0",
														Architecture:  "x86_64",
													},
													KubeletEndpoint: coreV1.DaemonEndpoint{Port: 5000},
												},
												"identifiers": []string{
													"urn:ip:/test-cluster-name:test-node-2:10.20.01.01",
													"urn:ip:/test-cluster-name:10.20.01.02",
													"urn:host:/test-node-2-test-cluster-name",
												},
												"generateName": "some-specified-generation",
											},
										},
										&topology.Component{
											ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-2",
											Type:       topology.Type{Name: "node"},
											Data: topology.Data{
												"name": "test-node-2",
												"tags": map[string]string{
													"test":           "label",
													"cluster-name":   "test-cluster-name",
													"cluster-type":   "kubernetes",
													"component-type": "kubernetes-node",
													"namespace":      "test-namespace",
												},
												"instanceId": "test-node-2-test-cluster-name",
												"sts_host":   "test-node-2-test-cluster-name",
												"identifiers": []string{
													"urn:ip:/test-cluster-name:test-node-2:10.20.01.01",
													"urn:ip:/test-cluster-name:10.20.01.02",
													"urn:host:/test-node-2-test-cluster-name",
												},
											},
											SourceProperties: map[string]interface{}{
												"apiVersion": "v1",
												"kind":       "Node",
												"metadata": map[string]interface{}{

													"creationTimestamp": creationTimeFormatted,
													"labels":            map[string]interface{}{"test": "label"},
													"name":              "test-node-2",
													"namespace":         "test-namespace",
													"uid":               "test-node-2",
													"generateName":      "some-specified-generation",
												},
												"spec": map[string]interface{}{},
												"status": map[string]interface{}{
													"phase": "Running",
													"nodeInfo": map[string]interface{}{
														"machineID":               "test-machine-id-2",
														"bootID":                  "",
														"containerRuntimeVersion": "",
														"kernelVersion":           "4.19.0",
														"architecture":            "x86_64",
														"kubeProxyVersion":        "",
														"kubeletVersion":          "",
														"operatingSystem":         "",
														"osImage":                 "",
														"systemUUID":              "",
													},
													"daemonEndpoints": map[string]interface{}{
														"kubeletEndpoint": map[string]interface{}{
															"Port": float64(5000),
														},
													},
												},
											},
										},
										&topology.Component{
											ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-2",
											Type:       topology.Type{Name: "node"},
											Data: topology.Data{
												"name": "test-node-2",
												"tags": map[string]string{
													"test":           "label",
													"cluster-name":   "test-cluster-name",
													"cluster-type":   "kubernetes",
													"component-type": "kubernetes-node",
													"namespace":      "test-namespace",
												},
												"instanceId": "test-node-2-test-cluster-name",
												"sts_host":   "test-node-2-test-cluster-name",
												"identifiers": []string{
													"urn:ip:/test-cluster-name:test-node-2:10.20.01.01",
													"urn:ip:/test-cluster-name:10.20.01.02",
													"urn:host:/test-node-2-test-cluster-name",
												},
											},
											SourceProperties: map[string]interface{}{
												"apiVersion": "v1",
												"kind":       "Node",
												"metadata": map[string]interface{}{
													"creationTimestamp": creationTimeFormatted,
													"labels":            map[string]interface{}{"test": "label"},
													"name":              "test-node-2",
													"namespace":         "test-namespace",
													"uid":               "test-node-2",
													"generateName":      "some-specified-generation",
													"resourceVersion":   "123",
												},
												"spec": map[string]interface{}{},
												"status": map[string]interface{}{
													"addresses": []interface{}{
														map[string]interface{}{
															"address": "10.20.01.01",
															"type":    "InternalIP",
														},
														map[string]interface{}{
															"address": "10.20.01.02",
															"type":    "ExternalIP",
														},
													},
													"phase": "Running",
													"nodeInfo": map[string]interface{}{
														"machineID":               "test-machine-id-2",
														"bootID":                  "",
														"containerRuntimeVersion": "",
														"kernelVersion":           "4.19.0",
														"architecture":            "x86_64",
														"kubeProxyVersion":        "",
														"kubeletVersion":          "",
														"operatingSystem":         "",
														"osImage":                 "",
														"systemUUID":              "",
													},
													"daemonEndpoints": map[string]interface{}{
														"kubeletEndpoint": map[string]interface{}{
															"Port": float64(5000),
														},
													},
												},
											},
										},
									)
								assert.EqualValues(t, expectedComponent, component)
							},
							func() {
								relation := <-relationChannel
								expectedRelation := &topology.Relation{
									ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-2->urn:cluster:/kubernetes:test-cluster-name",
									Type:       topology.Type{Name: "belongs_to"},
									SourceID:   "urn:kubernetes:/test-cluster-name:node/test-node-2",
									TargetID:   "urn:cluster:/kubernetes:test-cluster-name",
									Data:       map[string]interface{}{},
								}
								assert.EqualValues(t, expectedRelation, relation)
							},
							func() {
								nodeIdentifier := <-nodeIdentifierCorrelationChannel
								expectedNodeIdentifier := &NodeIdentifierCorrelation{
									NodeName:       "test-node-2",
									NodeIdentifier: "test-node-2-test-cluster-name",
									NodeExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-2",
								}
								assert.EqualValues(t, expectedNodeIdentifier, nodeIdentifier)
							},
						},
					},
					{
						testCase: "Test Node 3 - Complete",
						assertions: []func(){
							func() {
								component := <-componentChannel
								expectedComponent :=
									chooseBySourcePropertiesFeature(
										sourcePropertiesEnabled,
										kubernetesStatusEnabled,
										&topology.Component{
											ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-3",
											Type:       topology.Type{Name: "node"},
											Data: topology.Data{
												"name":              "test-node-3",
												"kind":              "Node",
												"creationTimestamp": creationTime,
												"tags": map[string]string{
													"test":           "label",
													"cluster-name":   "test-cluster-name",
													"cluster-type":   "kubernetes",
													"component-type": "kubernetes-node",
													"namespace":      "test-namespace",
												},
												"uid": types.UID("test-node-3"),
												"status": NodeStatus{
													Phase: coreV1.NodeRunning,
													NodeInfo: coreV1.NodeSystemInfo{
														MachineID:     "test-machine-id-3",
														KernelVersion: "4.19.0",
														Architecture:  "x86_64",
													},
													KubeletEndpoint: coreV1.DaemonEndpoint{Port: 5000},
												},
												"identifiers": []string{
													"urn:ip:/test-cluster-name:test-node-3:10.20.01.01",
													"urn:ip:/test-cluster-name:10.20.01.02",
													"urn:host:/test-cluster-name:cluster.internal.dns.test-node-3",
													"urn:host:/my-organization.test-node-3",
													"urn:host:/test-node-3-test-cluster-name",
												},
												"generateName": "some-specified-generation",
												"instanceId":   "test-node-3-test-cluster-name",
												"sts_host":     "test-node-3-test-cluster-name",
											},
										},
										&topology.Component{
											ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-3",
											Type:       topology.Type{Name: "node"},
											Data: topology.Data{
												"name": "test-node-3",
												"tags": map[string]string{
													"test":           "label",
													"cluster-name":   "test-cluster-name",
													"cluster-type":   "kubernetes",
													"component-type": "kubernetes-node",
													"namespace":      "test-namespace",
												},
												"identifiers": []string{
													"urn:ip:/test-cluster-name:test-node-3:10.20.01.01",
													"urn:ip:/test-cluster-name:10.20.01.02",
													"urn:host:/test-cluster-name:cluster.internal.dns.test-node-3",
													"urn:host:/my-organization.test-node-3",
													"urn:host:/test-node-3-test-cluster-name",
												},
												"instanceId": "test-node-3-test-cluster-name",
												"sts_host":   "test-node-3-test-cluster-name",
											},
											SourceProperties: map[string]interface{}{
												"apiVersion": "v1",
												"kind":       "Node",
												"metadata": map[string]interface{}{
													"creationTimestamp": creationTimeFormatted,
													"labels":            map[string]interface{}{"test": "label"},
													"name":              "test-node-3",
													"namespace":         "test-namespace",
													"uid":               "test-node-3",
													"generateName":      "some-specified-generation",
												},
												"spec": map[string]interface{}{
													"providerID": "aws:///us-east-1b/i-024b28584ed2e6321",
												},
												"status": map[string]interface{}{
													"phase": "Running",
													"nodeInfo": map[string]interface{}{
														"machineID":               "test-machine-id-3",
														"bootID":                  "",
														"containerRuntimeVersion": "",
														"kernelVersion":           "4.19.0",
														"architecture":            "x86_64",
														"kubeProxyVersion":        "",
														"kubeletVersion":          "",
														"operatingSystem":         "",
														"osImage":                 "",
														"systemUUID":              "",
													},
													"daemonEndpoints": map[string]interface{}{
														"kubeletEndpoint": map[string]interface{}{
															"Port": float64(5000),
														},
													},
												},
											},
										},
										&topology.Component{
											ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-3",
											Type:       topology.Type{Name: "node"},
											Data: topology.Data{
												"name": "test-node-3",
												"tags": map[string]string{
													"test":           "label",
													"cluster-name":   "test-cluster-name",
													"cluster-type":   "kubernetes",
													"component-type": "kubernetes-node",
													"namespace":      "test-namespace",
												},
												"identifiers": []string{
													"urn:ip:/test-cluster-name:test-node-3:10.20.01.01",
													"urn:ip:/test-cluster-name:10.20.01.02",
													"urn:host:/test-cluster-name:cluster.internal.dns.test-node-3",
													"urn:host:/my-organization.test-node-3",
													"urn:host:/test-node-3-test-cluster-name",
												},
												"instanceId": "test-node-3-test-cluster-name",
												"sts_host":   "test-node-3-test-cluster-name",
											},
											SourceProperties: map[string]interface{}{
												"apiVersion": "v1",
												"kind":       "Node",
												"metadata": map[string]interface{}{
													"creationTimestamp": creationTimeFormatted,
													"labels":            map[string]interface{}{"test": "label"},
													"name":              "test-node-3",
													"namespace":         "test-namespace",
													"uid":               "test-node-3",
													"generateName":      "some-specified-generation",
													"resourceVersion":   "123",
												},
												"spec": map[string]interface{}{
													"providerID": "aws:///us-east-1b/i-024b28584ed2e6321",
												},
												"status": map[string]interface{}{
													"addresses": []interface{}{
														map[string]interface{}{
															"address": "10.20.01.01",
															"type":    "InternalIP",
														},
														map[string]interface{}{
															"address": "10.20.01.02",
															"type":    "ExternalIP",
														},
														map[string]interface{}{
															"address": "cluster.internal.dns.test-node-3",
															"type":    "InternalDNS",
														},
														map[string]interface{}{
															"address": "my-organization.test-node-3",
															"type":    "ExternalDNS",
														},
													},
													"phase": "Running",
													"nodeInfo": map[string]interface{}{
														"machineID":               "test-machine-id-3",
														"bootID":                  "",
														"containerRuntimeVersion": "",
														"kernelVersion":           "4.19.0",
														"architecture":            "x86_64",
														"kubeProxyVersion":        "",
														"kubeletVersion":          "",
														"operatingSystem":         "",
														"osImage":                 "",
														"systemUUID":              "",
													},
													"daemonEndpoints": map[string]interface{}{
														"kubeletEndpoint": map[string]interface{}{
															"Port": float64(5000),
														},
													},
												},
											},
										},
									)
								assert.EqualValues(t, expectedComponent, component)
							},
							func() {
								relation := <-relationChannel
								expectedRelation := &topology.Relation{
									ExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-3->urn:cluster:/kubernetes:test-cluster-name",
									Type:       topology.Type{Name: "belongs_to"},
									SourceID:   "urn:kubernetes:/test-cluster-name:node/test-node-3",
									TargetID:   "urn:cluster:/kubernetes:test-cluster-name",
									Data:       map[string]interface{}{},
								}
								assert.EqualValues(t, expectedRelation, relation)
							},
							func() {
								nodeIdentifier := <-nodeIdentifierCorrelationChannel
								expectedNodeIdentifier := &NodeIdentifierCorrelation{
									NodeName:       "test-node-3",
									NodeIdentifier: "test-node-3-test-cluster-name",
									NodeExternalID: "urn:kubernetes:/test-cluster-name:node/test-node-3",
								}
								assert.EqualValues(t, expectedNodeIdentifier, nodeIdentifier)
							},
						},
					},
				} {
					t.Run(testCaseName(tc.testCase, sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {
						for _, assertion := range tc.assertions {
							assertion()
						}
					})
				}
			})()
		}
	}
}

func CreateBaseNode(id int) coreV1.Node {
	return coreV1.Node{
		TypeMeta: v1.TypeMeta{
			Kind: "Node",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:              fmt.Sprintf("test-node-%d", id),
			CreationTimestamp: creationTime,
			Namespace:         "test-namespace",
			Labels: map[string]string{
				"test": "label",
			},
			UID:             types.UID(fmt.Sprintf("test-node-%d", id)),
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
		Status: coreV1.NodeStatus{
			Phase: coreV1.NodeRunning,
			NodeInfo: coreV1.NodeSystemInfo{
				MachineID:     fmt.Sprintf("test-machine-id-%d", id),
				KernelVersion: "4.19.0",
				Architecture:  "x86_64",
			},
			DaemonEndpoints: coreV1.NodeDaemonEndpoints{KubeletEndpoint: coreV1.DaemonEndpoint{Port: 5000}},
		},
	}
}

type MockNodeAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (m MockNodeAPICollectorClient) GetNodes() ([]coreV1.Node, error) {
	nodes := make([]coreV1.Node, 0)

	node1 := CreateBaseNode(1)
	node1.Status.Addresses = []coreV1.NodeAddress{
		{Type: coreV1.NodeInternalIP, Address: "10.20.01.01"},
	}
	nodes = append(nodes, node1)

	node2 := CreateBaseNode(2)
	node2.ObjectMeta.GenerateName = "some-specified-generation"
	node2.Status.Addresses = []coreV1.NodeAddress{
		{Type: coreV1.NodeInternalIP, Address: "10.20.01.01"},
		{Type: coreV1.NodeExternalIP, Address: "10.20.01.02"},
	}
	nodes = append(nodes, node2)

	node3 := CreateBaseNode(3)
	node3.ObjectMeta.GenerateName = "some-specified-generation"
	node3.Spec.ProviderID = "aws:///us-east-1b/i-024b28584ed2e6321"
	node3.Status.Addresses = []coreV1.NodeAddress{
		{Type: coreV1.NodeInternalIP, Address: "10.20.01.01"},
		{Type: coreV1.NodeExternalIP, Address: "10.20.01.02"},
		{Type: coreV1.NodeInternalDNS, Address: "cluster.internal.dns.test-node-3"},
		{Type: coreV1.NodeExternalDNS, Address: "my-organization.test-node-3"},
	}
	nodes = append(nodes, node3)

	return nodes, nil
}
