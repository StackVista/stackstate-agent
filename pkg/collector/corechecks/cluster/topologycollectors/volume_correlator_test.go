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
	"testing"
	"time"
)

func TestVolumeCorrelator(t *testing.T) {
	clusterName := "test-cluster-name"
	namespace := "default"
	svcName := "kubernetes"
	pod1Name := "kube-apiserver-1"
	host1IP := "10.77.0.1"
	pod2Name := "kube-apiserver-2"
	host2IP := "10.77.0.2"
	pod3Name := "unrelated-3"
	host3IP := "10.77.0.3"
	pod4Name := "unrelated-4"
	someTimestamp := metav1.NewTime(time.Now())

	pod1 := podWithDownwardAPIVolume(namespace, pod1Name, someTimestamp)
	pod2 := podWithPersistentVolume(namespace, pod2Name, someTimestamp)
	pod3 := podWithConfigMapVolume(namespace, pod3Name, someTimestamp)

	components, relations := executeVolumeCorrelation(t,
		[]coreV1.Pod{pod1, pod2, pod3},
		[]coreV1.PersistentVolumeClaim{},
		true)

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

func podWithPersistentVolume(namespace string, name string, timestamp metav1.Time) coreV1.Pod {
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
					VolumeMounts: []coreV1.VolumeMount{
						{
							Name:             "data-1",
							ReadOnly:         false,
							MountPath:        "/data",
							SubPath:          "",
							MountPropagation: nil,
							SubPathExpr:      "",
						},
					},
				},
			},
			Volumes: []coreV1.Volume{
				{
					Name: "volume1",
					VolumeSource: coreV1.VolumeSource{
						PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
							ClaimName: "data",
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}
}

func podWithDownwardAPIVolume(namespace string, name string, timestamp metav1.Time) coreV1.Pod {
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
					VolumeMounts: []coreV1.VolumeMount{
						{
							Name:        "podinfo",
							ReadOnly:    true,
							MountPath:   "/etc/podinfo",
							SubPathExpr: "",
						},
					},
				},
			},
			Volumes: []coreV1.Volume{
				{
					Name: "podinfo",
					VolumeSource: coreV1.VolumeSource{
						DownwardAPI: &coreV1.DownwardAPIVolumeSource{
							Items: []coreV1.DownwardAPIVolumeFile{
								{
									Path:     "labels",
									FieldRef: &coreV1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.labels"},
								},
								{
									Path: "annotations",
									FieldRef: &coreV1.ObjectFieldSelector{
										APIVersion: "v1",
										FieldPath:  "metadata.annotations",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func podWithConfigMapVolume(namespace string, name string, timestamp metav1.Time) coreV1.Pod {
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
					VolumeMounts: []coreV1.VolumeMount{
						{
							Name:        "podinfo",
							ReadOnly:    true,
							MountPath:   "/etc/podinfo",
							SubPathExpr: "",
						},
					},
				},
			},
			Volumes: []coreV1.Volume{
				{
					Name: "config-map",
					VolumeSource: coreV1.VolumeSource{
						ConfigMap: &coreV1.ConfigMapVolumeSource{
							LocalObjectReference: coreV1.LocalObjectReference{
								Name: "config-map",
							},
							Items: []coreV1.KeyToPath{
								{
									Key:  "key",
									Path: "/path",
								},
							},
							Optional: nil,
						},
					},
				},
			},
		},
	}
}

func executeVolumeCorrelation(
	t *testing.T,
	pods []coreV1.Pod,
	pvcs []coreV1.PersistentVolumeClaim,
	claimsEnabled bool,
) ([]*topology.Component, []*topology.Relation) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	clusterAPIClient := MockVolumeCorrelatorAPIClient{
		pods: pods, pvcs: pvcs,
	}

	podCorrChannel := make(chan *PodEndpointCorrelation)
	containerCorrChannel := make(chan *ContainerCorrelation)
	volumeCorrChannel := make(chan *VolumeCorrelation)
	volumeCorrelator := NewVolumeCorrelator(
		componentChannel,
		relationChannel,
		volumeCorrChannel,
		NewTestCommonClusterCorrelator(clusterAPIClient),
		claimsEnabled,
	)
	podCollector := NewPodCollector(
		componentChannel, relationChannel,
		containerCorrChannel, volumeCorrChannel,
		podCorrChannel,
		NewTestCommonClusterCollector(clusterAPIClient, false),
	)

	collectorsFinishChan := make(chan bool)
	correlatorFinishChan := make(chan bool)
	go func() {
		err := podCollector.CollectorFunction()
		assert.NoError(t, err)
		collectorsFinishChan <- true
	}()

	go func() {
		err := volumeCorrelator.CorrelateFunction()
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

type MockVolumeCorrelatorAPIClient struct {
	pods []coreV1.Pod
	pvcs []coreV1.PersistentVolumeClaim
	apiserver.APICollectorClient
}

func (m MockVolumeCorrelatorAPIClient) GetPods() ([]coreV1.Pod, error) {
	return m.pods, nil
}

func (m MockVolumeCorrelatorAPIClient) GetPersistentVolumeClaims() ([]coreV1.PersistentVolumeClaim, error) {
	return m.pvcs, nil
}
