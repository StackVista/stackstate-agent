// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
//go:build kubeapiserver

package topologycollectors

import (
	"errors"
	"testing"
	"time"

	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
)

// failingPodAPICollectorClient fails the test if the Kubernetes API is consulted, proving a
// disabled Pod collector skips the list call rather than only discarding its results.
type failingPodAPICollectorClient struct {
	apiserver.APICollectorClient
}

func (failingPodAPICollectorClient) GetPods() ([]coreV1.Pod, error) {
	return nil, errors.New("GetPods must not be called when pod topology is disabled")
}

func TestPodCollectorDisabledClosesCorrelationChannels(t *testing.T) {
	componentChannel := make(chan *topology.Component)
	relationChannel := make(chan *topology.Relation)
	containerCorrChan := make(chan *ContainerCorrelation)
	volumeCorrChan := make(chan *VolumeCorrelation)
	podCorrChan := make(chan *PodLabelCorrelation)

	submitted := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-componentChannel:
				select {
				case submitted <- struct{}{}:
				default:
				}
			case <-relationChannel:
				select {
				case submitted <- struct{}{}:
				default:
				}
			}
		}
	}()

	common := NewTestCommonClusterCollector(failingPodAPICollectorClient{}, componentChannel, relationChannel)
	common.SetUseRelationCache(false)
	collector := NewPodCollector(containerCorrChan, volumeCorrChan, podCorrChan, common, false)

	done := make(chan error, 1)
	go func() { done <- collector.CollectorFunction() }()

	select {
	case err := <-done:
		// A non-nil error here means GetPods was called despite pods being disabled.
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("disabled Pod collector did not return")
	}

	// The correlators range over these channels, so each must be closed or the check hangs.
	for name, isClosed := range map[string]func() bool{
		"ContainerCorrChan": func() bool { _, ok := <-containerCorrChan; return !ok },
		"VolumeCorrChan":    func() bool { _, ok := <-volumeCorrChan; return !ok },
		"PodCorrChan":       func() bool { _, ok := <-podCorrChan; return !ok },
	} {
		assert.True(t, isClosed(), "%s was not closed by the disabled Pod collector", name)
	}

	select {
	case <-submitted:
		t.Fatal("disabled Pod collector submitted topology")
	default:
	}
}

func TestContainerCorrelatorDisabledDrainsChannel(t *testing.T) {
	componentChannel := make(chan *topology.Component)
	relationChannel := make(chan *topology.Relation)
	nodeIdentifierCorrChan := make(chan *NodeIdentifierCorrelation)
	containerCorrChan := make(chan *ContainerCorrelation)

	submitted := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-componentChannel:
				select {
				case submitted <- struct{}{}:
				default:
				}
			case <-relationChannel:
				select {
				case submitted <- struct{}{}:
				default:
				}
			}
		}
	}()

	common := NewClusterTopologyCorrelator(
		NewTestCommonClusterCollector(MockContainerAPICollectorClient{}, componentChannel, relationChannel),
	)
	common.SetUseRelationCache(false)
	correlator := NewContainerCorrelator(nodeIdentifierCorrChan, containerCorrChan, common, false)

	done := make(chan error, 1)
	go func() { done <- correlator.CorrelateFunction() }()

	// The producing collector must not block on an unbuffered send, so the disabled
	// correlator has to keep draining rather than returning early.
	close(nodeIdentifierCorrChan)
	containerCorrChan <- &ContainerCorrelation{
		Pod:               ContainerPod{},
		Containers:        []coreV1.Container{{Name: "c", Image: "i"}},
		ContainerStatuses: []coreV1.ContainerStatus{{Name: "c", Image: "i"}},
	}
	close(containerCorrChan)

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("disabled Container correlator did not drain and return")
	}

	select {
	case <-submitted:
		t.Fatal("disabled Container correlator submitted topology")
	default:
	}
}
