// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClusterCollector(t *testing.T) {
	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	cc := NewClusterCollector(NewTestCommonClusterCollector(MockClusterAPICollectorClient{}, componentChannel, relationChannel, true, true))
	expectedCollectorName := "Cluster Collector"
	RunCollectorTest(t, cc, expectedCollectorName)

	for _, tc := range []struct {
		testCase string
		expected *topology.Component
	}{
		{
			testCase: "Test Cluster component creation",
			expected: &topology.Component{
				ExternalID: "urn:cluster:/kubernetes:test-cluster-name",
				Type:       topology.Type{Name: "cluster"},
				Data: topology.Data{
					"name": "test-cluster-name",
					"tags": map[string]string{
						"cluster-name":   "test-cluster-name",
						"cluster-type":   "kubernetes",
						"component-type": "kubernetes-cluster",
					}},
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			clusterComponent := <-componentChannel
			assert.EqualValues(t, tc.expected, clusterComponent)
		})
	}
}

type MockClusterAPICollectorClient struct {
	apiserver.APICollectorClient
}
