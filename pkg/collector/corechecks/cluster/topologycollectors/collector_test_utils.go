// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
//go:build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/version"
	"testing"
)

func NewTestCommonClusterCollector(
	client apiserver.APICollectorClient,
	componentChan chan<- *topology.Component,
	relationChan chan<- *topology.Relation,
	sourcePropertiesEnabled bool,
	kubernetesStatusEnabled bool) ClusterTopologyCollector {
	instance := topology.Instance{Type: "kubernetes", URL: "test-cluster-name"}
	clusterType := Kubernetes

	k8sVersion := version.Info{
		Major: "1",
		Minor: "21",
	}

	clusterTopologyCommon := NewClusterTopologyCommon(instance, clusterType, client, sourcePropertiesEnabled, componentChan, relationChan, &k8sVersion, kubernetesStatusEnabled)
	return NewClusterTopologyCollector(clusterTopologyCommon)
}

func NewTestCommonClusterCollectorWithVersion(client apiserver.APICollectorClient, sourcePropertiesEnabled bool,
	componentChan chan<- *topology.Component,
	relationChan chan<- *topology.Relation,
	k8sVersion *version.Info,
	kubernetesStatusEnabled bool) ClusterTopologyCollector {
	instance := topology.Instance{Type: "kubernetes", URL: "test-cluster-name"}
	clusterType := Kubernetes

	clusterTopologyCommon := NewClusterTopologyCommon(instance, clusterType, client, sourcePropertiesEnabled, componentChan, relationChan, k8sVersion, kubernetesStatusEnabled)
	return NewClusterTopologyCollector(clusterTopologyCommon)
}

func RunCollectorTest(t *testing.T, collector ClusterTopologyCollector, expectedCollectorName string) {
	actualCollectorName := collector.GetName()
	assert.Equal(t, expectedCollectorName, actualCollectorName)

	// Trigger Collector Function
	go func() {
		log.Debugf("Starting cluster topology collector: %s\n", collector.GetName())
		err := collector.CollectorFunction()
		// assert no error occurred
		assert.Nil(t, err)
		// mark this collector as complete
		log.Debugf("Finished cluster topology collector: %s\n", collector.GetName())
	}()
}
