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

func TestVolumeCorrelator(t *testing.T) {
	clusterName := "test-cluster-name"
	namespace := "default"
	pod1Name := "pod-1"
	pod2Name := "pod-2"
	pod3Name := "pod-3"
	pvcName := "data"
	containerName := "client-container"
	configMapName := "config-map"
	someTimestamp := metav1.NewTime(time.Now())

	pod1 := podWithDownwardAPIVolume(namespace, pod1Name, containerName, someTimestamp)
	pod2 := podWithPersistentVolume(namespace, pod2Name, pvcName, someTimestamp)
	pod3 := podWithConfigMapVolume(namespace, pod3Name, configMapName, someTimestamp)
	pvc1 := pvc(pvcName)

	components, relations := executeVolumeCorrelation(t,
		[]coreV1.Pod{pod1, pod2, pod3},
		[]coreV1.PersistentVolumeClaim{pvc1},
		true)

	expectedNamespaceID := fmt.Sprintf("urn:kubernetes:/%s:namespace/%s", clusterName, namespace)
	expectedPod1ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod1Name)
	expectedPod2ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod2Name)
	expectedPod3ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod3Name)
	expectedPVID := fmt.Sprintf("urn:kubernetes:/%s:persistent-volume/%s", clusterName, pvc1.Spec.VolumeName)
	expectedCMID := fmt.Sprintf("urn:kubernetes:/%s:%s:configmap/%s", clusterName, namespace, "config-map")
	expectedContainerID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s:container/%s", clusterName, namespace, pod1Name, containerName)
	var expectedPropagation *coreV1.MountPropagationMode

	expectedComponents := []*topology.Component{
		podComponent(namespace, pod1Name, someTimestamp),
		podComponent(namespace, pod2Name, someTimestamp),
		podComponent(namespace, pod3Name, someTimestamp),
	}
	expectedRelations := []*topology.Relation{
		simpleRelation(expectedNamespaceID, expectedPod1ID, "encloses"),
		// DownwardAPI volume mounts the pod with some data (STAC-14851)
		simpleRelationWithData(expectedContainerID, expectedPod1ID, "mounts",
			map[string]interface{}{
				"mountPath":        "/etc/podinfo",
				"mountPropagation": expectedPropagation,
				"name":             "podinfo",
				"readOnly":         true,
				"subPath":          "",
			}),
		simpleRelation(expectedNamespaceID, expectedPod2ID, "encloses"),
		simpleRelation(expectedPod2ID, expectedPVID, "claims"),
		simpleRelation(expectedNamespaceID, expectedPod3ID, "encloses"),
		simpleRelation(expectedPod3ID, expectedCMID, "claims"),
	}

	assert.EqualValues(t, expectedComponents, components)
	for _, expected := range expectedRelations {
		for _, actual := range relations {
			if expected.ExternalID == actual.ExternalID {
				assert.EqualValues(t, expected, actual)
			}
		}
	}
	return
}

func podComponent(namespace string, name string, timestamp metav1.Time) *topology.Component {
	return &topology.Component{
		ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, name),
		Type: topology.Type{
			Name: "pod",
		},
		Data: topology.Data{
			"name": name,
			"tags": map[string]string{
				"cluster-name": "test-cluster-name",
				"namespace":    namespace,
			},
			"identifiers":       []string{},
			"creationTimestamp": timestamp,
			"uid":               types.UID(""),
			"restartPolicy":     coreV1.RestartPolicy(""),
			"status": coreV1.PodStatus{
				Phase:                      "",
				Conditions:                 nil,
				Message:                    "",
				Reason:                     "",
				NominatedNodeName:          "",
				HostIP:                     "",
				PodIP:                      "",
				PodIPs:                     nil,
				StartTime:                  nil,
				InitContainerStatuses:      nil,
				ContainerStatuses:          nil,
				QOSClass:                   "",
				EphemeralContainerStatuses: nil,
			},
		},
	}
}

func podWithPersistentVolume(namespace string, name string, pvcName string, timestamp metav1.Time) coreV1.Pod {
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
							ClaimName: pvcName,
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}
}

func podWithDownwardAPIVolume(namespace string, name string, containerName string, timestamp metav1.Time) coreV1.Pod {
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
					Name: containerName,
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

func podWithConfigMapVolume(namespace string, name string, configMapName string, timestamp metav1.Time) coreV1.Pod {
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
					Name: configMapName,
					VolumeSource: coreV1.VolumeSource{
						ConfigMap: &coreV1.ConfigMapVolumeSource{
							LocalObjectReference: coreV1.LocalObjectReference{
								Name: configMapName,
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

func pvc(name string) coreV1.PersistentVolumeClaim {
	return coreV1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: coreV1.PersistentVolumeClaimSpec{
			VolumeName: fmt.Sprintf("pvc-%s", name),
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
		volumeCorrChannel,
		NewTestCommonClusterCorrelator(clusterAPIClient, componentChannel, relationChannel),
		claimsEnabled,
	)
	podCollector := NewPodCollector(
		containerCorrChannel, volumeCorrChannel,
		podCorrChannel,
		NewTestCommonClusterCollector(clusterAPIClient, componentChannel, relationChannel, false),
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
