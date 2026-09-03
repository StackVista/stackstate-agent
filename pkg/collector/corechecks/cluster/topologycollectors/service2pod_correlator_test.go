//go:build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestService2PodCorrelator(t *testing.T) {
	clusterName := "test-cluster-name"
	namespace := "default"
	// namespace2 := "monitoring"

	svcName := "kubernetes"

	pod1Name := "kube-apiserver-1"
	host1Labels := map[string]string{"label1": "match1", "label2": "match2"}

	pod2Name := "kube-apiserver-2"
	host2Labels := map[string]string{"label1": "match1", "label2": "match2", "label3": "match3"}

	pod3Name := "unrelated-3"
	host3Labels := map[string]string{"label1": "match1", "label2": "matchFoo"}

	pod4Name := "unrelated-4"
	host4Labels := map[string]string{"label1": "match1", "labelFoo": "match2"}

	selector := map[string]string{"label1": "match1", "label2": "match2"}

	someTimestamp := metav1.NewTime(time.Now())

	pod1 := podWithLabels(namespace, pod1Name, someTimestamp, host1Labels)
	pod2 := podWithLabels(namespace, pod2Name, someTimestamp, host2Labels)
	pod3 := podWithLabels(namespace, pod3Name, someTimestamp, host3Labels)
	pod4 := podWithLabels(namespace, pod4Name, someTimestamp, host4Labels)
	service := serviceWithSelector(namespace, svcName, someTimestamp, selector)

	components, relations := executeCorrelation(t, []coreV1.Pod{
		pod1, pod2, pod3, pod4,
	}, []coreV1.Service{
		service,
	})

	expectedPod1ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod1Name)
	expectedPod2ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod2Name)
	expectedPod3ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod3Name)
	expectedPod4ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod4Name)
	expectedSvcID := fmt.Sprintf("urn:kubernetes:/%s:%s:service/%s", clusterName, namespace, svcName)
	expectedNamespaceID := fmt.Sprintf("urn:kubernetes:/%s:namespace/%s", clusterName, namespace)
	expectedComponents := []*topology.Component{
		expectedPodComponentWithLabels(pod1, host1Labels),
		expectedPodComponentWithLabels(pod2, host2Labels),
		expectedPodComponentWithLabels(pod3, host3Labels),
		expectedPodComponentWithLabels(pod4, host4Labels),
		expectedServiceComponent(service),
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
		// no relation to pod3 - the labels do not match
		// no relation to pod4 - the labels do not match
	}

	assert.EqualValues(t, expectedComponents, components)
	assert.EqualValues(t, expectedRelations, relations)
	return
}

func expectedServiceComponent(svc coreV1.Service) *topology.Component {
	clusterName := "test-cluster-name"
	externalID := fmt.Sprintf("urn:kubernetes:/%s:%s:service/%s", clusterName, svc.Namespace, svc.Name)
	component := &topology.Component{
		ExternalID: externalID,
		Type:       topology.Type{Name: "service"},
		Data: topology.Data{
			"name": svc.Name,
			"tags": map[string]string{
				"cluster-name":   clusterName,
				"cluster-type":   "kubernetes",
				"component-type": "kubernetes-service",
				"namespace":      svc.Namespace,
				"service-type":   string(svc.Spec.Type),
			},
			"identifiers": []string{
				fmt.Sprintf("urn:service:/%s:%s:%s", clusterName, svc.Namespace, svc.Name),
			},
		},
	}
	component.SourceProperties = makeSourcePropertiesFullDetails(&svc)
	return component
}

func expectedPodComponentWithLabels(pod coreV1.Pod, labels map[string]string) *topology.Component {
	clusterName := "test-cluster-name"
	externalID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, pod.Namespace, pod.Name)
	tags := map[string]string{
		"cluster-name":   clusterName,
		"cluster-type":   "kubernetes",
		"component-type": "kubernetes-pod",
		"namespace":      pod.Namespace,
	}

	for k, v := range labels {
		tags[k] = v
	}

	component := &topology.Component{
		ExternalID: externalID,
		Type:       topology.Type{Name: "pod"},
		Data: topology.Data{
			"name":        pod.Name,
			"tags":        tags,
			"identifiers": []string{},
			"status": map[string]interface{}{
				"phase": string(pod.Status.Phase),
			},
		},
	}
	component.SourceProperties = makeSourcePropertiesFullDetails(&pod)
	return component
}

func serviceWithSelector(namespace string, name string, timestamp metav1.Time, selector map[string]string) coreV1.Service {
	return coreV1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         namespace,
			Name:              name,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		TypeMeta: metav1.TypeMeta{Kind: "Service"},
		Spec: coreV1.ServiceSpec{
			Selector: selector,
		},
	}

}

func podWithLabels(namespace string, name string, timestamp metav1.Time, labels map[string]string) coreV1.Pod {
	return coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
			Labels:            labels,
		},
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
		Spec: coreV1.PodSpec{
			HostNetwork: true,
			Containers:  []coreV1.Container{{}},
		},
		Status: coreV1.PodStatus{
			Phase:     coreV1.PodRunning,
			StartTime: &timestamp,
		},
	}
}

func executeCorrelation(
	t *testing.T,
	pods []coreV1.Pod,
	services []coreV1.Service,
) ([]*topology.Component, []*topology.Relation) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	clusterAPIClient := MockS2PCorrelatorAPIClient{
		services: services, pods: pods,
	}

	podCorrChannel := make(chan *PodLabelCorrelation)
	serviceCorrChannel := make(chan *ServiceSelectorCorrelation)
	containerCorrChannel := make(chan *ContainerCorrelation)
	volumeCorrChannel := make(chan *VolumeCorrelation)
	commonClusterCollector := NewTestCommonClusterCollector(clusterAPIClient, componentChannel, relationChannel)
	commonClusterCollector.SetUseRelationCache(false)
	correlator := NewService2PodCorrelator(
		podCorrChannel,
		serviceCorrChannel,
		NewClusterTopologyCorrelator(commonClusterCollector),
	)
	podCollector := NewPodCollector(
		containerCorrChannel, volumeCorrChannel,
		podCorrChannel,
		commonClusterCollector,
		true,
	)
	svcCollector := NewServiceCollector(
		serviceCorrChannel,
		commonClusterCollector,
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
	pods     []coreV1.Pod
	services []coreV1.Service
	apiserver.APICollectorClient
}

func (m MockS2PCorrelatorAPIClient) GetServices() ([]coreV1.Service, error) {
	return m.services, nil
}

func (m MockS2PCorrelatorAPIClient) GetPods() ([]coreV1.Pod, error) {
	return m.pods, nil
}
