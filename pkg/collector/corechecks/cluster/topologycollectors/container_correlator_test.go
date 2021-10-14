// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

var startedAtTime = metav1.NewTime(time.Now())

func TestContainerCollector(t *testing.T) {
	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	nodeIdentifierCorrelationChannel := make(chan *NodeIdentifierCorrelation)
	containerCorrelationChannel := make(chan *ContainerCorrelation)

	cc := NewContainerCorrelator(componentChannel, relationChannel, nodeIdentifierCorrelationChannel,
		containerCorrelationChannel, NewTestCommonClusterCorrelator(MockContainerAPICollectorClient{}))
	expectedCollectorName := "Container Correlator"

	populateData(nodeIdentifierCorrelationChannel, containerCorrelationChannel)

	RunCorrelatorTest(t, cc, expectedCollectorName)

	for _, tc := range []struct {
		testCase   string
		assertions []func()
	}{
		{
			testCase: "Test Container 1",
			assertions: []func(){
				func() {
					component := <-componentChannel
					expectedComponent := &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace-1:pod/Pod-Name-1:container/container-1",
						Type:       topology.Type{Name: "container"},
						Data: topology.Data{
							"docker":       map[string]interface{}{"containerId": "containerID-1", "image": "image-1"},
							"identifiers":  []string{"urn:container:/nodeID-1:containerID-1"},
							"name":         "container-1",
							"pod":          "Pod-Name-1",
							"podIP":        "10.0.1.1",
							"podPhase":     "Running",
							"restartCount": int32(1),
							"tags":         map[string]string{"cluster-name": "test-cluster-name", "namespace": "namespace-1"},
						},
					}
					assert.EqualValues(t, expectedComponent, component)
				},
				func() {
					relation := <-relationChannel
					expectedRelation := &topology.Relation{
						ExternalID: "Pod-ExternalID-1->urn:kubernetes:/test-cluster-name:namespace-1:pod/Pod-Name-1:container/container-1",
						SourceID:   "Pod-ExternalID-1",
						TargetID:   "urn:kubernetes:/test-cluster-name:namespace-1:pod/Pod-Name-1:container/container-1",
						Type:       topology.Type{Name: "encloses"},
						Data:       topology.Data{},
					}
					assert.EqualValues(t, expectedRelation, relation)
				},
				func() {
					relation := <-relationChannel
					expectedRelation := &topology.Relation{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace-1:pod/Pod-Name-1:container/container-1->urn:kubernetes:/cluster:node/nodeID-1",
						SourceID:   "urn:kubernetes:/test-cluster-name:namespace-1:pod/Pod-Name-1:container/container-1",
						TargetID:   "urn:kubernetes:/cluster:node/nodeID-1",
						Type:       topology.Type{Name: "runs_on"},
						Data:       topology.Data{},
					}
					assert.EqualValues(t, expectedRelation, relation)
				},
			},
		},
		{
			testCase: "Test Container 2",
			assertions: []func(){
				func() {
					component := <-componentChannel
					expectedComponent := &topology.Component{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace-2:pod/Pod-Name-2:container/container-2",
						Type:       topology.Type{Name: "container"},
						Data: topology.Data{
							"containerPort": int32(1234),
							"docker":        map[string]interface{}{"containerId": "containerID-2", "image": "image-2"},
							"hostPort":      int32(8080),
							"identifiers":   []string{"urn:container:/nodeID-2:containerID-2"},
							"name":          "container-2",
							"pod":           "Pod-Name-2",
							"podIP":         "10.0.1.2",
							"podPhase":      "Running",
							"restartCount":  int32(2),
							"startTime":     startedAtTime,
							"tags":          map[string]string{"cluster-name": "test-cluster-name", "namespace": "namespace-2"},
						},
					}
					assert.EqualValues(t, expectedComponent, component)
				},
				func() {
					relation := <-relationChannel
					expectedRelation := &topology.Relation{
						ExternalID: "Pod-ExternalID-2->urn:kubernetes:/test-cluster-name:namespace-2:pod/Pod-Name-2:container/container-2",
						SourceID:   "Pod-ExternalID-2",
						TargetID:   "urn:kubernetes:/test-cluster-name:namespace-2:pod/Pod-Name-2:container/container-2",
						Type:       topology.Type{Name: "encloses"},
						Data:       topology.Data{},
					}
					assert.EqualValues(t, expectedRelation, relation)
				},
				func() {
					relation := <-relationChannel
					expectedRelation := &topology.Relation{
						ExternalID: "urn:kubernetes:/test-cluster-name:namespace-2:pod/Pod-Name-2:container/container-2->urn:kubernetes:/cluster:node/nodeID-2",
						SourceID:   "urn:kubernetes:/test-cluster-name:namespace-2:pod/Pod-Name-2:container/container-2",
						TargetID:   "urn:kubernetes:/cluster:node/nodeID-2",
						Type:       topology.Type{Name: "runs_on"},
						Data:       topology.Data{},
					}
					assert.EqualValues(t, expectedRelation, relation)
				},
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			for _, assertion := range tc.assertions {
				assertion()
			}
		})
	}
}

