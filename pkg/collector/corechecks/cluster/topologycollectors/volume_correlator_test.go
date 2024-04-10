//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

var (
	volumeSource = coreV1.VolumeSource{
		EmptyDir: &coreV1.EmptyDirVolumeSource{
			Medium:    coreV1.StorageMediumMemory,
			SizeLimit: resource.NewQuantity(10, resource.DecimalSI),
		},
	}
)

func TestVolumeCorrelator(t *testing.T) {
	clusterName := "test-cluster-name"
	namespace := "default"
	pod1Name := "pod-1"
	pod2Name := "pod-2"
	pod3Name := "pod-3"
	pod4Name := "pod-4"
	pvcName := "data"
	containerName := "client-container"
	configMapName := "config-map"
	secretName := "secret"
	volumeName := "volume"
	someTimestamp := metav1.NewTime(time.Now())
	someTimestampFormatted := someTimestamp.UTC().Format(time.RFC3339)

	pod1 := podWithDownwardAPIVolume(namespace, pod1Name, containerName, someTimestamp)
	pod2 := podWithPersistentVolume(namespace, pod2Name, pvcName, someTimestamp)
	pod3 := podWithConfigMapAndSecretVolume(namespace, pod3Name, configMapName, secretName, someTimestamp)
	pod4 := podWithEmptyDirVolume(namespace, pod4Name, volumeName, someTimestamp)
	pvc1 := pvc(pvcName)

	expectedNamespaceID := fmt.Sprintf("urn:kubernetes:/%s:namespace/%s", clusterName, namespace)
	expectedPod1ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod1Name)
	expectedPod2ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod2Name)
	expectedPod3ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod3Name)
	expectedPod4ID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s", clusterName, namespace, pod4Name)
	expectedPVCID := fmt.Sprintf("urn:kubernetes:/%s:%s:persistent-volume-claim/%s", clusterName, namespace, pvcName)
	expectedConfigMapID := fmt.Sprintf("urn:kubernetes:/%s:%s:configmap/%s", clusterName, namespace, "config-map")
	expectedVID := fmt.Sprintf("urn:kubernetes:/%s:empty-dir:volume/%s/%s/%s", clusterName, namespace, pod4Name, "volume")
	expectedContainerID := fmt.Sprintf("urn:kubernetes:/%s:%s:pod/%s:container/%s", clusterName, namespace, pod1Name, containerName)
	expectedSecretID := fmt.Sprintf("urn:kubernetes:/%s:%s:secret/%s", clusterName, namespace, secretName)

	var expectedPropagation *coreV1.MountPropagationMode

	for _, sourcePropertiesEnabled := range []bool{false, true} {
		for _, kubernetesStatusEnabled := range []bool{false, true} {

			expectedComponents := []*topology.Component{
				chooseBySourcePropertiesFeature(
					sourcePropertiesEnabled,
					kubernetesStatusEnabled,
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod1Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod1Name,
							"kind": "Pod",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers":       []string{},
							"creationTimestamp": someTimestamp,
							"uid":               types.UID(""),
							"restartPolicy":     coreV1.RestartPolicyAlways,
							"status": coreV1.PodStatus{
								Phase:     coreV1.PodRunning,
								StartTime: &someTimestamp,
								PodIP:     "10.0.0.1",
							},
						},
					},
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod1Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod1Name,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers": []string{},
							"status":      map[string]interface{}{"phase": "Running"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"creationTimestamp": someTimestampFormatted,
								"deletionTimestamp": someTimestampFormatted,
								"name":              pod1Name,
								"namespace":         namespace,
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":      "client-container",
										"resources": map[string]interface{}{},
										"volumeMounts": []interface{}{
											map[string]interface{}{"mountPath": "/etc/podinfo", "name": "podinfo", "readOnly": true},
										},
									},
								},
								"hostNetwork":   true,
								"restartPolicy": "Always",
								"volumes": []interface{}{
									map[string]interface{}{
										"name": "podinfo",
										"downwardAPI": map[string]interface{}{
											"items": []interface{}{
												map[string]interface{}{
													"fieldRef": map[string]interface{}{
														"apiVersion": "v1",
														"fieldPath":  "metadata.labels",
													},
													"path": "labels",
												},
												map[string]interface{}{
													"fieldRef": map[string]interface{}{
														"apiVersion": "v1",
														"fieldPath":  "metadata.annotations",
													},
													"path": "annotations",
												},
											},
										},
									},
								},
							},
							"status": map[string]interface{}{
								"phase":     "Running",
								"podIP":     "10.0.0.1",
								"startTime": someTimestampFormatted,
							},
						},
					},
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod1Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod1Name,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers": []string{},
							"status":      map[string]interface{}{"phase": "Running"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"creationTimestamp": someTimestampFormatted,
								"deletionTimestamp": someTimestampFormatted,
								"name":              pod1Name,
								"namespace":         namespace,
							},
							"spec": map[string]interface{}{
								"containers": []interface{}{
									map[string]interface{}{
										"name":      "client-container",
										"resources": map[string]interface{}{},
										"volumeMounts": []interface{}{
											map[string]interface{}{"mountPath": "/etc/podinfo", "name": "podinfo", "readOnly": true},
										},
									},
								},
								"hostNetwork":   true,
								"restartPolicy": "Always",
								"volumes": []interface{}{
									map[string]interface{}{
										"name": "podinfo",
										"downwardAPI": map[string]interface{}{
											"items": []interface{}{
												map[string]interface{}{
													"fieldRef": map[string]interface{}{
														"apiVersion": "v1",
														"fieldPath":  "metadata.labels",
													},
													"path": "labels",
												},
												map[string]interface{}{
													"fieldRef": map[string]interface{}{
														"apiVersion": "v1",
														"fieldPath":  "metadata.annotations",
													},
													"path": "annotations",
												},
											},
										},
									},
								},
							},
							"status": map[string]interface{}{
								"phase":     "Running",
								"podIP":     "10.0.0.1",
								"startTime": someTimestampFormatted,
							},
						},
					},
				),
				chooseBySourcePropertiesFeature(
					sourcePropertiesEnabled,
					kubernetesStatusEnabled,
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod2Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod2Name,
							"kind": "Pod",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers":       []string{},
							"creationTimestamp": someTimestamp,
							"uid":               types.UID(""),
							"restartPolicy":     coreV1.RestartPolicyAlways,
							"status": coreV1.PodStatus{
								Phase:     coreV1.PodRunning,
								StartTime: &someTimestamp,
								PodIP:     "10.0.0.1",
							},
						},
					},
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod2Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod2Name,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers": []string{},
							"status":      map[string]interface{}{"phase": "Running"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"creationTimestamp": someTimestampFormatted,
								"deletionTimestamp": someTimestampFormatted,
								"name":              pod2Name,
								"namespace":         namespace,
							},
							"spec": map[string]interface{}{
								"hostNetwork":   true,
								"restartPolicy": "Always",
								"containers": []interface{}{
									map[string]interface{}{
										"name":      "",
										"resources": map[string]interface{}{},
										"volumeMounts": []interface{}{
											map[string]interface{}{"mountPath": "/data", "name": "data-1"},
										},
									},
								},
								"volumes": []interface{}{
									map[string]interface{}{
										"name": "volume1",
										"persistentVolumeClaim": map[string]interface{}{
											"claimName": "data",
										},
									},
								},
							},
							"status": map[string]interface{}{
								"phase":     "Running",
								"podIP":     "10.0.0.1",
								"startTime": someTimestampFormatted,
							},
						},
					},
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod2Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod2Name,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers": []string{},
							"status":      map[string]interface{}{"phase": "Running"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"creationTimestamp": someTimestampFormatted,
								"deletionTimestamp": someTimestampFormatted,
								"name":              pod2Name,
								"namespace":         namespace,
							},
							"spec": map[string]interface{}{
								"hostNetwork":   true,
								"restartPolicy": "Always",
								"containers": []interface{}{
									map[string]interface{}{
										"name":      "",
										"resources": map[string]interface{}{},
										"volumeMounts": []interface{}{
											map[string]interface{}{"mountPath": "/data", "name": "data-1"},
										},
									},
								},
								"volumes": []interface{}{
									map[string]interface{}{
										"name": "volume1",
										"persistentVolumeClaim": map[string]interface{}{
											"claimName": "data",
										},
									},
								},
							},
							"status": map[string]interface{}{
								"phase":     "Running",
								"podIP":     "10.0.0.1",
								"startTime": someTimestampFormatted,
							},
						},
					},
				),
				chooseBySourcePropertiesFeature(
					sourcePropertiesEnabled,
					kubernetesStatusEnabled,
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod3Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod3Name,
							"kind": "Pod",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers":       []string{},
							"creationTimestamp": someTimestamp,
							"uid":               types.UID(""),
							"restartPolicy":     coreV1.RestartPolicyAlways,
							"status": coreV1.PodStatus{
								Phase:     coreV1.PodRunning,
								StartTime: &someTimestamp,
								PodIP:     "10.0.0.1",
							},
						},
					},
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod3Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod3Name,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers": []string{},
							"status":      map[string]interface{}{"phase": "Running"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"creationTimestamp": someTimestampFormatted,
								"deletionTimestamp": someTimestampFormatted,
								"name":              pod3Name,
								"namespace":         namespace,
							},
							"spec": map[string]interface{}{
								"hostNetwork":   true,
								"restartPolicy": "Always",
								"containers": []interface{}{
									map[string]interface{}{
										"name":      "",
										"resources": map[string]interface{}{},
										"volumeMounts": []interface{}{
											map[string]interface{}{"mountPath": "/etc/podinfo", "name": "podinfo", "readOnly": true},
										},
									},
								},
								"volumes": []interface{}{
									map[string]interface{}{
										"name": "config-map",
										"configMap": map[string]interface{}{
											"items": []interface{}{
												map[string]interface{}{"key": "key", "path": "/path"},
											},
											"name": "config-map",
										},
									},
									map[string]interface{}{
										"name": "secret",
										"secret": map[string]interface{}{
											"items": []interface{}{
												map[string]interface{}{"key": "key", "path": "/path"},
											},
											"secretName": "secret",
										},
									},
								},
							},
							"status": map[string]interface{}{
								"phase":     "Running",
								"podIP":     "10.0.0.1",
								"startTime": someTimestampFormatted,
							},
						},
					},
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod3Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod3Name,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers": []string{},
							"status":      map[string]interface{}{"phase": "Running"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"creationTimestamp": someTimestampFormatted,
								"deletionTimestamp": someTimestampFormatted,
								"name":              pod3Name,
								"namespace":         namespace,
							},
							"spec": map[string]interface{}{
								"hostNetwork":   true,
								"restartPolicy": "Always",
								"containers": []interface{}{
									map[string]interface{}{
										"name":      "",
										"resources": map[string]interface{}{},
										"volumeMounts": []interface{}{
											map[string]interface{}{"mountPath": "/etc/podinfo", "name": "podinfo", "readOnly": true},
										},
									},
								},
								"volumes": []interface{}{
									map[string]interface{}{
										"name": "config-map",
										"configMap": map[string]interface{}{
											"items": []interface{}{
												map[string]interface{}{"key": "key", "path": "/path"},
											},
											"name": "config-map",
										},
									},
									map[string]interface{}{
										"name": "secret",
										"secret": map[string]interface{}{
											"items": []interface{}{
												map[string]interface{}{"key": "key", "path": "/path"},
											},
											"secretName": "secret",
										},
									},
								},
							},
							"status": map[string]interface{}{
								"phase":     "Running",
								"podIP":     "10.0.0.1",
								"startTime": someTimestampFormatted,
							},
						},
					},
				),
				chooseBySourcePropertiesFeature(
					sourcePropertiesEnabled,
					kubernetesStatusEnabled,
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod4Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod4Name,
							"kind": "Pod",
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers":       []string{},
							"creationTimestamp": someTimestamp,
							"uid":               types.UID(""),
							"restartPolicy":     coreV1.RestartPolicyAlways,
							"status": coreV1.PodStatus{
								Phase:     coreV1.PodRunning,
								StartTime: &someTimestamp,
								PodIP:     "10.0.0.1",
							},
						},
					},
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod4Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod4Name,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers": []string{},
							"status":      map[string]interface{}{"phase": "Running"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"creationTimestamp": someTimestampFormatted,
								"deletionTimestamp": someTimestampFormatted,
								"name":              pod4Name,
								"namespace":         namespace,
							},
							"spec": map[string]interface{}{
								"hostNetwork":   true,
								"restartPolicy": "Always",
								"containers": []interface{}{
									map[string]interface{}{
										"name":      "",
										"resources": map[string]interface{}{},
										"volumeMounts": []interface{}{
											map[string]interface{}{"mountPath": "/etc/podinfo", "name": "podinfo", "readOnly": true},
										},
									},
								},
								"volumes": []interface{}{
									map[string]interface{}{
										"name": "volume",
										"emptyDir": map[string]interface{}{
											"medium":    "Memory",
											"sizeLimit": "10",
										},
									},
								},
							},
							"status": map[string]interface{}{
								"phase":     "Running",
								"podIP":     "10.0.0.1",
								"startTime": someTimestampFormatted,
							},
						},
					},
					&topology.Component{
						ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:pod/%s", namespace, pod4Name),
						Type: topology.Type{
							Name: "pod",
						},
						Data: topology.Data{
							"name": pod4Name,
							"tags": map[string]string{
								"cluster-name":   "test-cluster-name",
								"cluster-type":   "kubernetes",
								"component-type": "kubernetes-pod",
								"namespace":      namespace,
							},
							"identifiers": []string{},
							"status":      map[string]interface{}{"phase": "Running"},
						},
						SourceProperties: map[string]interface{}{
							"apiVersion": "v1",
							"kind":       "Pod",
							"metadata": map[string]interface{}{
								"creationTimestamp": someTimestampFormatted,
								"deletionTimestamp": someTimestampFormatted,
								"name":              pod4Name,
								"namespace":         namespace,
							},
							"spec": map[string]interface{}{
								"hostNetwork":   true,
								"restartPolicy": "Always",
								"containers": []interface{}{
									map[string]interface{}{
										"name":      "",
										"resources": map[string]interface{}{},
										"volumeMounts": []interface{}{
											map[string]interface{}{"mountPath": "/etc/podinfo", "name": "podinfo", "readOnly": true},
										},
									},
								},
								"volumes": []interface{}{
									map[string]interface{}{
										"name": "volume",
										"emptyDir": map[string]interface{}{
											"medium":    "Memory",
											"sizeLimit": "10",
										},
									},
								},
							},
							"status": map[string]interface{}{
								"phase":     "Running",
								"podIP":     "10.0.0.1",
								"startTime": someTimestampFormatted,
							},
						},
					},
				),
				volumeComponent(namespace, pod4Name, volumeName, "empty-dir", someTimestampFormatted, nil,
					map[string]string{"kind": "empty-dir"},
					volumeSource,
					sourcePropertiesEnabled, kubernetesStatusEnabled),
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
				simpleRelation(expectedPod2ID, expectedPVCID, "claims"),
				simpleRelation(expectedNamespaceID, expectedPod3ID, "encloses"),
				simpleRelation(expectedPod3ID, expectedConfigMapID, "claims"),
				simpleRelation(expectedPod3ID, expectedSecretID, "claims"),
				simpleRelation(expectedPod4ID, expectedVID, "claims"),
				simpleRelation(expectedNamespaceID, expectedPod4ID, "encloses"),
			}

			t.Run(testCaseName("Test volume correlator", sourcePropertiesEnabled, kubernetesStatusEnabled), func(t *testing.T) {

				components, relations := executeVolumeCorrelation(t,
					[]coreV1.Pod{pod1, pod2, pod3, pod4},
					[]coreV1.PersistentVolumeClaim{pvc1},
					true, sourcePropertiesEnabled, kubernetesStatusEnabled)

				assert.ElementsMatch(t, expectedComponents, components)
				assert.ElementsMatch(t, expectedRelations, relations)
			})
		}
	}

	return
}

