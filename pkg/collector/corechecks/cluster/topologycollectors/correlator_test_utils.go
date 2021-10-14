// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2019 Datadog, Inc.
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var creationTime v1.Time
var replicas int32
var pathType coreV1.HostPathType
var gcePersistentDisk coreV1.GCEPersistentDiskVolumeSource
var awsElasticBlockStore coreV1.AWSElasticBlockStoreVolumeSource
var hostPath coreV1.HostPathVolumeSource

func NewTestCommonClusterCorrelator(client apiserver.APICollectorClient) ClusterTopologyCorrelator {
	instance := topology.Instance{Type: "kubernetes", URL: "test-cluster-name"}

	clusterTopologyCommon := NewClusterTopologyCommon(instance, client)
	return NewClusterTopologyCorrelator(clusterTopologyCommon)
}

func RunCorrelatorTest(t *testing.T, correlator ClusterTopologyCorrelator, expectedCorrelatorName string) {
	actualCorrelatorName := correlator.GetName()
	assert.Equal(t, expectedCorrelatorName, actualCorrelatorName)

	// Trigger Correlator Function
	go func() {
		log.Debugf("Starting cluster topology correlator: %s\n", correlator.GetName())
		err := correlator.CorrelateFunction()
		// assert no error occurred
		assert.Nil(t, err)
		// mark this correlator as complete
		log.Debugf("Finished cluster topology correlator: %s\n", correlator.GetName())
	}()
}