func populateData(nodeIdentifierCorrelationChannel chan *NodeIdentifierCorrelation, containerCorrelationChannel chan *ContainerCorrelation) {
	go func() {
		fmt.Println("start nodeIdentifierCorrelationChannel")
		defer close(nodeIdentifierCorrelationChannel)
		for i := 1; i <= 2; i++ {
			nodeIdentifierCorrelationChannel <- CreateNodeIdentifierCorrelation(i)
		}
		fmt.Println("end nodeIdentifierCorrelationChannel")
	}()
	go func() {
		fmt.Println("start containerCorrelationChannel")
		defer close(containerCorrelationChannel)
		containerCorrelation1 := CreateContainerCorrelation(1, false, false)
		containerCorrelationChannel <- containerCorrelation1
		containerCorrelation2 := CreateContainerCorrelation(2, true, true)
		containerCorrelationChannel <- containerCorrelation2
		fmt.Println("end containerCorrelationChannel")
	}()
}

func CreateContainerCorrelation(id int, isRunning bool, hasPort bool) *ContainerCorrelation {
	var running *v1.ContainerStateRunning = nil
	if isRunning {
		running = &v1.ContainerStateRunning{
			StartedAt: startedAtTime,
		}
	}

	ports := []v1.ContainerPort{}
	if hasPort {
		ports = []v1.ContainerPort{
			{"port1", 8080, 1234, v1.ProtocolTCP, "10.0.0.1"},
		}
	}

	return &ContainerCorrelation{
		Pod: ContainerPod{
			ExternalID: fmt.Sprintf("Pod-ExternalID-%d", id),
			Name:       fmt.Sprintf("Pod-Name-%d", id),
			Labels: map[string]string{
				fmt.Sprintf("tag-%d", id): fmt.Sprintf("value-%d", id),
			},
			PodIP:     fmt.Sprintf("10.0.1.%d", id),
			Namespace: fmt.Sprintf("namespace-%d", id),
			NodeName:  fmt.Sprintf("node-%d", id),
			Phase:     "Running",
		},
		Containers: []v1.Container{
			{
				Name:  fmt.Sprintf("container-%d", id),
				Image: fmt.Sprintf("image-%d", id),
				Ports: ports,
			},
		},
		ContainerStatuses: []v1.ContainerStatus{
			{
				Name:         fmt.Sprintf("container-%d", id),
				ContainerID:  fmt.Sprintf("containerID-%d", id),
				Image:        fmt.Sprintf("image-%d", id),
				RestartCount: int32(id),
				State: v1.ContainerState{
					Waiting:    nil,
					Running:    running,
					Terminated: nil,
				},
			},
		},
	}
}

func CreateNodeIdentifierCorrelation(id int) *NodeIdentifierCorrelation {
	return &NodeIdentifierCorrelation{
		NodeName:       fmt.Sprintf("node-%d", id),
		NodeIdentifier: fmt.Sprintf("nodeID-%d", id),
		NodeExternalID: fmt.Sprintf("urn:kubernetes:/cluster:node/nodeID-%d", id),
	}
}

type MockContainerAPICollectorClient struct {
	apiserver.APICollectorClient
}
