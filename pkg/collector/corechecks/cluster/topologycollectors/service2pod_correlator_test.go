//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

func TestService2PodCorrelator(t *testing.T) {
	clusterName := "test-cluster-name"
	namespace := "default"
	svcName := "kubernetes"
	hostPort := int32(443)
	pod1Name := "kube-apiserver-1"
	host1IP := "10.77.0.1"
	pod2Name := "kube-apiserver-2"
	host2IP := "10.77.0.2"
	pod3Name := "unrelated-3"
	host3IP := "10.77.0.3"
	pod4Name := "unrelated-4"
	host4IP := "10.77.0.4"
	hostUnrelatedPort := int32(888)
	someTimestamp := metav1.NewTime(time.Now())

	pod1 := podWithHostPortExposed(namespace, pod1Name, someTimestamp, host1IP, hostPort)
	pod2 := podWithHostPortExposed(namespace, pod2Name, someTimestamp, host2IP, hostPort)
	pod3 := podWithHostPortExposed(namespace, pod3Name, someTimestamp, host3IP, hostUnrelatedPort)
	pod4 := podWithHostPortExposed(namespace, pod4Name, someTimestamp, host4IP, hostUnrelatedPort)
	service := serviceWithSinglePort(namespace, svcName, someTimestamp, hostPort)
	endpoint := endpointsForASinglePort(namespace, svcName, someTimestamp, hostPort, host1IP, host2IP, host4IP)

	components, relations := executeCorrelation(t, []coreV1.Pod{
		pod1, pod2, pod3, pod4,
	}, []coreV1.Service{
		service,
	}, []coreV1.Endpoints{
		endpoint,
	})

	expectedPod1ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod1Name)
	expectedPod2ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod2Name)
	expectedPod3ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod3Name)
	expectedPod4ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod4Name)
	expectedSvcID := fmt.Sprintf("urn:kubernetes:/%s:%s:service/%s", clusterName, namespace, svcName)
	expectedNamespaceID := fmt.Sprintf("urn:kubernetes:/%s:namespace/%s", clusterName, namespace)
	expectedComponents := []*topology.Component{
		podComponentWithHostPortExposed(clusterName, namespace, pod1Name, expectedPod1ID, host1IP, someTimestamp, pod1.Status),
		podComponentWithHostPortExposed(clusterName, namespace, pod2Name, expectedPod2ID, host2IP, someTimestamp, pod2.Status),
		podComponentWithHostPortExposed(clusterName, namespace, pod3Name, expectedPod3ID, host3IP, someTimestamp, pod3.Status),
		podComponentWithHostPortExposed(clusterName, namespace, pod4Name, expectedPod4ID, host4IP, someTimestamp, pod4.Status),
		serviceComponentWithSinglePort(clusterName, namespace, expectedSvcID, svcName, someTimestamp),
	}
	expectedRelations := []*topology.Relation{
		simpleRelation(expectedNamespaceID, expectedPod1ID, "encloses"),
		simpleRelation(expectedNamespaceID, expectedPod2ID, "encloses"),
		simpleRelation(expectedNamespaceID, expectedPod3ID, "encloses"),
		simpleRelation(expectedNamespaceID, expectedPod4ID, "encloses"),
		simpleRelation(expectedNamespaceID, expectedSvcID, "encloses"),
		// these are the most important in our test case
		simpleRelation(expectedSvcID, expectedPod1ID, "exposes"),
		simpleRelation(expectedSvcID, expectedPod2ID, "exposes"),
		// no relation to pod3 - it's not mentioned in the endpoints and has different port
		// no relation to pod4 - it's mentioned in the endpoints, but has different port
	}

	assert.EqualValues(t, expectedComponents, components)
	assert.EqualValues(t, expectedRelations, relations)
	return
}

func simpleRelation(sourceID string, targetID string, typ string) *topology.Relation {
	return &topology.Relation{
		ExternalID: fmt.Sprintf("%s->%s", sourceID, targetID),
		SourceID:   sourceID,
		TargetID:   targetID,
		Type:       topology.Type{Name: typ},
		Data:       map[string]interface{}{},
	}
}

func serviceComponentWithSinglePort(clusterName string, namespace string, externalID string, name string, timestamp metav1.Time) *topology.Component {
	return &topology.Component{
		ExternalID: externalID,
		Type:       topology.Type{Name: "service"},
		Data: topology.Data{
			"name": name,
			"tags": map[string]string{
				"cluster-name": clusterName,
				"namespace":    namespace,
				"service-type": "",
			},
			"identifiers": []string{
				fmt.Sprintf("urn:service:/%s:%s:%s", clusterName, namespace, name),
			},
			"creationTimestamp": timestamp,
			"uid":               types.UID(""),
		},
	}
}