func podWithPersistentVolume(namespace, name, pvcName string, timestamp metav1.Time) coreV1.Pod {
	return coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
		Status: coreV1.PodStatus{
			Phase:     coreV1.PodRunning,
			StartTime: &timestamp,
			PodIP:     "10.0.0.1",
		},
		Spec: coreV1.PodSpec{
			HostNetwork:   true,
			RestartPolicy: coreV1.RestartPolicyAlways,
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

func podWithDownwardAPIVolume(namespace, name, containerName string, timestamp metav1.Time) coreV1.Pod {
	return coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
		Status: coreV1.PodStatus{
			Phase:     coreV1.PodRunning,
			StartTime: &timestamp,
			PodIP:     "10.0.0.1",
		},
		Spec: coreV1.PodSpec{
			HostNetwork:   true,
			RestartPolicy: coreV1.RestartPolicyAlways,
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

func podWithConfigMapAndSecretVolume(namespace, name, configMapName string, secretName string, timestamp metav1.Time) coreV1.Pod {
	return coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
		Status: coreV1.PodStatus{
			Phase:     coreV1.PodRunning,
			StartTime: &timestamp,
			PodIP:     "10.0.0.1",
		},
		Spec: coreV1.PodSpec{
			HostNetwork:   true,
			RestartPolicy: coreV1.RestartPolicyAlways,
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
				{
					Name: secretName,
					VolumeSource: coreV1.VolumeSource{
						Secret: &coreV1.SecretVolumeSource{
							SecretName: secretName,
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

func podWithEmptyDirVolume(namespace, name, dirName string, timestamp metav1.Time) coreV1.Pod {
	return coreV1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
		TypeMeta: metav1.TypeMeta{Kind: "Pod"},
		Status: coreV1.PodStatus{
			Phase:     coreV1.PodRunning,
			StartTime: &timestamp,
			PodIP:     "10.0.0.1",
		},
		Spec: coreV1.PodSpec{
			HostNetwork:   true,
			RestartPolicy: coreV1.RestartPolicyAlways,
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
					Name:         dirName,
					VolumeSource: volumeSource,
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

func volumeComponent(namespace, podName, volumeName, volumeType, someTimestampFormatted string, identifiers []string, extraTags map[string]string,
	volumeSource coreV1.VolumeSource, sourcePropertiesEnabled, kubernetesStatusEnabled bool) *topology.Component {
	tags := map[string]string{
		"cluster-name":   "test-cluster-name",
		"cluster-type":   "kubernetes",
		"component-type": "kubernetes-volume",
		"namespace":      namespace,
	}
	for k, v := range extraTags {
		tags[k] = v
	}

	data := topology.Data{
		"name": volumeName,
		"tags": tags,
	}
	if identifiers != nil {
		data["identifiers"] = identifiers
	}

	if sourcePropertiesEnabled {
		if kubernetesStatusEnabled {
			return &topology.Component{
				ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:volume/%s/%s/%s", volumeType, namespace, podName, volumeName),
				Type:       topology.Type{Name: "volume"},
				Data:       data,
				SourceProperties: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Volume",
					"metadata": map[string]interface{}{
						"name":              "volume",
						"creationTimestamp": someTimestampFormatted,
						"namespace":         "default",
					},
					"volume": map[string]interface{}{
						"name": "volume",
						"emptyDir": map[string]interface{}{
							"medium":    "Memory",
							"sizeLimit": "10",
						},
					},
				},
			}
		}
		return &topology.Component{
			ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:volume/%s/%s/%s", volumeType, namespace, podName, volumeName),
			Type:       topology.Type{Name: "volume"},
			Data:       data,
			SourceProperties: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Volume",
				"metadata": map[string]interface{}{
					"name":              "volume",
					"creationTimestamp": someTimestampFormatted,
					"namespace":         "default",
				},
				"volume": map[string]interface{}{
					"name": "volume",
					"emptyDir": map[string]interface{}{
						"medium":    "Memory",
						"sizeLimit": "10",
					},
				},
			},
		}
	}
	data["source"] = volumeSource
	return &topology.Component{
		ExternalID: fmt.Sprintf("urn:kubernetes:/test-cluster-name:%s:volume/%s/%s/%s", volumeType, namespace, podName, volumeName),
		Type:       topology.Type{Name: "volume"},
		Data:       data,
	}
}

func executeVolumeCorrelation(
	t *testing.T,
	pods []coreV1.Pod,
	pvcs []coreV1.PersistentVolumeClaim,
	claimsEnabled bool,
	sourcePropertiesEnabled, kubernetesStatusEnabled bool,
) ([]*topology.Component, []*topology.Relation) {

	componentChannel := make(chan *topology.Component)
	defer close(componentChannel)
	relationChannel := make(chan *topology.Relation)
	defer close(relationChannel)

	clusterAPIClient := MockVolumeCorrelatorAPIClient{
		pods: pods, pvcs: pvcs,
	}

	podCorrChannel := make(chan *PodLabelCorrelation)
	// Pod correlation is just a no-op sink to assure progress
	go func() {
		for range podCorrChannel {
		}
	}()

	containerCorrChannel := make(chan *ContainerCorrelation)
	volumeCorrChannel := make(chan *VolumeCorrelation)
	commonClusterCollector := NewTestCommonClusterCollector(clusterAPIClient, componentChannel, relationChannel, sourcePropertiesEnabled, kubernetesStatusEnabled)
	commonClusterCollector.SetUseRelationCache(false)
	volumeCorrelator := NewVolumeCorrelator(
		volumeCorrChannel,
		NewClusterTopologyCorrelator(commonClusterCollector),
		claimsEnabled,
	)
	podCollector := NewPodCollector(
		containerCorrChannel, volumeCorrChannel,
		podCorrChannel,
		commonClusterCollector,
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