func podComponentWithHostPortExposed(clusterName string, namespace string, name string, externalID string, hostIP string, timestamp metav1.Time, status coreV1.PodStatus) *topology.Component {
	return &topology.Component{
		ExternalID: externalID,
		Type:       topology.Type{Name: "pod"},
		Data: topology.Data{
			"name": name,
			"tags": map[string]string{
				"cluster-name": clusterName,
				"namespace":    namespace,
			},
			"identifiers": []string{
				fmt.Sprintf("urn:ip:/%s:%s:%s:%s", clusterName, namespace, name, hostIP),
			},
			"creationTimestamp": timestamp,
			"uid":               types.UID(""),
			"restartPolicy":     coreV1.RestartPolicy(""),
			"status":            status,
		},
	}
}

func endpointsForASinglePort(namespace string, name string, timestamp metav1.Time, port int32, IPs ...string) coreV1.Endpoints {
	addresses := make([]coreV1.EndpointAddress, len(IPs))
	for _, ip := range IPs {
		addresses = append(addresses, coreV1.EndpointAddress{
			IP: ip,
		})
	}
	return coreV1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         namespace,
			Name:              name,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		Subsets: []coreV1.EndpointSubset{
			{
				Addresses: addresses,
				Ports: []coreV1.EndpointPort{
					{
						Port: port,
					},
				},
			},
		},
	}

}

func serviceWithSinglePort(namespace string, name string, timestamp metav1.Time, port int32) coreV1.Service {
	return coreV1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         namespace,
			Name:              name,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		Spec: coreV1.ServiceSpec{
			Ports: []coreV1.ServicePort{
				{
					Port: port,
				},
			},
		},
	}

}

func podWithHostPortExposed(namespace string, name string, timestamp metav1.Time, hostIP string, hostPort int32) coreV1.Pod {
	return coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		Spec: coreV1.PodSpec{
			HostNetwork: true,
			Containers: []coreV1.Container{
				{
					Ports: []coreV1.ContainerPort{
						{
							HostPort:      hostPort,
							ContainerPort: hostPort,
						},
					},
				},
			},
		},
		Status: coreV1.PodStatus{
			Phase:     coreV1.PodRunning,
			HostIP:    hostIP,
			PodIP:     hostIP,
			StartTime: &timestamp,
		},
	}
}

func executeCorrelation(t *testing.T, pods []coreV1.Pod, services []coreV1.Service, endpoints []coreV1.Endpoints) ([]*topology.Component, []*topology.Relation) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	clusterAPIClient := MockS2PCorrelatorAPIClient{
		services: services, pods: pods, endpoints: endpoints,
	}

	podCorrChannel := make(chan *PodEndpointCorrelation)
	serviceCorrChannel := make(chan *ServiceEndpointCorrelation)
	containerCorrChannel := make(chan *ContainerCorrelation)
	volumeCorrChannel := make(chan *VolumeCorrelation)
	correlator := NewService2PodCorrelator(
		componentChannel, relationChannel,
		podCorrChannel,
		serviceCorrChannel,
		NewTestCommonClusterCorrelator(clusterAPIClient),
	)
	podCollector := NewPodCollector(
		componentChannel, relationChannel,
		containerCorrChannel, volumeCorrChannel,
		podCorrChannel,
		NewTestCommonClusterCollector(clusterAPIClient, false),
	)
	svcCollector := NewServiceCollector(
		componentChannel, relationChannel,
		serviceCorrChannel,
		NewTestCommonClusterCollector(clusterAPIClient, false),
	)

	collectorsFinishChan := make(chan bool)
	correlatorFinishChan := make(chan bool)
	go func() {
		err := podCollector.CollectorFunction()
		assert.NoError(t, err)
		err = svcCollector.CollectorFunction()
		assert.NoError(t, err)
		collectorsFinishChan <- true
	}()

	go func() {
		err := correlator.CorrelateFunction()
		assert.NoError(t, err)
		correlatorFinishChan <- true
	}()

	components := make([]*topology.Component, 0)
	relations := make([]*topology.Relation, 0)

	collectorsFinished := false
	correlatorFinished := false

L:
	for {
		select {
		case c := <-componentChannel:
			components = append(components, c)
		case r := <-relationChannel:
			relations = append(relations, r)
		case <-collectorsFinishChan:
			if correlatorFinished {
				break L
			}
			collectorsFinished = true
		case <-correlatorFinishChan:
			if collectorsFinished {
				break L
			}
			correlatorFinished = true
		}
	}

	return components, relations
}

type MockS2PCorrelatorAPIClient struct {
	pods      []coreV1.Pod
	services  []coreV1.Service
	endpoints []coreV1.Endpoints
	apiserver.APICollectorClient
}

func (m MockS2PCorrelatorAPIClient) GetEndpoints() ([]coreV1.Endpoints, error) {
	return m.endpoints, nil
}

func (m MockS2PCorrelatorAPIClient) GetServices() ([]coreV1.Service, error) {
	return m.services, nil
}

func (m MockS2PCorrelatorAPIClient) GetPods() ([]coreV1.Pod, error) {
	return m.pods, nil
}
